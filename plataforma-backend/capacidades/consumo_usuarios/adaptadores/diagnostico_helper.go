package adaptadores

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func diagnosticarPool(contexto context.Context, pool *pgxpool.Pool, baseDatos string) (*Diagnostico, error) {
	filas, err := pool.Query(contexto, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE'
		  AND table_schema NOT IN ('crdb_internal','information_schema','pg_catalog','pg_extension')
		ORDER BY table_schema, table_name
	`)
	if err != nil {
		return nil, err
	}
	type entrada struct{ schema, nombre string }
	var lista []entrada
	for filas.Next() {
		var e entrada
		if err := filas.Scan(&e.schema, &e.nombre); err != nil {
			filas.Close()
			return nil, err
		}
		lista = append(lista, e)
	}
	filas.Close()

	resultado := &Diagnostico{BaseDatos: baseDatos, Tablas: make([]TablaConCount, 0, len(lista))}
	for _, e := range lista {
		var n int64
		consulta := fmt.Sprintf(`SELECT count(*) FROM %q.%q`, e.schema, e.nombre)
		if err := pool.QueryRow(contexto, consulta).Scan(&n); err != nil {
			resultado.Tablas = append(resultado.Tablas, TablaConCount{Schema: e.schema, Nombre: e.nombre, Filas: -1})
			continue
		}
		resultado.Tablas = append(resultado.Tablas, TablaConCount{Schema: e.schema, Nombre: e.nombre, Filas: n})
	}
	return resultado, nil
}
