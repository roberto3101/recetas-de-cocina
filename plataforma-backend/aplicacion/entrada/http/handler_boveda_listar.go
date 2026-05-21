package http

import (
	"net/http"

	"github.com/google/uuid"

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

// Query params soportados por GET /cocina/boveda/accesos:
//   limite       int  (default 50, max 200)
//   offset       int  (default 0)
//   orden        str  ("creado_en" | "titulo", default "creado_en")
//   direccion    str  ("asc" | "desc", default "desc")
//   estado       str  ("ACTIVO" | "REVOCADO" | "" para ambos, default "")
//   sistema_id   uuid (filtra a un sistema específico)
//   q            str  (búsqueda de texto en título, nombre y URL del sistema)
//
// Respuesta:
//   {accesos: [...], total: N, limite: L, offset: O}
//   total es el conteo filtrado de la BD sin paginar — refleja exactamente
//   cuántas filas matchean q+sistema_id+estado, para que el cliente sepa
//   cuántas páginas hay.
func ConstruirHandlerListarAccesos(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}

		// Defaults seguros: 50 por página, descendente por fecha.
		limite := LeerParametroEnteroConDefault(peticion, "limite", 50, 200)
		offset := LeerParametroEnteroConDefault(peticion, "offset", 0, 0)
		orden := peticion.URL.Query().Get("orden")
		direccion := peticion.URL.Query().Get("direccion")
		estado := peticion.URL.Query().Get("estado")
		busqueda := peticion.URL.Query().Get("q")

		// sistema_id: validamos UUID parseable. Si viene basura, lo ignoramos
		// silenciosamente — no rompemos el endpoint por un param malformado.
		sistemaId := ""
		if raw := peticion.URL.Query().Get("sistema_id"); raw != "" {
			if _, err := uuid.Parse(raw); err == nil {
				sistemaId = raw
			}
		}

		opts := boveda.OpcionesListarAccesos{
			OrdenarPor:       orden,
			Direccion:        direccion,
			Limite:           limite,
			Offset:           offset,
			FiltroEstado:     estado,
			SistemaDestinoId: sistemaId,
			Busqueda:         busqueda,
		}

		// Primero el total (rápido, COUNT con índice). Después la página.
		total, err := boveda.ContarAccesos(peticion.Context(), conexion.Pool(), opts)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		listado, err := boveda.ListarAccesos(peticion.Context(), conexion.Pool(), clavesCifrado, opts)
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
		ResponderExito(escritor, http.StatusOK, map[string]any{
			"accesos": items,
			"total":   total,
			"limite":  limite,
			"offset":  offset,
		})
	}
}
