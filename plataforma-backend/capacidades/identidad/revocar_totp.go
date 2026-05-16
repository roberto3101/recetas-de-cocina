package identidad

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosRevocarTotp struct {
	UsuarioId     uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

func RevocarTotpActivo(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosRevocarTotp) error {
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := RevocarTotpActivoDeUsuarioDb(contexto, trans, datos.UsuarioId); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.UsuarioId,
		SesionId:      datos.SesionId,
		Modulo:        "SEGUNDO_FACTOR",
		Accion:        "TOTP_REVOCADO",
		Entidad:       "usuario_totp",
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
