package boveda

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
)

var ErrAccesoNoEncontrado = errors.New("acceso guardado no encontrado")
var ErrUsuarioRequerido = errors.New("usuario es obligatorio")
var ErrPasswordRequerida = errors.New("password es obligatoria")
var ErrSistemaRequerido = errors.New("sistema_destino_id es obligatorio")

type AccesoGuardado struct {
	Id                uuid.UUID
	Titulo            string
	SistemaDestinoId  uuid.UUID
	UsuarioExterno    string
	PasswordCifrada   []byte
	Observaciones     string
	Estado            string
	CreadoEn          time.Time
	CreadoPor         *uuid.UUID
	ActualizadoEn     *time.Time
	ActualizadoPor    *uuid.UUID
}

func GuardarAcceso(contexto context.Context, ejecutor cockroach.EjecutorSql, a *AccesoGuardado) error {
	if strings.TrimSpace(a.UsuarioExterno) == "" {
		return ErrUsuarioRequerido
	}
	if len(a.PasswordCifrada) == 0 {
		return ErrPasswordRequerida
	}
	if a.SistemaDestinoId == uuid.Nil {
		return ErrSistemaRequerido
	}
	titulo := strings.TrimSpace(a.Titulo)
	return ejecutor.QueryRow(contexto, `
		INSERT INTO acceso_guardado (
			titulo, sistema_destino_id, usuario_externo,
			password_cifrada, observaciones, estado, creado_por
		) VALUES ($1, $2, $3, $4, $5, 'ACTIVO', $6)
		ON CONFLICT (sistema_destino_id, usuario_externo) WHERE estado = 'ACTIVO'
		DO UPDATE SET
			titulo = EXCLUDED.titulo,
			password_cifrada = EXCLUDED.password_cifrada,
			observaciones = EXCLUDED.observaciones,
			actualizado_en = now(),
			actualizado_por = EXCLUDED.creado_por
		RETURNING id, creado_en
	`,
		titulo, a.SistemaDestinoId, strings.TrimSpace(a.UsuarioExterno),
		a.PasswordCifrada, nullableTexto(a.Observaciones), a.CreadoPor,
	).Scan(&a.Id, &a.CreadoEn)
}

func ConsultarAccesoPorId(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID) (*AccesoGuardado, error) {
	a := &AccesoGuardado{}
	var observaciones *string
	err := ejecutor.QueryRow(contexto, `
		SELECT id, titulo, sistema_destino_id, usuario_externo,
		       password_cifrada, observaciones, estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM acceso_guardado
		WHERE id = $1 AND estado = 'ACTIVO'
	`, id).Scan(
		&a.Id, &a.Titulo, &a.SistemaDestinoId, &a.UsuarioExterno,
		&a.PasswordCifrada, &observaciones, &a.Estado, &a.CreadoEn, &a.CreadoPor, &a.ActualizadoEn, &a.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAccesoNoEncontrado
	}
	if observaciones != nil {
		a.Observaciones = *observaciones
	}
	return a, err
}

func ListarAccesosActivos(contexto context.Context, ejecutor cockroach.EjecutorSql) ([]AccesoGuardado, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT id, titulo, sistema_destino_id, usuario_externo,
		       coalesce(observaciones,''), estado, creado_en
		FROM acceso_guardado
		WHERE estado = 'ACTIVO'
		ORDER BY creado_en DESC
	`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	resultado := make([]AccesoGuardado, 0)
	for filas.Next() {
		a := AccesoGuardado{}
		if err := filas.Scan(
			&a.Id, &a.Titulo, &a.SistemaDestinoId, &a.UsuarioExterno,
			&a.Observaciones, &a.Estado, &a.CreadoEn,
		); err != nil {
			return nil, err
		}
		resultado = append(resultado, a)
	}
	return resultado, filas.Err()
}

func ActualizarAcceso(contexto context.Context, ejecutor cockroach.EjecutorSql, a *AccesoGuardado, actualizadoPor uuid.UUID) error {
	if strings.TrimSpace(a.UsuarioExterno) == "" {
		return ErrUsuarioRequerido
	}
	if a.SistemaDestinoId == uuid.Nil {
		return ErrSistemaRequerido
	}
	tag, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET titulo = $2,
		    sistema_destino_id = $3,
		    usuario_externo = $4,
		    password_cifrada = CASE WHEN $5::BYTES IS NULL THEN password_cifrada ELSE $5 END,
		    observaciones = $6,
		    actualizado_en = now(),
		    actualizado_por = $7
		WHERE id = $1 AND estado = 'ACTIVO'
	`,
		a.Id,
		strings.TrimSpace(a.Titulo),
		a.SistemaDestinoId,
		strings.TrimSpace(a.UsuarioExterno),
		nullableBytes(a.PasswordCifrada),
		nullableTexto(a.Observaciones),
		actualizadoPor,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAccesoNoEncontrado
	}
	return nil
}

func EliminarAccesoLogico(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, eliminadoPor uuid.UUID) error {
	tag, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET estado = 'ELIMINADO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado = 'ACTIVO'
	`, id, eliminadoPor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAccesoNoEncontrado
	}
	return nil
}

func nullableTexto(valor string) any {
	if valor == "" {
		return nil
	}
	return valor
}

func nullableBytes(valor []byte) any {
	if len(valor) == 0 {
		return nil
	}
	return valor
}
