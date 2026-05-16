package http

import (
	"errors"
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoRegistrarOperador struct {
	CorreoElectronico string `json:"correo_electronico"`
	PasswordPlana     string `json:"password_plana"`
	ActivarInmediato  bool   `json:"activar_inmediato"`
}

func ConstruirHandlerRegistrarOperador(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		var cuerpo cuerpoRegistrarOperador
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		datos := identidad.DatosRegistroUsuario{
			CorreoElectronico: cuerpo.CorreoElectronico,
			PasswordPlana:     cuerpo.PasswordPlana,
			RegistradoPor:     &sesion.UsuarioId,
			IpOrigen:          obtenerIpRemota(peticion),
			AgenteUsuario:     peticion.UserAgent(),
		}

		var resultado *identidad.ResultadoRegistroUsuario
		var err error
		if cuerpo.ActivarInmediato {
			resultado, err = identidad.RegistrarUsuarioYActivar(peticion.Context(), conexion, datos)
		} else {
			resultado, err = identidad.RegistrarUsuario(peticion.Context(), conexion, datos)
		}

		if errors.Is(err, identidad.ErrCorreoYaRegistrado) {
			ResponderError(escritor, http.StatusConflict, codigosError.CodigoCorreoYaRegistrado, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusCreated, map[string]any{
			"operador_id": resultado.UsuarioId.String(),
			"estado":      resultado.Estado,
		})
	}
}
