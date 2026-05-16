package http

import (
	"errors"
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoCambioPassword struct {
	PasswordActual string `json:"password_actual"`
	PasswordNueva  string `json:"password_nueva"`
}

func ConstruirHandlerCambiarPassword(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		var cuerpo cuerpoCambioPassword
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		err := identidad.CambiarPassword(peticion.Context(), conexion, identidad.DatosCambioPassword{
			UsuarioId:      sesion.UsuarioId,
			PasswordActual: cuerpo.PasswordActual,
			PasswordNueva:  cuerpo.PasswordNueva,
			SesionId:       &sesion.SesionId,
			IpOrigen:       obtenerIpRemota(peticion),
			AgenteUsuario:  peticion.UserAgent(),
		})

		if errors.Is(err, identidad.ErrPasswordActualInvalido) {
			ResponderError(escritor, http.StatusUnauthorized, codigosError.CodigoCredencialesInvalidas, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"password_cambiado": true})
	}
}
