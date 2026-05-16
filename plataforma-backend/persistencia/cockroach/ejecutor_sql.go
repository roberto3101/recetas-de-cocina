package cockroach

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EjecutorSql interface {
	Exec(contexto context.Context, sql string, argumentos ...any) (pgconn.CommandTag, error)
	Query(contexto context.Context, sql string, argumentos ...any) (pgx.Rows, error)
	QueryRow(contexto context.Context, sql string, argumentos ...any) pgx.Row
}
