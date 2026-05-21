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
	return listarSistemasFiltrado(contexto, ejecutor, "ACTIVO")
}

// ListarSistemasArchivadosDb devuelve los sistemas con estado ELIMINADO
// (lo que el frontend llama "archivados"). El nombre es por consistencia
// con el verbo del usuario, aunque internamente el estado se llame ELIMINADO.
func ListarSistemasArchivadosDb(contexto context.Context, ejecutor cockroach.EjecutorSql) ([]SistemaDestino, error) {
	return listarSistemasFiltrado(contexto, ejecutor, "ELIMINADO")
}

// listarSistemasFiltrado: helper interno que centraliza la query y solo
// varía el filtro de estado. Mantiene una única fuente de verdad para el
// ORDER BY y la lista de columnas, evitando que se desincronicen.
func listarSistemasFiltrado(contexto context.Context, ejecutor cockroach.EjecutorSql, estado string) ([]SistemaDestino, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT id, codigo, nombre, url_acceso,
		       url_login, nombre_campo_usuario, nombre_campo_password, metodo_login,
		       estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM sistema_destino
		WHERE estado = $1
		ORDER BY nombre ASC
	`, estado)
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

// ReactivarSistemaLogico es el inverso simétrico de EliminarSistemaLogico:
// pone el sistema en estado ACTIVO y revive en cascada todos sus accesos
// que estaban REVOCADO. Si el operador quiere mantener algún acceso
// inactivo después, puede desactivarlo individualmente.
//
// Devuelve cuántos accesos volvieron a ACTIVO.
func ReactivarSistemaLogico(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, reactivadoPor uuid.UUID) (int, error) {
	// 1) Volver el sistema a ACTIVO. Solo si estaba ELIMINADO — no tocamos
	//    sistemas ya activos (sería operación no idempotente confusa).
	tagSistema, err := ejecutor.Exec(contexto, `
		UPDATE sistema_destino
		SET estado = 'ACTIVO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado = 'ELIMINADO'
	`, id, reactivadoPor)
	if err != nil {
		return 0, err
	}
	if tagSistema.RowsAffected() == 0 {
		return 0, ErrSistemaNoEncontrado
	}

	// 2) Cascada: volver REVOCADO → ACTIVO para todos los accesos del sistema.
	tagAccesos, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET estado = 'ACTIVO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE sistema_destino_id = $1 AND estado = 'REVOCADO'
	`, id, reactivadoPor)
	if err != nil {
		return 0, err
	}
	return int(tagAccesos.RowsAffected()), nil
}

// ContarAccesosRevocadosDeSistema devuelve cuántos accesos REVOCADOS hay
// vinculados al sistema. El frontend lo usa para mostrar el número en el
// modal de reactivación.
func ContarAccesosRevocadosDeSistema(contexto context.Context, ejecutor cockroach.EjecutorSql, sistemaId uuid.UUID) (int, error) {
	var total int
	err := ejecutor.QueryRow(contexto, `
		SELECT count(*) FROM acceso_guardado
		WHERE sistema_destino_id = $1 AND estado = 'REVOCADO'
	`, sistemaId).Scan(&total)
	return total, err
}

// ContarAccesosActivosDeSistema devuelve cuántos accesos NO ELIMINADOS
// están apuntando a este sistema. Lo usamos para mostrar en el modal de
// confirmación cuántos se van a desactivar en cascada.
func ContarAccesosActivosDeSistema(contexto context.Context, ejecutor cockroach.EjecutorSql, sistemaId uuid.UUID) (int, error) {
	var total int
	err := ejecutor.QueryRow(contexto, `
		SELECT count(*) FROM acceso_guardado
		WHERE sistema_destino_id = $1 AND estado != 'ELIMINADO'
	`, sistemaId).Scan(&total)
	return total, err
}

// EliminarSistemaLogico marca el sistema como ELIMINADO (soft delete) y
// hace cascada sobre sus accesos: todos los ACTIVOS pasan a REVOCADO.
//
// Por qué cascada en lugar de bloquear:
//   - Nada se borra físicamente — los accesos quedan recuperables si el
//     operador reactiva el sistema (cambia su estado de vuelta a ACTIVO)
//   - El FK ON DELETE RESTRICT sigue garantizando que nadie pueda hacer
//     DELETE físico por SQL directo — solo soft delete autorizado por la app
//   - UX simple: una acción del operador, una intención clara, el sistema
//     hace el trabajo de mantener consistencia
//
// Devuelve el número de accesos que fueron afectados (REVOCADOS en cascada).
func EliminarSistemaLogico(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, eliminadoPor uuid.UUID) (int, error) {
	// 1) Revocar accesos vinculados que estén ACTIVOS.
	tag, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET estado = 'REVOCADO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE sistema_destino_id = $1 AND estado = 'ACTIVO'
	`, id, eliminadoPor)
	if err != nil {
		return 0, err
	}
	accesosAfectados := int(tag.RowsAffected())

	// 2) Marcar el sistema como ELIMINADO.
	if _, err := ejecutor.Exec(contexto, `
		UPDATE sistema_destino
		SET estado = 'ELIMINADO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, id, eliminadoPor); err != nil {
		return 0, err
	}
	return accesosAfectados, nil
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
