package catalogo_sistemas

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
)

func ConsultarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, id uuid.UUID) (*SistemaDestino, error) {
	return ConsultarSistemaPorIdDb(contexto, conexion.Pool(), id)
}
