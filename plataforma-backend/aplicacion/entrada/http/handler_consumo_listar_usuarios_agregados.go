package http

import (
	"net/http"

	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerListarUsuariosAgregados(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}

		filtro := adaptadores.FiltroBusquedaUsuarios{
			Texto:        peticion.URL.Query().Get("q"),
			Estado:       peticion.URL.Query().Get("estado"),
			Pagina:       LeerParametroEnteroConDefault(peticion, "pagina", 1, 0),
			TamanoPagina: LeerParametroEnteroConDefault(peticion, "tamano_pagina", 100, 200),
		}

		resultado, err := consumo_usuarios.ListarUsuariosDeTodosLosSistemas(peticion.Context(), conexion, resolver, consumo_usuarios.DatosListarAgregado{
			Filtro: filtro,
		})
		if err != nil {
			ResponderError(escritor, http.StatusBadGateway, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, resultado)
	}
}
