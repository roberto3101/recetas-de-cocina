package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerProbarConexionSistema(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		idBruto := chi.URLParam(peticion, "id")
		id, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id inválido")
			return
		}
		probador := &consumo_usuarios.ProbadorConexionPgx{ClavesCifrado: clavesCifrado}
		if err := catalogo_sistemas.ProbarConexion(peticion.Context(), conexion, id, probador); err != nil {
			ResponderError(escritor, http.StatusServiceUnavailable, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"conexion_ok": true})
	}
}
