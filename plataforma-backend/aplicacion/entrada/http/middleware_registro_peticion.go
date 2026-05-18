package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// rutasSilenciosas son endpoints sondeados con frecuencia donde un 2xx no aporta
// info al log. Solo se loguea si responden con error o tardan demasiado.
var rutasSilenciosas = []string{
	"/salud",
	"/cocina/identidad/perfil",
	"/cocina/sistemas",
	"/cocina/boveda/accesos",
}

const umbralLento = 500 * time.Millisecond

func esRutaSilenciosa(ruta string) bool {
	for _, prefijo := range rutasSilenciosas {
		if ruta == prefijo || strings.HasPrefix(ruta, prefijo+"/") || strings.HasPrefix(ruta, prefijo+"?") {
			return true
		}
	}
	return false
}

// RegistrarPeticionEntrante loguea cada petición HTTP. Filtra el ruido:
//   - 2xx/3xx en rutas "silenciosas" (perfil, lista, salud) → silencio
//   - 4xx/5xx → siempre log
//   - >500ms → siempre log (auditar latencia)
func RegistrarPeticionEntrante(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		inicio := time.Now()
		envoltorio := middleware.NewWrapResponseWriter(escritor, peticion.ProtoMajor)

		siguiente.ServeHTTP(envoltorio, peticion)

		duracion := time.Since(inicio)
		estado := envoltorio.Status()
		ruta := peticion.URL.Path

		// Decisión de logging
		exitoso := estado >= 200 && estado < 400
		silencioso := esRutaSilenciosa(ruta) && exitoso && duracion < umbralLento

		if silencioso {
			return
		}

		nivel := slog.LevelInfo
		if estado >= 500 {
			nivel = slog.LevelError
		} else if estado >= 400 {
			nivel = slog.LevelWarn
		}

		slog.Log(peticion.Context(), nivel, "peticion_atendida",
			"metodo", peticion.Method,
			"ruta", ruta,
			"estado", estado,
			"bytes", envoltorio.BytesWritten(),
			"duracion_ms", duracion.Milliseconds(),
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
