package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

// ConstruirHandlerDesactivarAcceso marca el acceso como REVOCADO (reversible).
func ConstruirHandlerDesactivarAcceso(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id inválido")
			return
		}
		if err := boveda.DesactivarAcceso(peticion.Context(), conexion.Pool(), id, sesion.UsuarioId); err != nil {
			if errors.Is(err, boveda.ErrAccesoNoEncontrado) {
				ResponderError(escritor, http.StatusNotFound, "ACCESO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"acceso_desactivado": true})
	}
}

// ConstruirHandlerReactivarAcceso vuelve a poner el acceso en ACTIVO.
func ConstruirHandlerReactivarAcceso(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id inválido")
			return
		}
		if err := boveda.ReactivarAcceso(peticion.Context(), conexion.Pool(), id, sesion.UsuarioId); err != nil {
			if errors.Is(err, boveda.ErrAccesoNoEncontrado) {
				ResponderError(escritor, http.StatusNotFound, "ACCESO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"acceso_reactivado": true})
	}
}
