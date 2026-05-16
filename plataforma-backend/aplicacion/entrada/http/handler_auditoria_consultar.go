package http

import (
	"net/http"
	"time"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/persistencia/repositorios"
	"sistemas-unificados/plataforma/gobierno/errores"
)

func ConstruirHandlerConsultarAuditoria(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if _, ok := ObtenerSesionDelContexto(escritor, peticion); !ok {
			return
		}

		filtro := repositorios.FiltroAuditoria{
			Modulo:       peticion.URL.Query().Get("modulo"),
			Accion:       peticion.URL.Query().Get("accion"),
			UsuarioId:    LeerParametroUuidOpcional(peticion, "usuario_id"),
			Pagina:       LeerParametroEnteroConDefault(peticion, "pagina", 1, 0),
			TamanoPagina: LeerParametroEnteroConDefault(peticion, "tamano_pagina", 100, 200),
		}

		if desdeStr := peticion.URL.Query().Get("desde"); desdeStr != "" {
			if t, err := time.Parse(time.RFC3339, desdeStr); err == nil {
				filtro.DesdeFecha = &t
			}
		}
		if hastaStr := peticion.URL.Query().Get("hasta"); hastaStr != "" {
			if t, err := time.Parse(time.RFC3339, hastaStr); err == nil {
				filtro.HastaFecha = &t
			}
		}

		resultado, err := repositorios.ConsultarAuditoria(peticion.Context(), conexion.Pool(), filtro)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{
			"registros":       resultado.Registros,
			"total_registros": resultado.TotalRegistros,
			"pagina":          resultado.Pagina,
			"tamano_pagina":   resultado.TamanoPagina,
		})
	}
}
