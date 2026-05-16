package identidad

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

const emisorOtpAuth = "GestorCredencialesCodeplex"

type DatosIniciarActivacionTotp struct {
	UsuarioId         uuid.UUID
	CorreoElectronico string
	SesionId          *uuid.UUID
	IpOrigen          string
	AgenteUsuario     string
}

type ResultadoIniciarActivacionTotp struct {
	SecretoBase32   string
	UrlOtpAuth      string
	CodigosRespaldo []string
	TotpId          uuid.UUID
}

func IniciarActivacionTotp(contexto context.Context, conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado, datos DatosIniciarActivacionTotp) (*ResultadoIniciarActivacionTotp, error) {
	existenteActivo, err := ConsultarTotpActivoDeUsuario(contexto, conexion.Pool(), datos.UsuarioId)
	if err != nil && !errors.Is(err, ErrTotpNoEncontrado) {
		return nil, err
	}
	if existenteActivo != nil {
		return nil, ErrTotpYaActivo
	}

	clave, err := cripto.GenerarSecretoTotp(emisorOtpAuth, datos.CorreoElectronico)
	if err != nil {
		return nil, err
	}

	secretoCifrado, err := cripto.CifrarConAesGcm(clavesCifrado.ClaveBoveda(), []byte(clave.Secret()))
	if err != nil {
		return nil, err
	}

	codigosPlanos, codigosHash, err := cripto.GenerarCodigosRespaldoTotp()
	if err != nil {
		return nil, err
	}

	totp := &UsuarioTotp{
		UsuarioId:       datos.UsuarioId,
		SecretoCifrado:  secretoCifrado,
		CodigosRespaldo: codigosHash,
		Estado:          EstadoTotpPendiente,
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	if err := InsertarTotpPendiente(contexto, trans, totp); err != nil {
		return nil, err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.UsuarioId,
		SesionId:      datos.SesionId,
		Modulo:        "SEGUNDO_FACTOR",
		Accion:        "TOTP_ACTIVACION_INICIADA",
		Entidad:       "usuario_totp",
		EntidadId:     &totp.Id,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return nil, err
	}

	if err := trans.Commit(contexto); err != nil {
		return nil, err
	}

	return &ResultadoIniciarActivacionTotp{
		SecretoBase32:   clave.Secret(),
		UrlOtpAuth:      clave.URL(),
		CodigosRespaldo: codigosPlanos,
		TotpId:          totp.Id,
	}, nil
}
