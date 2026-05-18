package http

import (
	"net/http"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type itemAccesoListado struct {
	Id               string `json:"id"`
	Titulo           string `json:"titulo"`
	SistemaDestinoId string `json:"sistema_destino_id"`
	UsuarioExterno   string `json:"usuario_externo"`
	Observaciones    string `json:"observaciones"`
	Tipo             string `json:"tipo"`
	Puerto           *int16 `json:"puerto,omitempty"`
	Estado           string `json:"estado"`
	CreadoEn         string `json:"creado_en"`
}

func ConstruirHandlerListarAccesos(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		listado, err := boveda.ListarAccesos(peticion.Context(), conexion.Pool(), clavesCifrado)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		items := make([]itemAccesoListado, 0, len(listado))
		for _, a := range listado {
			items = append(items, itemAccesoListado{
				Id:               a.Id.String(),
				Titulo:           a.Titulo,
				SistemaDestinoId: a.SistemaDestinoId.String(),
				UsuarioExterno:   a.UsuarioExterno,
				Observaciones:    a.Observaciones,
				Tipo:             a.Tipo,
				Puerto:           a.Puerto,
				Estado:           a.Estado,
				CreadoEn:         a.CreadoEn.UTC().Format("2006-01-02T15:04:05Z"),
			})
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"accesos": items, "total": len(items)})
	}
}
