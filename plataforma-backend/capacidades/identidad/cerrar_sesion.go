package identidad

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosCerrarSesion struct {
	SesionId      uuid.UUID
	UsuarioId     uuid.UUID
	IpOrigen      string
	AgenteUsuario string
	Motivo        string
}

func CerrarSesion(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosCerrarSesion) error {
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	motivo := datos.Motivo
	if motivo == "" {
		motivo = "logout solicitado"
	}

	if err := RevocarSesionEnDb(contexto, trans, datos.SesionId, datos.UsuarioId, motivo); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.UsuarioId,
		SesionId:      &datos.SesionId,
		Modulo:        "SESIONES",
		Accion:        "SESION_REVOCADA",
		Entidad:       "sesion_global",
		EntidadId:     &datos.SesionId,
		Detalle:       motivo,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
