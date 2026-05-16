package consumo_usuarios

import (
	"context"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/plataforma/cripto"
)

type ProbadorConexionPgx struct {
	ClavesCifrado *cripto.ClavesCifrado
}

func (p *ProbadorConexionPgx) IntentarConexion(contexto context.Context, conexion *catalogo_sistemas.ConexionLectura) error {
	pool, err := AbrirPoolExterno(contexto, conexion, p.ClavesCifrado)
	if err != nil {
		return err
	}
	defer pool.Close()
	return pool.Ping(contexto)
}
