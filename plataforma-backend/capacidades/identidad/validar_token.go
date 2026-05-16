package identidad

import (
	"context"
	"errors"
	"time"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/sesion"
)

func ValidarToken(contexto context.Context, conexion *cockroach.ConexionBaseDatos, tokenPlano string) (*sesion.SesionActual, error) {
	if tokenPlano == "" {
		return nil, ErrSesionNoEncontrada
	}

	tokenHash := cripto.HashearTokenConSha256(tokenPlano)
	s, err := ConsultarSesionPorTokenHash(contexto, conexion.Pool(), tokenHash)
	if err != nil {
		if errors.Is(err, ErrSesionNoEncontrada) {
			return nil, ErrSesionNoEncontrada
		}
		return nil, err
	}

	if s.Estado == EstadoSesionRevocada || s.Estado == EstadoSesionInvalidada {
		return nil, ErrSesionRevocada
	}
	if time.Now().After(s.ExpiraEn) {
		return nil, ErrSesionExpirada
	}
	if s.Estado == EstadoSesionExpirada {
		return nil, ErrSesionExpirada
	}

	usuario, err := ConsultarUsuarioPorId(contexto, conexion.Pool(), s.UsuarioId)
	if err != nil {
		return nil, err
	}
	if usuario.Estado != EstadoUsuarioActivo {
		return nil, ErrUsuarioInactivo
	}

	_ = ActualizarUltimoAccesoSesion(contexto, conexion.Pool(), s.Id)

	return &sesion.SesionActual{
		SesionId:              s.Id,
		UsuarioId:             s.UsuarioId,
		CorreoElectronico:     usuario.CorreoElectronico,
		EmitidaEn:             s.EmitidaEn,
		ExpiraEn:              s.ExpiraEn,
		SegundoFactorValidado: s.SegundoFactorValidado,
		IpOrigen:              s.IpOrigen,
		AgenteUsuario:         s.AgenteUsuario,
	}, nil
}
