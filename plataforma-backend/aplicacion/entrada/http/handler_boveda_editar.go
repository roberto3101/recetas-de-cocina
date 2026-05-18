package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoEditarAcceso struct {
	Titulo           string  `json:"titulo"`
	SistemaDestinoId string  `json:"sistema_destino_id"`
	UsuarioExterno   string  `json:"usuario_externo"`
	PasswordPlana    *string `json:"password,omitempty"`
	Observaciones    string  `json:"observaciones"`
	Tipo             string  `json:"tipo"`
	Puerto           *int16  `json:"puerto,omitempty"`
}

func ConstruirHandlerEditarAcceso(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
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
		var cuerpo cuerpoEditarAcceso
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}
		sistemaId, err := uuid.Parse(cuerpo.SistemaDestinoId)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "sistema_destino_id inválido")
			return
		}

		var passwordPlana string
		if cuerpo.PasswordPlana != nil {
			passwordPlana = *cuerpo.PasswordPlana
		}

		trans, err := conexion.Pool().Begin(peticion.Context())
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		defer trans.Rollback(peticion.Context())

		acceso := &boveda.AccesoGuardado{
			Id:               id,
			Titulo:           cuerpo.Titulo,
			SistemaDestinoId: sistemaId,
			UsuarioExterno:   cuerpo.UsuarioExterno,
			PasswordPlana:    passwordPlana, // si vacío, el repo conserva la original
			Observaciones:    cuerpo.Observaciones,
			Tipo:             cuerpo.Tipo,
			Puerto:           cuerpo.Puerto,
		}
		if err := boveda.ActualizarAcceso(peticion.Context(), trans, clavesCifrado, acceso, sesion.UsuarioId); err != nil {
			if errors.Is(err, boveda.ErrAccesoNoEncontrado) {
				ResponderError(escritor, http.StatusNotFound, "ACCESO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoValidacionFallida, err.Error())
			return
		}

		_ = auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
			UsuarioId: &sesion.UsuarioId,
			SesionId:  &sesion.SesionId,
			Modulo:    "BOVEDA",
			Accion:    "ACCESO_ACTUALIZADO",
			Entidad:   "acceso_guardado",
			EntidadId: &id,
			DatosNuevos: map[string]any{
				"sistema_destino_id": sistemaId.String(),
				"password_cambiada":  passwordPlana != "",
			},
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		})

		if err := trans.Commit(peticion.Context()); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"acceso_actualizado": true})
	}
}
