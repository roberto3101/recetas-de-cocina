package http

import (
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
	"sistemas-unificados/plataforma/sesion"
)

func ConstruirHandlerCerrarSesion(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		s, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		if err := identidad.CerrarSesion(peticion.Context(), conexion, identidad.DatosCerrarSesion{
			SesionId:      s.SesionId,
			UsuarioId:     s.UsuarioId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
			Motivo:        "logout solicitado",
		}); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		sesion.LimpiarCookieSesion(escritor)
		ResponderExito(escritor, http.StatusOK, map[string]any{"sesion_cerrada": true})
	}
}
