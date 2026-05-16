package identidad

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)


type DatosConfirmarActivacionTotp struct {
	UsuarioId       uuid.UUID
	CodigoIngresado string
	SesionId        *uuid.UUID
	IpOrigen        string
	AgenteUsuario   string
}

func ConfirmarActivacionTotp(contexto context.Context, conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado, datos DatosConfirmarActivacionTotp) error {
	pendiente, err := ConsultarTotpPendienteDeUsuario(contexto, conexion.Pool(), datos.UsuarioId)
	if err != nil {
		if errors.Is(err, ErrTotpNoEncontrado) {
			return ErrTotpNoEncontrado
		}
		return err
	}

	secretoPlano, err := cripto.DescifrarConAesGcm(clavesCifrado.ClaveBoveda(), pendiente.SecretoCifrado)
	if err != nil {
		return err
	}

	if !cripto.ValidarCodigoTotp(datos.CodigoIngresado, string(secretoPlano)) {
		return ErrCodigoTotpInvalido
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := ActivarTotp(contexto, trans, pendiente.Id); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.UsuarioId,
		SesionId:      datos.SesionId,
		Modulo:        "SEGUNDO_FACTOR",
		Accion:        "TOTP_ACTIVADO",
		Entidad:       "usuario_totp",
		EntidadId:     &pendiente.Id,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
