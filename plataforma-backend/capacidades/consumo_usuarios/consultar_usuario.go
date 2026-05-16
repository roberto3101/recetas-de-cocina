package consumo_usuarios

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
)

func ConsultarUsuarioDeSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, sistemaId uuid.UUID, idExterno string) (*adaptadores.UsuarioExterno, error) {
	sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(contexto, conexion.Pool(), sistemaId)
	if err != nil {
		return nil, err
	}
	if sistema.Estado != catalogo_sistemas.EstadoSistemaActivo {
		return nil, errors.New("el sistema no está activo")
	}
	adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
	if err != nil {
		return nil, err
	}
	return adaptador.ConsultarUsuario(contexto, idExterno)
}
