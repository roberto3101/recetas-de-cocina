package catalogo_sistemas

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
)

var ErrCredencialUsuarioNoEncontrada = errors.New("credencial de usuario externo no encontrada")

type CredencialUsuarioExterno struct {
	Id                uuid.UUID
	SistemaDestinoId  uuid.UUID
	IdExternoUsuario  string
	CorreoUsuario     string
	PasswordCifrada   []byte
	Estado            string
	CreadoEn          time.Time
	CreadoPor         *uuid.UUID
	ActualizadoEn     *time.Time
	ActualizadoPor    *uuid.UUID
}

func GuardarCredencialUsuarioExterno(contexto context.Context, ejecutor cockroach.EjecutorSql, c *CredencialUsuarioExterno) error {
	return ejecutor.QueryRow(contexto, `
		INSERT INTO credencial_usuario_externo (
			sistema_destino_id, id_externo_usuario, correo_usuario,
			password_cifrada, estado, creado_por
		) VALUES ($1, $2, $3, $4, 'ACTIVO', $5)
		ON CONFLICT (sistema_destino_id, id_externo_usuario) WHERE estado = 'ACTIVO'
		DO UPDATE SET
			password_cifrada = EXCLUDED.password_cifrada,
			correo_usuario = EXCLUDED.correo_usuario,
			actualizado_en = now(),
			actualizado_por = EXCLUDED.creado_por
		RETURNING id, creado_en
	`,
		c.SistemaDestinoId, c.IdExternoUsuario, nullableTexto(c.CorreoUsuario),
		c.PasswordCifrada, c.CreadoPor,
	).Scan(&c.Id, &c.CreadoEn)
}

func ConsultarCredencialUsuarioExterno(contexto context.Context, ejecutor cockroach.EjecutorSql, sistemaId uuid.UUID, idExterno string) (*CredencialUsuarioExterno, error) {
	c := &CredencialUsuarioExterno{}
	var correo *string
	err := ejecutor.QueryRow(contexto, `
		SELECT id, sistema_destino_id, id_externo_usuario, correo_usuario,
		       password_cifrada, estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM credencial_usuario_externo
		WHERE sistema_destino_id = $1 AND id_externo_usuario = $2 AND estado = 'ACTIVO'
		ORDER BY creado_en DESC
		LIMIT 1
	`, sistemaId, idExterno).Scan(
		&c.Id, &c.SistemaDestinoId, &c.IdExternoUsuario, &correo,
		&c.PasswordCifrada, &c.Estado, &c.CreadoEn, &c.CreadoPor, &c.ActualizadoEn, &c.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCredencialUsuarioNoEncontrada
	}
	if correo != nil {
		c.CorreoUsuario = *correo
	}
	return c, err
}
