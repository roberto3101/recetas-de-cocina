package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

// ConstruirHandlerAutofillConHash devuelve un HTML que auto-envia el form de
// login del sistema externo usando el PASSWORD_HASH crudo de la BD externa
// como valor de la clave. Sirve solo si el sistema destino guarda la
// password en texto plano disfrazada de hash, o tiene un bypass que compara
// la entrada literal con stored_hash. Es un experimento.
func ConstruirHandlerAutofillConHash(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		sistemaId, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id sistema inválido")
			return
		}
		idExterno := chi.URLParam(peticion, "usuario_id")
		if idExterno == "" {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "usuario_id requerido")
			return
		}

		sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(peticion.Context(), conexion.Pool(), sistemaId)
		if err != nil {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoSistemaNoEncontrado, err.Error())
			return
		}
		if sistema.UrlLogin == "" {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoSistemaInvalido, "el sistema no tiene URL de login configurada")
			return
		}

		adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		usuario, err := adaptador.ConsultarUsuario(peticion.Context(), idExterno)
		if err != nil {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoUsuarioNoEncontrado, err.Error())
			return
		}
		if usuario.PasswordHash == "" {
			ResponderError(escritor, http.StatusNotFound, "HASH_NO_DISPONIBLE",
				"este usuario no tiene password_hash en la BD externa (o el adaptador no lo expone)")
			return
		}

		correoUsuario := usuario.CorreoElectronico
		if correoUsuario == "" {
			correoUsuario = idExterno
		}

		trans, err := conexion.Pool().Begin(peticion.Context())
		if err == nil {
			_ = auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
				UsuarioId: &sesion.UsuarioId,
				SesionId:  &sesion.SesionId,
				Modulo:    "AUTOFILL",
				Accion:    "AUTOFILL_HASH_GENERADO",
				Entidad:   "sistema_destino",
				EntidadId: &sistema.Id,
				DatosNuevos: map[string]any{
					"sistema_codigo":     sistema.Codigo,
					"id_externo_usuario": idExterno,
				},
				IpOrigen:      obtenerIpRemota(peticion),
				AgenteUsuario: peticion.UserAgent(),
			})
			_ = trans.Commit(peticion.Context())
		}

		paginaHtml := construirPaginaAutofill(
			sistema.UrlLogin,
			sistema.MetodoLogin,
			sistema.NombreCampoUsuario,
			sistema.NombreCampoPassword,
			correoUsuario,
			usuario.PasswordHash,
			sistema.Nombre+" (con hash)",
		)
		escritor.Header().Set("Content-Type", "text/html; charset=utf-8")
		escritor.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		escritor.WriteHeader(http.StatusOK)
		_, _ = escritor.Write([]byte(paginaHtml))
	}
}
