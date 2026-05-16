package consumo_usuarios

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

func ConstruirProveedorPoolExterno(conexionInterna *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) adaptadores.ProveedorPoolExterno {
	return func(contexto context.Context, idSistema uuid.UUID) (*pgxpool.Pool, error) {
		conexionLectura, err := catalogo_sistemas.ConsultarConexionLecturaActiva(contexto, conexionInterna.Pool(), idSistema)
		if err != nil {
			if errors.Is(err, catalogo_sistemas.ErrConexionLecturaNoEncontrada) {
				return nil, errors.New("el sistema no tiene conexión de lectura activa")
			}
			return nil, err
		}
		return AbrirPoolExterno(contexto, conexionLectura, clavesCifrado)
	}
}
