package http

import (
	"errors"
	"net/http"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoRegistrarSistema struct {
	Codigo              string `json:"codigo"`
	Nombre              string `json:"nombre"`
	Descripcion         string `json:"descripcion"`
	UrlAcceso           string `json:"url_acceso"`
	Motor               string `json:"motor"`
	ClaveAdaptador      string `json:"clave_adaptador"`
	RequiereLoginGlobal bool   `json:"requiere_login_global"`
	SoportaLectura      bool   `json:"soporta_lectura"`
	SoportaAutoregistro bool   `json:"soporta_autoregistro"`

	HostLectura         string `json:"host_lectura,omitempty"`
	PuertoLectura       int64  `json:"puerto_lectura,omitempty"`
	BaseDatosLectura    string `json:"base_datos_lectura,omitempty"`
	UsuarioLectura      string `json:"usuario_lectura,omitempty"`
	PasswordLectura     string `json:"password_lectura,omitempty"`
	SslModoLectura      string `json:"ssl_modo_lectura,omitempty"`

	AlgoritmoHashDestino string `json:"algoritmo_hash_destino,omitempty"`
	CostoHashDestino     *int64 `json:"costo_hash_destino,omitempty"`
	SalEstrategia        string `json:"sal_estrategia,omitempty"`
}

func ConstruirHandlerRegistrarSistema(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		var cuerpo cuerpoRegistrarSistema
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		resultado, err := catalogo_sistemas.RegistrarSistema(peticion.Context(), conexion, clavesCifrado, catalogo_sistemas.DatosRegistrarSistema{
			Codigo:               cuerpo.Codigo,
			Nombre:               cuerpo.Nombre,
			Descripcion:          cuerpo.Descripcion,
			UrlAcceso:            cuerpo.UrlAcceso,
			Motor:                cuerpo.Motor,
			ClaveAdaptador:       cuerpo.ClaveAdaptador,
			RequiereLoginGlobal:  cuerpo.RequiereLoginGlobal,
			SoportaLectura:       cuerpo.SoportaLectura,
			SoportaAutoregistro:  cuerpo.SoportaAutoregistro,
			HostLectura:          cuerpo.HostLectura,
			PuertoLectura:        cuerpo.PuertoLectura,
			BaseDatosLectura:     cuerpo.BaseDatosLectura,
			UsuarioLecturaPlano:  cuerpo.UsuarioLectura,
			PasswordLecturaPlana: cuerpo.PasswordLectura,
			SslModoLectura:       cuerpo.SslModoLectura,
			AlgoritmoHashDestino: cuerpo.AlgoritmoHashDestino,
			CostoHashDestino:     cuerpo.CostoHashDestino,
			SalEstrategia:        cuerpo.SalEstrategia,
			CreadoPor:            sesion.UsuarioId,
			SesionId:             &sesion.SesionId,
			IpOrigen:             obtenerIpRemota(peticion),
			AgenteUsuario:        peticion.UserAgent(),
		})
		if errors.Is(err, catalogo_sistemas.ErrCodigoSistemaDuplicado) {
			ResponderError(escritor, http.StatusConflict, codigosError.CodigoCodigoSistemaDuplicado, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusCreated, map[string]any{"sistema_id": resultado.SistemaId.String()})
	}
}
