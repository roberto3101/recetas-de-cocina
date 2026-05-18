package catalogo_sistemas

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"sistemas-unificados/persistencia/cockroach"
)

func InsertarSistemaDestino(contexto context.Context, ejecutor cockroach.EjecutorSql, s *SistemaDestino) error {
	metodoLogin := s.MetodoLogin
	if metodoLogin == "" {
		metodoLogin = "POST"
	}
	nombreUsuario := s.NombreCampoUsuario
	if nombreUsuario == "" {
		nombreUsuario = "correo_electronico"
	}
	nombrePassword := s.NombreCampoPassword
	if nombrePassword == "" {
		nombrePassword = "password"
	}
	return ejecutor.QueryRow(contexto, `
		INSERT INTO sistema_destino (
			codigo, nombre, url_acceso, url_login,
			nombre_campo_usuario, nombre_campo_password, metodo_login,
			estado, creado_por
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, creado_en
	`,
		strings.ToLower(strings.TrimSpace(s.Codigo)),
		s.Nombre,
		s.UrlAcceso,
		s.UrlLogin,
		nombreUsuario,
		nombrePassword,
		metodoLogin,
		s.Estado,
		s.CreadoPor,
	).Scan(&s.Id, &s.CreadoEn)
}

func ConsultarSistemaPorIdDb(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID) (*SistemaDestino, error) {
	s := &SistemaDestino{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, codigo, nombre, url_acceso,
		       url_login, nombre_campo_usuario, nombre_campo_password, metodo_login,
		       estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM sistema_destino
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, id).Scan(
		&s.Id, &s.Codigo, &s.Nombre, &s.UrlAcceso,
		&s.UrlLogin, &s.NombreCampoUsuario, &s.NombreCampoPassword, &s.MetodoLogin,
		&s.Estado, &s.CreadoEn, &s.CreadoPor, &s.ActualizadoEn, &s.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSistemaNoEncontrado
	}
	return s, err
}

func ListarSistemasActivosDb(contexto context.Context, ejecutor cockroach.EjecutorSql) ([]SistemaDestino, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT id, codigo, nombre, url_acceso,
		       url_login, nombre_campo_usuario, nombre_campo_password, metodo_login,
		       estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM sistema_destino
		WHERE estado = 'ACTIVO'
		ORDER BY nombre ASC
	`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	resultado := make([]SistemaDestino, 0)
	for filas.Next() {
		s := SistemaDestino{}
		if err := filas.Scan(
			&s.Id, &s.Codigo, &s.Nombre, &s.UrlAcceso,
			&s.UrlLogin, &s.NombreCampoUsuario, &s.NombreCampoPassword, &s.MetodoLogin,
			&s.Estado, &s.CreadoEn, &s.CreadoPor, &s.ActualizadoEn, &s.ActualizadoPor,
		); err != nil {
			return nil, err
		}
		resultado = append(resultado, s)
	}
	return resultado, filas.Err()
}

func ActualizarSistemaDb(contexto context.Context, ejecutor cockroach.EjecutorSql, s *SistemaDestino, actualizadoPor uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE sistema_destino
		SET nombre = $2, url_acceso = $3, url_login = $4,
		    nombre_campo_usuario = $5, nombre_campo_password = $6, metodo_login = $7,
		    estado = $8,
		    actualizado_en = now(), actualizado_por = $9
		WHERE id = $1 AND estado != 'ELIMINADO'
	`,
		s.Id, s.Nombre, s.UrlAcceso, s.UrlLogin,
		s.NombreCampoUsuario, s.NombreCampoPassword, s.MetodoLogin,
		s.Estado, actualizadoPor,
	)
	return err
}

func EliminarSistemaLogico(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, eliminadoPor uuid.UUID) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE sistema_destino
		SET estado = 'ELIMINADO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, id, eliminadoPor)
	return err
}

func ErrorEsDuplicadoCodigoSistema(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if !errorAs(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}

func errorAs(err error, target any) bool {
	if err == nil {
		return false
	}
	type asunwrap interface{ Unwrap() error }
	for {
		if pe, ok := target.(**pgconn.PgError); ok {
			if pgErr, conv := err.(*pgconn.PgError); conv {
				*pe = pgErr
				return true
			}
		}
		u, ok := err.(asunwrap)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
}

func nullableTexto(valor string) any {
	if valor == "" {
		return nil
	}
	return valor
}
