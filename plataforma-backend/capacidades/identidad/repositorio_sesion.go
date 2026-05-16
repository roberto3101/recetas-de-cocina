package identidad

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
)

func InsertarSesion(contexto context.Context, ejecutor cockroach.EjecutorSql, s *SesionGlobal) error {
	return ejecutor.QueryRow(contexto, `
		INSERT INTO sesion_global (
			usuario_id, token_hash, refresh_token_hash, dispositivo_id,
			ip_origen, agente_usuario, segundo_factor_validado,
			expira_en, estado
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, emitida_en
	`,
		s.UsuarioId,
		s.TokenHash,
		s.RefreshTokenHash,
		s.DispositivoId,
		s.IpOrigen,
		s.AgenteUsuario,
		s.SegundoFactorValidado,
		s.ExpiraEn,
		s.Estado,
	).Scan(&s.Id, &s.EmitidaEn)
}

func ConsultarSesionPorTokenHash(contexto context.Context, ejecutor cockroach.EjecutorSql, tokenHash string) (*SesionGlobal, error) {
	s := &SesionGlobal{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, usuario_id, token_hash, refresh_token_hash, dispositivo_id,
		       ip_origen, agente_usuario, segundo_factor_validado,
		       emitida_en, expira_en, estado,
		       ultimo_acceso_en, revocado_en, revocado_por, coalesce(motivo_revocacion,'')
		FROM sesion_global
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&s.Id, &s.UsuarioId, &s.TokenHash, &s.RefreshTokenHash, &s.DispositivoId,
		&s.IpOrigen, &s.AgenteUsuario, &s.SegundoFactorValidado,
		&s.EmitidaEn, &s.ExpiraEn, &s.Estado,
		&s.UltimoAccesoEn, &s.RevocadoEn, &s.RevocadoPor, &s.MotivoRevocacion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSesionNoEncontrada
	}
	return s, err
}

func MarcarSegundoFactorValidadoEnSesion(contexto context.Context, ejecutor cockroach.EjecutorSql, sesionId uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE sesion_global
		SET segundo_factor_validado = true,
		    estado = 'ACTIVA',
		    ultimo_acceso_en = now()
		WHERE id = $1 AND estado IN ('PENDIENTE_SEGUNDO_FACTOR','ACTIVA')
	`, sesionId)
	return err
}

func ActualizarUltimoAccesoSesion(contexto context.Context, ejecutor cockroach.EjecutorSql, sesionId uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE sesion_global
		SET ultimo_acceso_en = now()
		WHERE id = $1
	`, sesionId)
	return err
}

func RevocarSesionEnDb(contexto context.Context, ejecutor cockroach.EjecutorSql, sesionId uuid.UUID, revocadoPor uuid.UUID, motivo string) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE sesion_global
		SET estado = 'REVOCADA',
		    revocado_en = now(),
		    revocado_por = $2,
		    motivo_revocacion = $3
		WHERE id = $1 AND estado IN ('ACTIVA','PENDIENTE_SEGUNDO_FACTOR')
	`, sesionId, revocadoPor, motivo)
	return err
}

func InvalidarTodasLasSesionesDeUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID, motivo string) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE sesion_global
		SET estado = 'INVALIDADA',
		    revocado_en = now(),
		    motivo_revocacion = $2
		WHERE usuario_id = $1 AND estado IN ('ACTIVA','PENDIENTE_SEGUNDO_FACTOR')
	`, usuarioId, motivo)
	return err
}

func ListarSesionesActivasDeUsuarioDb(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID) ([]SesionGlobal, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT id, usuario_id, token_hash, refresh_token_hash, dispositivo_id,
		       ip_origen, agente_usuario, segundo_factor_validado,
		       emitida_en, expira_en, estado,
		       ultimo_acceso_en, revocado_en, revocado_por, coalesce(motivo_revocacion,'')
		FROM sesion_global
		WHERE usuario_id = $1 AND estado = 'ACTIVA' AND expira_en > now()
		ORDER BY emitida_en DESC
	`, usuarioId)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	resultado := make([]SesionGlobal, 0)
	for filas.Next() {
		s := SesionGlobal{}
		if err := filas.Scan(
			&s.Id, &s.UsuarioId, &s.TokenHash, &s.RefreshTokenHash, &s.DispositivoId,
			&s.IpOrigen, &s.AgenteUsuario, &s.SegundoFactorValidado,
			&s.EmitidaEn, &s.ExpiraEn, &s.Estado,
			&s.UltimoAccesoEn, &s.RevocadoEn, &s.RevocadoPor, &s.MotivoRevocacion,
		); err != nil {
			return nil, err
		}
		resultado = append(resultado, s)
	}
	return resultado, filas.Err()
}
