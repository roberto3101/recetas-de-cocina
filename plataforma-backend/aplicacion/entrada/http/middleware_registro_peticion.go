package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RegistrarPeticionEntrante(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		inicio := time.Now()
		envoltorio := middleware.NewWrapResponseWriter(escritor, peticion.ProtoMajor)

		siguiente.ServeHTTP(envoltorio, peticion)

		slog.Info("peticion_atendida",
			"metodo", peticion.Method,
			"ruta", peticion.URL.Path,
			"estado", envoltorio.Status(),
			"bytes", envoltorio.BytesWritten(),
			"duracion_ms", time.Since(inicio).Milliseconds(),
			"ip", obtenerIpRemota(peticion),
		)
	})
}

func obtenerIpRemota(peticion *http.Request) string {
	if cabeceraReenvio := peticion.Header.Get("X-Forwarded-For"); cabeceraReenvio != "" {
		return cabeceraReenvio
	}
	if cabeceraReal := peticion.Header.Get("X-Real-IP"); cabeceraReal != "" {
		return cabeceraReal
	}
	return peticion.RemoteAddr
}
