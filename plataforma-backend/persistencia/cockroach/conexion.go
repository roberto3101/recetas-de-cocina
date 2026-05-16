package cockroach

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ConexionBaseDatos struct {
	pool *pgxpool.Pool
}

func AbrirConexion(contexto context.Context, urlConexion string) (*ConexionBaseDatos, error) {
	if urlConexion == "" {
		return nil, errors.New("BASE_DATOS_URL no está definida")
	}

	configuracion, err := pgxpool.ParseConfig(urlConexion)
	if err != nil {
		return nil, err
	}
	configuracion.MaxConns = 20
	configuracion.MinConns = 2
	configuracion.MaxConnLifetime = 30 * time.Minute
	configuracion.MaxConnIdleTime = 5 * time.Minute
	configuracion.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(contexto, configuracion)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(contexto); err != nil {
		pool.Close()
		return nil, err
	}

	return &ConexionBaseDatos{pool: pool}, nil
}

func (c *ConexionBaseDatos) Pool() *pgxpool.Pool {
	return c.pool
}

func (c *ConexionBaseDatos) VerificarSalud(contexto context.Context) error {
	return c.pool.Ping(contexto)
}

func (c *ConexionBaseDatos) Cerrar() {
	if c.pool != nil {
		c.pool.Close()
	}
}
