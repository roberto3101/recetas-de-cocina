package http

import (
	"net/http"
	"sync"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type DiagnosticoSistemaItem struct {
	SistemaDestinoId string                  `json:"sistema_destino_id"`
	SistemaCodigo    string                  `json:"sistema_codigo"`
	SistemaNombre    string                  `json:"sistema_nombre"`
	Diagnostico      *adaptadores.Diagnostico `json:"diagnostico,omitempty"`
	Error            string                  `json:"error,omitempty"`
}

type RespuestaDiagnosticoExternos struct {
	Items []DiagnosticoSistemaItem `json:"items"`
}

func ConstruirHandlerDiagnosticoExternos(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		sistemas, err := catalogo_sistemas.ListarSistemasActivosDb(peticion.Context(), conexion.Pool())
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		items := make([]DiagnosticoSistemaItem, len(sistemas))
		var grupo sync.WaitGroup
		for i, sistema := range sistemas {
			items[i] = DiagnosticoSistemaItem{
				SistemaDestinoId: sistema.Id.String(),
				SistemaCodigo:    sistema.Codigo,
				SistemaNombre:    sistema.Nombre,
			}
			if !sistema.SoportaLectura {
				items[i].Error = "sistema no soporta lectura"
				continue
			}
			grupo.Add(1)
			go func(idx int, s catalogo_sistemas.SistemaDestino) {
				defer grupo.Done()
				adaptador, err := resolver.Resolver(s.ClaveAdaptador, s.Id)
				if err != nil {
					items[idx].Error = err.Error()
					return
				}
				diag, err := adaptador.Diagnosticar(peticion.Context())
				if err != nil {
					items[idx].Error = err.Error()
					return
				}
				items[idx].Diagnostico = diag
			}(i, sistema)
		}
		grupo.Wait()
		ResponderExito(escritor, http.StatusOK, RespuestaDiagnosticoExternos{Items: items})
	}
}
