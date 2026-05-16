package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerRevocarSesion(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		idBruto := chi.URLParam(peticion, "id")
		sesionId, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id de sesión inválido")
			return
		}

		if err := identidad.CerrarSesion(peticion.Context(), conexion, identidad.DatosCerrarSesion{
			SesionId:      sesionId,
			UsuarioId:     sesion.UsuarioId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
			Motivo:        "revocada por el operador",
		}); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"sesion_revocada": true})
	}
}
