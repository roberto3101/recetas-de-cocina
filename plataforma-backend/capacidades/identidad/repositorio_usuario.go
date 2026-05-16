package identidad

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
)

func InsertarUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, u *Usuario) error {
	return ejecutor.QueryRow(contexto, `
		INSERT INTO usuario (
			correo_electronico, password_hash, correo_electronico_verificado,
			estado, intentos_fallidos, creado_por
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, creado_en
	`,
		strings.ToLower(strings.TrimSpace(u.CorreoElectronico)),
		u.PasswordHash,
		u.CorreoElectronicoVerificado,
		u.Estado,
		u.IntentosFallidos,
		u.CreadoPor,
	).Scan(&u.Id, &u.CreadoEn)
}

func ConsultarUsuarioPorId(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID) (*Usuario, error) {
	u := &Usuario{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, correo_electronico, password_hash, correo_electronico_verificado,
		       estado, intentos_fallidos, bloqueado_hasta, ultimo_inicio_sesion_en,
		       creado_en, creado_por, actualizado_en, actualizado_por
		FROM usuario
		WHERE id = $1
	`, id).Scan(
		&u.Id, &u.CorreoElectronico, &u.PasswordHash, &u.CorreoElectronicoVerificado,
		&u.Estado, &u.IntentosFallidos, &u.BloqueadoHasta, &u.UltimoInicioSesionEn,
		&u.CreadoEn, &u.CreadoPor, &u.ActualizadoEn, &u.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUsuarioNoEncontrado
	}
	return u, err
}

func ConsultarUsuarioPorCorreo(contexto context.Context, ejecutor cockroach.EjecutorSql, correo string) (*Usuario, error) {
	u := &Usuario{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, correo_electronico, password_hash, correo_electronico_verificado,
		       estado, intentos_fallidos, bloqueado_hasta, ultimo_inicio_sesion_en,
		       creado_en, creado_por, actualizado_en, actualizado_por
		FROM usuario
		WHERE lower(correo_electronico) = lower($1)
	`, strings.TrimSpace(correo)).Scan(
		&u.Id, &u.CorreoElectronico, &u.PasswordHash, &u.CorreoElectronicoVerificado,
		&u.Estado, &u.IntentosFallidos, &u.BloqueadoHasta, &u.UltimoInicioSesionEn,
		&u.CreadoEn, &u.CreadoPor, &u.ActualizadoEn, &u.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUsuarioNoEncontrado
	}
	return u, err
}

func ActualizarPasswordUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID, nuevoHash string, actualizadoPor uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario
		SET password_hash = $2, actualizado_en = now(), actualizado_por = $3
		WHERE id = $1
	`, usuarioId, nuevoHash, actualizadoPor)
	return err
}

func RegistrarIntentoFallido(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID, maxIntentos int, minutosBloqueo int) error {
	bloqueoHasta := time.Now().Add(time.Duration(minutosBloqueo) * time.Minute)
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario
		SET intentos_fallidos = intentos_fallidos + 1,
		    estado = CASE
		        WHEN intentos_fallidos + 1 >= $2 THEN 'BLOQUEADO'
		        ELSE estado
		    END,
		    bloqueado_hasta = CASE
		        WHEN intentos_fallidos + 1 >= $2 THEN $3::TIMESTAMPTZ
		        ELSE bloqueado_hasta
		    END,
		    actualizado_en = now()
		WHERE id = $1
	`, usuarioId, maxIntentos, bloqueoHasta)
	return err
}

func ReiniciarIntentosFallidosYMarcarInicio(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario
		SET intentos_fallidos = 0,
		    bloqueado_hasta = NULL,
		    ultimo_inicio_sesion_en = now(),
		    actualizado_en = now()
		WHERE id = $1
	`, usuarioId)
	return err
}

func MarcarCorreoVerificadoYActivar(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario
		SET correo_electronico_verificado = true,
		    estado = 'ACTIVO',
		    actualizado_en = now()
		WHERE id = $1 AND estado = 'PENDIENTE'
	`, usuarioId)
	return err
}

func DesbloquearUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID, desbloqueadoPor uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE usuario
		SET estado = 'ACTIVO',
		    intentos_fallidos = 0,
		    bloqueado_hasta = NULL,
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado = 'BLOQUEADO'
	`, usuarioId, desbloqueadoPor)
	return err
}
