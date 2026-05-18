package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"sistemas-unificados/aplicacion/entrada/http"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

func main() {
	registroEventos := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: nivelLogDesdeEntorno()}))
	slog.SetDefault(registroEventos)

	if err := godotenv.Load(); err != nil {
		slog.Warn("no se cargó archivo .env (puede no existir aún)", "detalle", err.Error())
	}

	clavesCifrado, err := cripto.CargarClavesDesdeEntorno()
	if err != nil {
		slog.Error("no se pudieron cargar claves de cifrado", "detalle", err.Error())
		os.Exit(1)
	}

	contextoArranque, cancelarArranque := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelarArranque()

	conexionBaseDatos, err := cockroach.AbrirConexion(contextoArranque, os.Getenv("BASE_DATOS_URL"))
	if err != nil {
		slog.Error("no se pudo conectar a la base de datos", "detalle", err.Error())
		os.Exit(1)
	}
	defer conexionBaseDatos.Cerrar()

	servidor := http.ConstruirServidor(http.OpcionesServidor{
		Direccion:         obtenerDireccionEscucha(),
		ConexionBaseDatos: conexionBaseDatos,
		ClavesCifrado:     clavesCifrado,
	})

	contextoSenales, cancelarSenales := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelarSenales()

	slog.Info("servidor iniciado", "direccion", servidor.Direccion())

	if err := servidor.EscucharHasta(contextoSenales); err != nil {
		slog.Error("servidor finalizó con error", "detalle", err.Error())
		os.Exit(1)
	}

	slog.Info("servidor detenido limpiamente")
}

func obtenerDireccionEscucha() string {
	direccion := os.Getenv("DIRECCION_ESCUCHA")
	if direccion == "" {
		return ":8080"
	}
	return direccion
}

// nivelLogDesdeEntorno lee NIVEL_LOG del .env: DEBUG | INFO | WARN | ERROR.
// Default INFO. En producción recomendamos WARN para reducir ruido.
func nivelLogDesdeEntorno() slog.Level {
	switch os.Getenv("NIVEL_LOG") {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
