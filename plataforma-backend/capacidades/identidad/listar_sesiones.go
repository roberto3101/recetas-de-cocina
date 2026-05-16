package identidad

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
)

func ListarSesionesActivasDeUsuario(contexto context.Context, conexion *cockroach.ConexionBaseDatos, usuarioId uuid.UUID) ([]SesionGlobal, error) {
	return ListarSesionesActivasDeUsuarioDb(contexto, conexion.Pool(), usuarioId)
}
