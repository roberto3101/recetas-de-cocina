package catalogo_sistemas

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
)

type ProbadorConexion interface {
	IntentarConexion(contexto context.Context, conexion *ConexionLectura) error
}

func ProbarConexion(contexto context.Context, conexion *cockroach.ConexionBaseDatos, sistemaId uuid.UUID, probador ProbadorConexion) error {
	conexionLectura, err := ConsultarConexionLecturaActiva(contexto, conexion.Pool(), sistemaId)
	if err != nil {
		if errors.Is(err, ErrConexionLecturaNoEncontrada) {
			return errors.New("el sistema no tiene una conexión de lectura activa configurada")
		}
		return err
	}

	if err := probador.IntentarConexion(contexto, conexionLectura); err != nil {
		return fmt.Errorf("conexión externa falló: %w", err)
	}
	return nil
}
