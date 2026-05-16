package semilla

import (
	"context"
	"log/slog"
	"os"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

type definicionSistema struct {
	Codigo              string
	Nombre              string
	Descripcion         string
	UrlAcceso           string
	UrlLogin            string
	NombreCampoUsuario  string
	NombreCampoPassword string
	MetodoLogin         string
	Motor               string
	ClaveAdaptador      string
	SoportaLectura      bool
	SoportaAutoregistro bool
	HostLectura         string
	PuertoLectura       int64
	BaseDatosLectura    string
}

var sistemasIniciales = []definicionSistema{
	{
		Codigo:              "crm_codeplex",
		Nombre:              "CRM Codeplex",
		Descripcion:         "CRM con integración WhatsApp",
		UrlAcceso:           "https://codeplex.pe/crm/admin/authentication",
		UrlLogin:            "https://codeplex.pe/crm/admin/authentication",
		NombreCampoUsuario:  "correo_electronico",
		NombreCampoPassword: "password",
		MetodoLogin:         "POST",
		Motor:               "COCKROACHDB",
		ClaveAdaptador:      "codeplex_ventas",
		SoportaLectura:      true,
		SoportaAutoregistro: true,
		HostLectura:         "144.217.163.120",
		PuertoLectura:       26257,
		BaseDatosLectura:    "codeplex_ventas",
	},
	{
		Codigo:              "dashpanel_control",
		Nombre:              "Dashpanel de Control",
		Descripcion:         "Panel de control de facturación",
		UrlAcceso:           "https://dashpanelcontrolv2.codeplex.pe/acceder",
		UrlLogin:            "https://dashpanelcontrolv2.codeplex.pe/acceder",
		NombreCampoUsuario:  "correo_electronico",
		NombreCampoPassword: "password",
		MetodoLogin:         "POST",
		Motor:               "COCKROACHDB",
		ClaveAdaptador:      "panel_control_facturacion_2026",
		SoportaLectura:      true,
		SoportaAutoregistro: false,
		HostLectura:         "144.217.163.120",
		PuertoLectura:       26257,
		BaseDatosLectura:    "panel_control_facturacion_2026",
	},
}

func SembrarSistemasIniciales(contexto context.Context, conexion *cockroach.ConexionBaseDatos, claves *cripto.ClavesCifrado) error {
	usuarioDb := os.Getenv("DB_EXTERNA_USUARIO")
	if usuarioDb == "" {
		usuarioDb = "dev_user"
	}
	passwordDb := os.Getenv("DB_EXTERNA_PASSWORD")
	if passwordDb == "" {
		slog.Info("DB_EXTERNA_PASSWORD no definida; seed de sistemas se salta para no asumir credenciales de producción")
		return nil
	}

	clave := claves.ClaveBoveda()

	for _, def := range sistemasIniciales {
		existente, err := consultarPorCodigoSiExiste(contexto, conexion, def.Codigo)
		if err != nil {
			return err
		}
		if existente != nil {
			continue
		}

		nuevo := &catalogo_sistemas.SistemaDestino{
			Codigo:              def.Codigo,
			Nombre:              def.Nombre,
			Descripcion:         def.Descripcion,
			UrlAcceso:           def.UrlAcceso,
			UrlLogin:            def.UrlLogin,
			NombreCampoUsuario:  def.NombreCampoUsuario,
			NombreCampoPassword: def.NombreCampoPassword,
			MetodoLogin:         def.MetodoLogin,
			Motor:               def.Motor,
			ClaveAdaptador:      def.ClaveAdaptador,
			SoportaLectura:      def.SoportaLectura,
			SoportaAutoregistro: def.SoportaAutoregistro,
			Estado:              "ACTIVO",
		}
		if err := catalogo_sistemas.InsertarSistemaDestino(contexto, conexion.Pool(), nuevo); err != nil {
			return err
		}

		usuarioCifrado, err := cripto.CifrarConAesGcm(clave, []byte(usuarioDb))
		if err != nil {
			return err
		}
		passwordCifrada, err := cripto.CifrarConAesGcm(clave, []byte(passwordDb))
		if err != nil {
			return err
		}

		conexionLectura := &catalogo_sistemas.ConexionLectura{
			SistemaDestinoId:  nuevo.Id,
			Host:              def.HostLectura,
			Puerto:            def.PuertoLectura,
			BaseDatos:         def.BaseDatosLectura,
			UsuarioDbCifrado:  usuarioCifrado,
			PasswordDbCifrada: passwordCifrada,
			SslModo:           "require",
			Estado:            "ACTIVO",
		}
		if err := catalogo_sistemas.InsertarConexionLectura(contexto, conexion.Pool(), conexionLectura); err != nil {
			return err
		}

		slog.Info("sistema_sembrado", "codigo", def.Codigo, "adaptador", def.ClaveAdaptador)
	}

	return nil
}

func consultarPorCodigoSiExiste(contexto context.Context, conexion *cockroach.ConexionBaseDatos, codigo string) (*string, error) {
	var id string
	err := conexion.Pool().QueryRow(contexto, `
		SELECT id::TEXT FROM sistema_destino WHERE lower(codigo) = lower($1) AND estado != 'ELIMINADO'
	`, codigo).Scan(&id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, nil
	}
	return &id, nil
}
