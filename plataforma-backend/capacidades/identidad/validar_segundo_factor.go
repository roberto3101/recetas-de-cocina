package identidad

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosValidarSegundoFactor struct {
	UsuarioId       uuid.UUID
	SesionId        uuid.UUID
	CodigoIngresado string
	IpOrigen        string
	AgenteUsuario   string
}

func ValidarSegundoFactor(contexto context.Context, conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado, datos DatosValidarSegundoFactor) error {
	activo, err := ConsultarTotpActivoDeUsuario(contexto, conexion.Pool(), datos.UsuarioId)
	if err != nil {
		if errors.Is(err, ErrTotpNoEncontrado) {
			return ErrTotpNoActivo
		}
		return err
	}

	secretoPlano, err := cripto.DescifrarConAesGcm(clavesCifrado.ClaveBoveda(), activo.SecretoCifrado)
	if err != nil {
		return err
	}

	codigoValido := cripto.ValidarCodigoTotp(datos.CodigoIngresado, string(secretoPlano))
	codigoRespaldoConsumido := false

	if !codigoValido {
		hashIngresado := cripto.HashearTokenConSha256(datos.CodigoIngresado)
		consumido, err := ConsumirCodigoRespaldo(contexto, conexion.Pool(), activo.Id, hashIngresado)
		if err != nil {
			return err
		}
		codigoRespaldoConsumido = consumido
	}

	if !codigoValido && !codigoRespaldoConsumido {
		return ErrCodigoTotpInvalido
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := MarcarSegundoFactorValidadoEnSesion(contexto, trans, datos.SesionId); err != nil {
		return err
	}

	if err := ReiniciarIntentosFallidosYMarcarInicio(contexto, trans, datos.UsuarioId); err != nil {
		return err
	}

	accion := "SEGUNDO_FACTOR_VALIDADO"
	if codigoRespaldoConsumido {
		accion = "CODIGO_RESPALDO_USADO"
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.UsuarioId,
		SesionId:      &datos.SesionId,
		Modulo:        "SEGUNDO_FACTOR",
		Accion:        accion,
		Entidad:       "sesion_global",
		EntidadId:     &datos.SesionId,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
