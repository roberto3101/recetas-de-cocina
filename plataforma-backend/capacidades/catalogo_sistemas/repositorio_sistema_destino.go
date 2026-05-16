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
		nombreUsuario = "usuario"
	}
	nombrePassword := s.NombreCampoPassword
	if nombrePassword == "" {
		nombrePassword = "password"
	}
	return ejecutor.QueryRow(contexto, `
		INSERT INTO sistema_destino (
			codigo, nombre, descripcion, url_acceso, url_login,
			nombre_campo_usuario, nombre_campo_password, metodo_login,
			motor, clave_adaptador,
			requiere_login_global, soporta_lectura, soporta_autoregistro,
			estado, creado_por
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, creado_en
	`,
		strings.ToLower(strings.TrimSpace(s.Codigo)),
		s.Nombre,
		nullableTexto(s.Descripcion),
		s.UrlAcceso,
		s.UrlLogin,
		nombreUsuario,
		nombrePassword,
		metodoLogin,
		s.Motor,
		s.ClaveAdaptador,
		s.RequiereLoginGlobal,
		s.SoportaLectura,
		s.SoportaAutoregistro,
		s.Estado,
		s.CreadoPor,
	).Scan(&s.Id, &s.CreadoEn)
}

func ConsultarSistemaPorIdDb(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID) (*SistemaDestino, error) {
	s := &SistemaDestino{}
	var descripcion *string
	err := ejecutor.QueryRow(contexto, `
		SELECT id, codigo, nombre, descripcion, url_acceso,
		       url_login, nombre_campo_usuario, nombre_campo_password, metodo_login,
		       motor, clave_adaptador,
		       requiere_login_global, soporta_lectura, soporta_autoregistro,
		       estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM sistema_destino
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, id).Scan(
		&s.Id, &s.Codigo, &s.Nombre, &descripcion, &s.UrlAcceso,
		&s.UrlLogin, &s.NombreCampoUsuario, &s.NombreCampoPassword, &s.MetodoLogin,
		&s.Motor, &s.ClaveAdaptador,
		&s.RequiereLoginGlobal, &s.SoportaLectura, &s.SoportaAutoregistro,
		&s.Estado, &s.CreadoEn, &s.CreadoPor, &s.ActualizadoEn, &s.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSistemaNoEncontrado
	}
	if descripcion != nil {
		s.Descripcion = *descripcion
	}
	return s, err
}

func ListarSistemasActivosDb(contexto context.Context, ejecutor cockroach.EjecutorSql) ([]SistemaDestino, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT id, codigo, nombre, coalesce(descripcion,''), url_acceso,
		       url_login, nombre_campo_usuario, nombre_campo_password, metodo_login,
		       motor, clave_adaptador,
		       requiere_login_global, soporta_lectura, soporta_autoregistro,
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
			&s.Id, &s.Codigo, &s.Nombre, &s.Descripcion, &s.UrlAcceso,
			&s.UrlLogin, &s.NombreCampoUsuario, &s.NombreCampoPassword, &s.MetodoLogin,
			&s.Motor, &s.ClaveAdaptador,
			&s.RequiereLoginGlobal, &s.SoportaLectura, &s.SoportaAutoregistro,
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
		SET nombre = $2, descripcion = $3, url_acceso = $4, motor = $5,
		    clave_adaptador = $6, requiere_login_global = $7,
		    soporta_lectura = $8, soporta_autoregistro = $9, estado = $10,
		    actualizado_en = now(), actualizado_por = $11
		WHERE id = $1 AND estado != 'ELIMINADO'
	`,
		s.Id, s.Nombre, nullableTexto(s.Descripcion), s.UrlAcceso, s.Motor,
		s.ClaveAdaptador, s.RequiereLoginGlobal,
		s.SoportaLectura, s.SoportaAutoregistro, s.Estado,
		actualizadoPor,
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

func InsertarConexionLectura(contexto context.Context, ejecutor cockroach.EjecutorSql, c *ConexionLectura) error {
	return ejecutor.QueryRow(contexto, `
		INSERT INTO conexion_lectura (
			sistema_destino_id, host, puerto, base_datos,
			usuario_db_cifrado, password_db_cifrada, ssl_modo,
			parametros_extra, estado, creado_por
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, creado_en
	`,
		c.SistemaDestinoId, c.Host, c.Puerto, c.BaseDatos,
		c.UsuarioDbCifrado, c.PasswordDbCifrada, c.SslModo,
		c.ParametrosExtra, c.Estado, c.CreadoPor,
	).Scan(&c.Id, &c.CreadoEn)
}

func ConsultarConexionLecturaActiva(contexto context.Context, ejecutor cockroach.EjecutorSql, sistemaDestinoId uuid.UUID) (*ConexionLectura, error) {
	c := &ConexionLectura{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, sistema_destino_id, host, puerto, base_datos,
		       usuario_db_cifrado, password_db_cifrada, ssl_modo,
		       coalesce(parametros_extra, '{}'::JSONB), estado, creado_en, creado_por
		FROM conexion_lectura
		WHERE sistema_destino_id = $1 AND estado = 'ACTIVO'
		ORDER BY creado_en DESC
		LIMIT 1
	`, sistemaDestinoId).Scan(
		&c.Id, &c.SistemaDestinoId, &c.Host, &c.Puerto, &c.BaseDatos,
		&c.UsuarioDbCifrado, &c.PasswordDbCifrada, &c.SslModo,
		&c.ParametrosExtra, &c.Estado, &c.CreadoEn, &c.CreadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConexionLecturaNoEncontrada
	}
	return c, err
}

func InsertarParametroHash(contexto context.Context, ejecutor cockroach.EjecutorSql, p *ParametroHashDestino) error {
	if p.EsDefault {
		if _, err := ejecutor.Exec(contexto, `
			UPDATE parametro_hash_destino SET es_default = false WHERE sistema_destino_id = $1
		`, p.SistemaDestinoId); err != nil {
			return err
		}
	}
	return ejecutor.QueryRow(contexto, `
		INSERT INTO parametro_hash_destino (
			sistema_destino_id, algoritmo, costo, salt_estrategia,
			es_default, notas, creado_por
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, creado_en
	`,
		p.SistemaDestinoId, p.Algoritmo, p.Costo, p.SalEstrategia,
		p.EsDefault, nullableTexto(p.Notas), p.CreadoPor,
	).Scan(&p.Id, &p.CreadoEn)
}

func ConsultarParametroHashDefault(contexto context.Context, ejecutor cockroach.EjecutorSql, sistemaDestinoId uuid.UUID) (*ParametroHashDestino, error) {
	p := &ParametroHashDestino{}
	var notas *string
	err := ejecutor.QueryRow(contexto, `
		SELECT id, sistema_destino_id, algoritmo, costo, salt_estrategia,
		       es_default, notas, creado_en, creado_por
		FROM parametro_hash_destino
		WHERE sistema_destino_id = $1 AND es_default = true
		LIMIT 1
	`, sistemaDestinoId).Scan(
		&p.Id, &p.SistemaDestinoId, &p.Algoritmo, &p.Costo, &p.SalEstrategia,
		&p.EsDefault, &notas, &p.CreadoEn, &p.CreadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("no hay parámetro de hash por defecto para el sistema")
	}
	if notas != nil {
		p.Notas = *notas
	}
	return p, err
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
