package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

// ConstruirHandlerEliminarPermanente borra físicamente el acceso de la BD.
// Acción destructiva e irreversible — el frontend debe protegerlo con un
// modal de confirmación type-to-confirm. Acá del lado servidor SOLO valida
// sesión + 2FA + UUID; la confirmación es responsabilidad del cliente.
//
// El audit_accion se escribe ANTES del DELETE (en la misma transacción), así
// queda registro inmutable de quién/cuándo/de-dónde borró aunque la fila
// destino ya no exista.
func ConstruirHandlerEliminarPermanente(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
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

		trans, err := conexion.Pool().Begin(peticion.Context())
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		defer trans.Rollback(peticion.Context())

		// Audit primero — si después falla el DELETE, el rollback descarta el
		// log junto con el cambio. Así nunca queda "X borró Y" sin Y borrado,
		// ni "Y borrado" sin saber quién.
		if err := auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
			UsuarioId:     &sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			Modulo:        "BOVEDA",
			Accion:        "ACCESO_ELIMINADO_PERMANENTE",
			Entidad:       "acceso_guardado",
			EntidadId:     &id,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		}); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		if err := boveda.EliminarAccesoPermanente(peticion.Context(), trans, id, sesion.UsuarioId); err != nil {
			if errors.Is(err, boveda.ErrAccesoNoEncontrado) {
				ResponderError(escritor, http.StatusNotFound, "ACCESO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		if err := trans.Commit(peticion.Context()); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"eliminado_permanente": true})
	}
}
