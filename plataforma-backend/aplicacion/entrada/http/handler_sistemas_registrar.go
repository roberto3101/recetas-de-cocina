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
	UrlAcceso           string `json:"url_acceso"`
	UrlLogin            string `json:"url_login"`
	NombreCampoUsuario  string `json:"nombre_campo_usuario"`
	NombreCampoPassword string `json:"nombre_campo_password"`
	MetodoLogin         string `json:"metodo_login"`
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
			Codigo:              cuerpo.Codigo,
			Nombre:              cuerpo.Nombre,
			UrlAcceso:           cuerpo.UrlAcceso,
			UrlLogin:            cuerpo.UrlLogin,
			NombreCampoUsuario:  cuerpo.NombreCampoUsuario,
			NombreCampoPassword: cuerpo.NombreCampoPassword,
			MetodoLogin:         cuerpo.MetodoLogin,
			CreadoPor:           sesion.UsuarioId,
			SesionId:            &sesion.SesionId,
			IpOrigen:            obtenerIpRemota(peticion),
			AgenteUsuario:       peticion.UserAgent(),
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
