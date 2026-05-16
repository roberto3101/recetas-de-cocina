package catalogo_sistemas

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosEliminarSistema struct {
	SistemaId     uuid.UUID
	EliminadoPor  uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

func EliminarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosEliminarSistema) error {
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := EliminarSistemaLogico(contexto, trans, datos.SistemaId, datos.EliminadoPor); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.EliminadoPor,
		SesionId:      datos.SesionId,
		Modulo:        "CATALOGO_SISTEMAS",
		Accion:        "SISTEMA_ELIMINADO",
		Entidad:       "sistema_destino",
		EntidadId:     &datos.SistemaId,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
