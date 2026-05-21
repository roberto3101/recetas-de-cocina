package catalogo_sistemas

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosReactivarSistema struct {
	SistemaId     uuid.UUID
	ReactivadoPor uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

type ResultadoReactivarSistema struct {
	AccesosReactivados int
}

// ReactivarSistema saca al sistema del archivo (estado ELIMINADO → ACTIVO)
// y revive en cascada todos sus accesos que estuviesen REVOCADOS.
//
// Espejo simétrico de EliminarSistema. Como toda operación de cambio de
// estado, va en una transacción única con audit log incluido — si algo
// falla, rollback y nada quedó a medias.
func ReactivarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosReactivarSistema) (*ResultadoReactivarSistema, error) {
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	accesosReactivados, err := ReactivarSistemaLogico(contexto, trans, datos.SistemaId, datos.ReactivadoPor)
	if err != nil {
		return nil, err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.ReactivadoPor,
		SesionId:  datos.SesionId,
		Modulo:    "CATALOGO_SISTEMAS",
		Accion:    "SISTEMA_REACTIVADO",
		Entidad:   "sistema_destino",
		EntidadId: &datos.SistemaId,
		DatosNuevos: map[string]any{
			"accesos_reactivados_en_cascada": accesosReactivados,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return nil, err
	}

	if err := trans.Commit(contexto); err != nil {
		return nil, err
	}
	return &ResultadoReactivarSistema{AccesosReactivados: accesosReactivados}, nil
}
