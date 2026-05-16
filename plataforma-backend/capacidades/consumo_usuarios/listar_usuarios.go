package consumo_usuarios

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
)

type DatosListarUsuarios struct {
	SistemaId uuid.UUID
	Filtro    adaptadores.FiltroBusquedaUsuarios
}

func ListarUsuariosDeSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, datos DatosListarUsuarios) (*adaptadores.ResultadoBusquedaUsuarios, error) {
	sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(contexto, conexion.Pool(), datos.SistemaId)
	if err != nil {
		return nil, err
	}
	if !sistema.SoportaLectura {
		return nil, errors.New("el sistema no soporta lectura de usuarios")
	}
	if sistema.Estado != catalogo_sistemas.EstadoSistemaActivo {
		return nil, errors.New("el sistema no está activo")
	}

	adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
	if err != nil {
		return nil, err
	}
	return adaptador.ListarUsuarios(contexto, datos.Filtro)
}
