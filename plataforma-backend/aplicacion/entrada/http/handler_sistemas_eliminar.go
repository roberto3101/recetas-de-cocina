package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerEliminarSistema(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		idBruto := chi.URLParam(peticion, "id")
		id, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id inválido")
			return
		}
		if err := catalogo_sistemas.EliminarSistema(peticion.Context(), conexion, catalogo_sistemas.DatosEliminarSistema{
			SistemaId:     id,
			EliminadoPor:  sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		}); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"sistema_eliminado": true})
	}
}
