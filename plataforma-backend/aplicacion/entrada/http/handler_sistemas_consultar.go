package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerConsultarSistema(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		idBruto := chi.URLParam(peticion, "id")
		id, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id inválido")
			return
		}
		s, err := catalogo_sistemas.ConsultarSistema(peticion.Context(), conexion, id)
		if errors.Is(err, catalogo_sistemas.ErrSistemaNoEncontrado) {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoSistemaNoEncontrado, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{
			"id":                    s.Id.String(),
			"codigo":                s.Codigo,
			"nombre":                s.Nombre,
			"url_acceso":            s.UrlAcceso,
			"url_login":             s.UrlLogin,
			"nombre_campo_usuario":  s.NombreCampoUsuario,
			"nombre_campo_password": s.NombreCampoPassword,
			"metodo_login":          s.MetodoLogin,
			"estado":                s.Estado,
		})
	}
}
