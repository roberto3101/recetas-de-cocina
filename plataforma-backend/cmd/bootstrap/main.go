package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no se cargó .env", "detalle", err.Error())
	}

	correo := os.Getenv("OPERADOR_INICIAL_CORREO")
	password := os.Getenv("OPERADOR_INICIAL_PASSWORD")
	if correo == "" || password == "" {
		fmt.Println("ERROR: OPERADOR_INICIAL_CORREO y OPERADOR_INICIAL_PASSWORD requeridos")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conexion, err := cockroach.AbrirConexion(ctx, os.Getenv("BASE_DATOS_URL"))
	if err != nil {
		fmt.Printf("ERROR conexión: %v\n", err)
		os.Exit(1)
	}
	defer conexion.Cerrar()

	resultado, err := identidad.RegistrarUsuarioYActivar(ctx, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correo,
		PasswordPlana:     password,
		IpOrigen:          "bootstrap",
		AgenteUsuario:     "cmd/bootstrap",
	})
	if err != nil {
		if errors.Is(err, identidad.ErrCorreoYaRegistrado) {
			fmt.Printf("OPERADOR_YA_EXISTE: %s\n", correo)
			return
		}
		fmt.Printf("ERROR registro: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OPERADOR_CREADO id=%s correo=%s estado=%s\n", resultado.UsuarioId, correo, resultado.Estado)
}
