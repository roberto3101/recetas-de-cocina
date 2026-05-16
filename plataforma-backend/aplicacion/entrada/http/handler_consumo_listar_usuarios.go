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

func ConstruirHandlerListarUsuariosDeSistema(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}
		idBruto := chi.URLParam(peticion, "id")
		sistemaId, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id inválido")
			return
		}

		filtro := adaptadores.FiltroBusquedaUsuarios{
			Texto:        peticion.URL.Query().Get("q"),
			Estado:       peticion.URL.Query().Get("estado"),
			Pagina:       LeerParametroEnteroConDefault(peticion, "pagina", 1, 0),
			TamanoPagina: LeerParametroEnteroConDefault(peticion, "tamano_pagina", 100, 200),
		}

		resultado, err := consumo_usuarios.ListarUsuariosDeSistema(peticion.Context(), conexion, resolver, consumo_usuarios.DatosListarUsuarios{
			SistemaId: sistemaId,
			Filtro:    filtro,
		})
		if err != nil {
			ResponderError(escritor, http.StatusBadGateway, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{
			"usuarios":        resultado.Usuarios,
			"total_registros": resultado.TotalRegistros,
			"pagina":          resultado.Pagina,
			"tamano_pagina":   resultado.TamanoPagina,
		})
	}
}
