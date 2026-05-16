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
	Descripcion         string `json:"descripcion"`
	UrlAcceso           string `json:"url_acceso"`
	Motor               string `json:"motor"`
	ClaveAdaptador      string `json:"clave_adaptador"`
	RequiereLoginGlobal bool   `json:"requiere_login_global"`
	SoportaLectura      bool   `json:"soporta_lectura"`
	SoportaAutoregistro bool   `json:"soporta_autoregistro"`
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
				Descripcion:         s.Descripcion,
				UrlAcceso:           s.UrlAcceso,
				Motor:               s.Motor,
				ClaveAdaptador:      s.ClaveAdaptador,
				RequiereLoginGlobal: s.RequiereLoginGlobal,
				SoportaLectura:      s.SoportaLectura,
				SoportaAutoregistro: s.SoportaAutoregistro,
				Estado:              s.Estado,
			})
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"sistemas": items, "total": len(items)})
	}
}
