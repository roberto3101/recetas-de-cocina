package http

import (
	"net/http"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoGuardarAcceso struct {
	Titulo           string `json:"titulo"`
	SistemaDestinoId string `json:"sistema_destino_id"`
	UsuarioExterno   string `json:"usuario_externo"`
	PasswordPlana    string `json:"password"`
	Observaciones    string `json:"observaciones"`
}

func ConstruirHandlerGuardarAcceso(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		var cuerpo cuerpoGuardarAcceso
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}
		sistemaId, err := uuid.Parse(cuerpo.SistemaDestinoId)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "sistema_destino_id inválido")
			return
		}
		if cuerpo.PasswordPlana == "" {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoValidacionFallida, "password es obligatoria")
			return
		}

		passwordCifrada, err := cripto.CifrarConAesGcm(clavesCifrado.ClaveBoveda(), []byte(cuerpo.PasswordPlana))
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		acceso := &boveda.AccesoGuardado{
			Titulo:           cuerpo.Titulo,
			SistemaDestinoId: sistemaId,
			UsuarioExterno:   cuerpo.UsuarioExterno,
			PasswordCifrada:  passwordCifrada,
			Observaciones:    cuerpo.Observaciones,
			CreadoPor:        &sesion.UsuarioId,
		}

		trans, err := conexion.Pool().Begin(peticion.Context())
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		defer trans.Rollback(peticion.Context())

		if err := boveda.GuardarAcceso(peticion.Context(), trans, acceso); err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoValidacionFallida, err.Error())
			return
		}

		_ = auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
			UsuarioId: &sesion.UsuarioId,
			SesionId:  &sesion.SesionId,
			Modulo:    "BOVEDA",
			Accion:    "ACCESO_GUARDADO",
			Entidad:   "acceso_guardado",
			EntidadId: &acceso.Id,
			DatosNuevos: map[string]any{
				"sistema_destino_id": sistemaId.String(),
				"usuario_externo":    cuerpo.UsuarioExterno,
			},
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		})

		if err := trans.Commit(peticion.Context()); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusCreated, map[string]any{"id": acceso.Id.String()})
	}
}
