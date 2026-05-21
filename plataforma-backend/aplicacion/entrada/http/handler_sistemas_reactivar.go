package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

// ConstruirHandlerReactivarSistema saca al sistema del archivo (estado
// ELIMINADO → ACTIVO) y revive sus accesos en cascada (REVOCADO → ACTIVO).
// Devuelve cuántos accesos volvieron a ACTIVO.
func ConstruirHandlerReactivarSistema(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id inválido")
			return
		}
		resultado, err := catalogo_sistemas.ReactivarSistema(peticion.Context(), conexion, catalogo_sistemas.DatosReactivarSistema{
			SistemaId:     id,
			ReactivadoPor: sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		})
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{
			"sistema_reactivado":  true,
			"accesos_reactivados": resultado.AccesosReactivados,
		})
	}
}

// ConstruirHandlerContarAccesosRevocadosDeSistema devuelve cuántos accesos
// quedaron REVOCADOS para el sistema. El frontend lo usa para mostrar en
// el modal de reactivación cuántos accesos van a revivir.
func ConstruirHandlerContarAccesosRevocadosDeSistema(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id inválido")
			return
		}
		total, err := catalogo_sistemas.ContarAccesosRevocadosDeSistema(peticion.Context(), conexion.Pool(), id)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"accesos_revocados": total})
	}
}
