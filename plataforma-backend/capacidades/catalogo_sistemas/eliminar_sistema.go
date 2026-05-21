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

// ResultadoEliminarSistema reporta cuántos accesos fueron revocados en
// cascada al desactivar el sistema. El frontend lo muestra al usuario para
// que tenga feedback claro del efecto de su acción.
type ResultadoEliminarSistema struct {
	AccesosRevocados int
}

func EliminarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosEliminarSistema) (*ResultadoEliminarSistema, error) {
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	accesosRevocados, err := EliminarSistemaLogico(contexto, trans, datos.SistemaId, datos.EliminadoPor)
	if err != nil {
		return nil, err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.EliminadoPor,
		SesionId:  datos.SesionId,
		Modulo:    "CATALOGO_SISTEMAS",
		Accion:    "SISTEMA_ELIMINADO",
		Entidad:   "sistema_destino",
		EntidadId: &datos.SistemaId,
		DatosNuevos: map[string]any{
			"accesos_revocados_en_cascada": accesosRevocados,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return nil, err
	}

	if err := trans.Commit(contexto); err != nil {
		return nil, err
	}
	return &ResultadoEliminarSistema{AccesosRevocados: accesosRevocados}, nil
}
