package http

import (
	"net/http"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type itemSistemaListado struct {
	Id                  string `json:"id"`
	Codigo              string `json:"codigo"`
	Nombre              string `json:"nombre"`
	UrlAcceso           string `json:"url_acceso"`
	UrlLogin            string `json:"url_login"`
	NombreCampoUsuario  string `json:"nombre_campo_usuario"`
	NombreCampoPassword string `json:"nombre_campo_password"`
	MetodoLogin         string `json:"metodo_login"`
	Estado              string `json:"estado"`
}

func ConstruirHandlerListarSistemas(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		listado, err := catalogo_sistemas.ListarSistemasActivos(peticion.Context(), conexion)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		items := make([]itemSistemaListado, 0, len(listado))
		for _, s := range listado {
			items = append(items, itemSistemaListado{
				Id:                  s.Id.String(),
				Codigo:              s.Codigo,
				Nombre:              s.Nombre,
				UrlAcceso:           s.UrlAcceso,
				UrlLogin:            s.UrlLogin,
				NombreCampoUsuario:  s.NombreCampoUsuario,
				NombreCampoPassword: s.NombreCampoPassword,
				MetodoLogin:         s.MetodoLogin,
				Estado:              s.Estado,
			})
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"sistemas": items, "total": len(items)})
	}
}
