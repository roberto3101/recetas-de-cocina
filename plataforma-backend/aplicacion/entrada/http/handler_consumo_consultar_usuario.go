package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerConsultarUsuarioExterno(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		sistemaIdBruto := chi.URLParam(peticion, "id")
		sistemaId, err := uuid.Parse(sistemaIdBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id sistema inválido")
			return
		}
		idExterno := chi.URLParam(peticion, "usuario_id")
		if idExterno == "" {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "usuario_id requerido")
			return
		}
		usuario, err := consumo_usuarios.ConsultarUsuarioDeSistema(peticion.Context(), conexion, resolver, sistemaId, idExterno)
		if err != nil {
			ResponderError(escritor, http.StatusBadGateway, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, usuario)
	}
}
