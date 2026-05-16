package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoCambiarPasswordUsuarioExterno struct {
	PasswordNueva string `json:"password_nueva"`
}

func ConstruirHandlerCambiarPasswordUsuarioExterno(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		sistemaId, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id sistema inválido")
			return
		}
		idExterno := chi.URLParam(peticion, "usuario_id")
		if idExterno == "" {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "usuario_id requerido")
			return
		}

		var cuerpo cuerpoCambiarPasswordUsuarioExterno
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}
		if len(cuerpo.PasswordNueva) < 8 {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoValidacionFallida, "la nueva contraseña debe tener al menos 8 caracteres")
			return
		}

		if err := consumo_usuarios.CambiarPasswordUsuarioEnSistema(peticion.Context(), conexion, resolver, clavesCifrado, consumo_usuarios.DatosCambiarPasswordUsuarioExterno{
			SistemaId:     sistemaId,
			IdExterno:     idExterno,
			PasswordNueva: cuerpo.PasswordNueva,
			OperadorId:    sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		}); err != nil {
			ResponderError(escritor, http.StatusBadGateway, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"password_externo_cambiado": true})
	}
}
