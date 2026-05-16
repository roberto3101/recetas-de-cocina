package http

import (
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
	"sistemas-unificados/plataforma/sesion"
)

func RequiereSesion(conexion *cockroach.ConexionBaseDatos) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
			tokenPlano := sesion.LeerTokenDeCookie(peticion)
			if tokenPlano == "" {
				responder404Sigiloso(escritor)
				return
			}

			s, err := identidad.ValidarToken(peticion.Context(), conexion, tokenPlano)
			if err != nil {
				sesion.LimpiarCookieSesion(escritor)
				responder404Sigiloso(escritor)
				return
			}

			contextoConSesion := sesion.ContextoConSesion(peticion.Context(), s)
			siguiente.ServeHTTP(escritor, peticion.WithContext(contextoConSesion))
		})
	}
}

func RequiereSegundoFactor(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		s, err := sesion.SesionDesdeContexto(peticion.Context())
		if err != nil {
			responder404Sigiloso(escritor)
			return
		}
		if !s.SegundoFactorValidado {
			ResponderError(escritor, http.StatusForbidden, errores.CodigoSegundoFactorPendiente, "segundo factor pendiente de validación")
			return
		}
		siguiente.ServeHTTP(escritor, peticion)
	})
}

func responder404Sigiloso(escritor http.ResponseWriter) {
	escritor.Header().Set("Content-Type", "text/html; charset=utf-8")
	escritor.WriteHeader(http.StatusNotFound)
	_, _ = escritor.Write([]byte("<!doctype html><title>404</title><h1>404</h1>"))
}
