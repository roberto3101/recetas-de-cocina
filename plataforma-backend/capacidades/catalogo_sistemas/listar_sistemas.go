package catalogo_sistemas

import (
	"context"

	"sistemas-unificados/persistencia/cockroach"
)

func ListarSistemasActivos(contexto context.Context, conexion *cockroach.ConexionBaseDatos) ([]SistemaDestino, error) {
	return ListarSistemasActivosDb(contexto, conexion.Pool())
}
