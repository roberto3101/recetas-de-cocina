package identidad

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
)

func InsertarTotpPendiente(contexto context.Context, ejecutor cockroach.EjecutorSql, t *UsuarioTotp) error {
	return ejecutor.QueryRow(contexto, `
		INSERT INTO usuario_totp (
			usuario_id, secreto_cifrado, codigos_respaldo, estado
		) VALUES ($1, $2, $3, 'PENDIENTE')
		RETURNING id, creado_en
	`,
		t.UsuarioId,
		t.SecretoCifrado,
		t.CodigosRespaldo,
	).Scan(&t.Id, &t.CreadoEn)
}

func ConsultarTotpActivoDeUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID) (*UsuarioTotp, error) {
	t := &UsuarioTotp{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, usuario_id, secreto_cifrado, codigos_respaldo, estado,
		       activado_en, revocado_en, creado_en
		FROM usuario_totp
		WHERE usuario_id = $1 AND estado = 'ACTIVO'
		LIMIT 1
	`, usuarioId).Scan(
		&t.Id, &t.UsuarioId, &t.SecretoCifrado, &t.CodigosRespaldo, &t.Estado,
		&t.ActivadoEn, &t.RevocadoEn, &t.CreadoEn,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTotpNoEncontrado
	}
	return t, err
}

func ConsultarTotpPendienteDeUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID) (*UsuarioTotp, error) {
	t := &UsuarioTotp{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, usuario_id, secreto_cifrado, codigos_respaldo, estado,
		       activado_en, revocado_en, creado_en
		FROM usuario_totp
		WHERE usuario_id = $1 AND estado = 'PENDIENTE'
		ORDER BY creado_en DESC
		LIMIT 1
	`, usuarioId).Scan(
		&t.Id, &t.UsuarioId, &t.SecretoCifrado, &t.CodigosRespaldo, &t.Estado,
		&t.ActivadoEn, &t.RevocadoEn, &t.CreadoEn,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTotpNoEncontrado
	}
	return t, err
}

func ActivarTotp(contexto context.Context, ejecutor cockroach.EjecutorSql, totpId uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario_totp
		SET estado = 'ACTIVO', activado_en = now()
		WHERE id = $1 AND estado = 'PENDIENTE'
	`, totpId)
	return err
}

func RevocarTotpActivoDeUsuarioDb(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario_totp
		SET estado = 'REVOCADO', revocado_en = now()
		WHERE usuario_id = $1 AND estado = 'ACTIVO'
	`, usuarioId)
	return err
}

func ConsumirCodigoRespaldo(contexto context.Context, ejecutor cockroach.EjecutorSql, totpId uuid.UUID, hashCodigoUsado string) (bool, error) {
	tag, err := ejecutor.Exec(contexto, `
		UPDATE usuario_totp
		SET codigos_respaldo = array_remove(codigos_respaldo, $2)
		WHERE id = $1 AND $2 = ANY(codigos_respaldo)
	`, totpId, hashCodigoUsado)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
