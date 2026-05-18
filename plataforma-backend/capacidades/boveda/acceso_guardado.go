package boveda

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

var ErrAccesoNoEncontrado = errors.New("acceso guardado no encontrado")
var ErrUsuarioRequerido = errors.New("usuario es obligatorio")
var ErrPasswordRequerida = errors.New("password es obligatoria")
var ErrSistemaRequerido = errors.New("sistema_destino_id es obligatorio")

// AccesoGuardado representa el acceso en memoria (campos en plano).
// En BD usuario_externo, observaciones y password viven cifrados con AES-256-GCM.
// usuario_externo_hash (HMAC-SHA256) se usa para el UNIQUE INDEX.
type AccesoGuardado struct {
	Id                uuid.UUID
	Titulo            string
	SistemaDestinoId  uuid.UUID
	UsuarioExterno    string // plano en memoria, cifrado en BD
	PasswordPlana     string // solo para entrada (al guardar/editar); vacío al leer
	Observaciones     string // plano en memoria, cifrado en BD
	Tipo              string // WEB | ESCRITORIO | FTP | OTRO
	Puerto            *int16 // opcional, 1..65535
	Estado            string
	CreadoEn          time.Time
	CreadoPor         *uuid.UUID
	ActualizadoEn     *time.Time
	ActualizadoPor    *uuid.UUID
}

const (
	TipoAccesoWeb        = "WEB"
	TipoAccesoEscritorio = "ESCRITORIO"
	TipoAccesoFtp        = "FTP"
	TipoAccesoOtro       = "OTRO"
)

func tipoValido(t string) bool {
	switch t {
	case TipoAccesoWeb, TipoAccesoEscritorio, TipoAccesoFtp, TipoAccesoOtro:
		return true
	}
	return false
}

func GuardarAcceso(contexto context.Context, ejecutor cockroach.EjecutorSql, claves *cripto.ClavesCifrado, a *AccesoGuardado) error {
	usuario := strings.TrimSpace(a.UsuarioExterno)
	if usuario == "" {
		return ErrUsuarioRequerido
	}
	if a.PasswordPlana == "" {
		return ErrPasswordRequerida
	}
	if a.SistemaDestinoId == uuid.Nil {
		return ErrSistemaRequerido
	}
	clave := claves.ClaveBoveda()
	usuarioCifrado, err := cripto.CifrarConAesGcm(clave, []byte(usuario))
	if err != nil {
		return err
	}
	usuarioHash := cripto.HmacSha256(clave, usuario)
	passwordCifrada, err := cripto.CifrarConAesGcm(clave, []byte(a.PasswordPlana))
	if err != nil {
		return err
	}
	var observacionesCifradas any = nil
	if obs := strings.TrimSpace(a.Observaciones); obs != "" {
		oc, err := cripto.CifrarConAesGcm(clave, []byte(obs))
		if err != nil {
			return err
		}
		observacionesCifradas = oc
	}
	titulo := strings.TrimSpace(a.Titulo)
	tipo := strings.TrimSpace(a.Tipo)
	if tipo == "" {
		tipo = TipoAccesoWeb
	}
	if !tipoValido(tipo) {
		return errors.New("tipo inválido (use WEB, ESCRITORIO, FTP u OTRO)")
	}
	var puerto any = nil
	if a.Puerto != nil {
		if *a.Puerto < 1 || *a.Puerto > 32767 {
			return errors.New("puerto fuera de rango (1..65535)")
		}
		puerto = *a.Puerto
	}
	return ejecutor.QueryRow(contexto, `
		INSERT INTO acceso_guardado (
			titulo, sistema_destino_id, usuario_externo, usuario_externo_hash,
			password_cifrada, observaciones, tipo, puerto, estado, creado_por
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ACTIVO', $9)
		ON CONFLICT (sistema_destino_id, usuario_externo_hash) WHERE estado = 'ACTIVO'
		DO UPDATE SET
			titulo = EXCLUDED.titulo,
			usuario_externo = EXCLUDED.usuario_externo,
			password_cifrada = EXCLUDED.password_cifrada,
			observaciones = EXCLUDED.observaciones,
			tipo = EXCLUDED.tipo,
			puerto = EXCLUDED.puerto,
			actualizado_en = now(),
			actualizado_por = EXCLUDED.creado_por
		RETURNING id, creado_en
	`,
		titulo, a.SistemaDestinoId, usuarioCifrado, usuarioHash,
		passwordCifrada, observacionesCifradas, tipo, puerto, a.CreadoPor,
	).Scan(&a.Id, &a.CreadoEn)
}

func ConsultarAccesoPorId(contexto context.Context, ejecutor cockroach.EjecutorSql, claves *cripto.ClavesCifrado, id uuid.UUID) (*AccesoGuardado, error) {
	a := &AccesoGuardado{}
	var usuarioCifrado, passwordCifrada []byte
	var observacionesCifradas []byte
	err := ejecutor.QueryRow(contexto, `
		SELECT id, titulo, sistema_destino_id, usuario_externo,
		       password_cifrada, observaciones, tipo, puerto,
		       estado, creado_en, creado_por, actualizado_en, actualizado_por
		FROM acceso_guardado
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, id).Scan(
		&a.Id, &a.Titulo, &a.SistemaDestinoId, &usuarioCifrado,
		&passwordCifrada, &observacionesCifradas, &a.Tipo, &a.Puerto,
		&a.Estado, &a.CreadoEn, &a.CreadoPor, &a.ActualizadoEn, &a.ActualizadoPor,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAccesoNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	clave := claves.ClaveBoveda()
	usuario, err := cripto.DescifrarConAesGcm(clave, usuarioCifrado)
	if err != nil {
		return nil, err
	}
	a.UsuarioExterno = string(usuario)
	passwordPlana, err := cripto.DescifrarConAesGcm(clave, passwordCifrada)
	if err != nil {
		return nil, err
	}
	a.PasswordPlana = string(passwordPlana)
	if len(observacionesCifradas) > 0 {
		obs, err := cripto.DescifrarConAesGcm(clave, observacionesCifradas)
		if err != nil {
			return nil, err
		}
		a.Observaciones = string(obs)
	}
	return a, nil
}

// ListarAccesos devuelve ACTIVO + REVOCADO. Descifra usuario_externo y observaciones.
// NO descifra password (no se devuelve en listado).
func ListarAccesos(contexto context.Context, ejecutor cockroach.EjecutorSql, claves *cripto.ClavesCifrado) ([]AccesoGuardado, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT id, titulo, sistema_destino_id, usuario_externo,
		       coalesce(observaciones, ''::BYTES), tipo, puerto, estado, creado_en
		FROM acceso_guardado
		WHERE estado != 'ELIMINADO'
		ORDER BY creado_en DESC
	`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	clave := claves.ClaveBoveda()
	resultado := make([]AccesoGuardado, 0)
	for filas.Next() {
		a := AccesoGuardado{}
		var usuarioCifrado, observacionesCifradas []byte
		if err := filas.Scan(
			&a.Id, &a.Titulo, &a.SistemaDestinoId, &usuarioCifrado,
			&observacionesCifradas, &a.Tipo, &a.Puerto, &a.Estado, &a.CreadoEn,
		); err != nil {
			return nil, err
		}
		// Tolerante a filas cifradas con un KEK distinto (residuos de tests/migraciones):
		// si no se puede descifrar, logueamos y la omitimos en lugar de tumbar todo el listado.
		usuario, err := cripto.DescifrarConAesGcm(clave, usuarioCifrado)
		if err != nil {
			slog.Warn("acceso_guardado.descifrado_omitido",
				"id", a.Id.String(),
				"detalle", "no se pudo descifrar usuario_externo; probablemente cifrado con otra KEK",
				"error", err.Error())
			continue
		}
		a.UsuarioExterno = string(usuario)
		if len(observacionesCifradas) > 0 {
			obs, err := cripto.DescifrarConAesGcm(clave, observacionesCifradas)
			if err != nil {
				slog.Warn("acceso_guardado.observaciones_no_descifradas",
					"id", a.Id.String(),
					"error", err.Error())
				a.Observaciones = ""
			} else {
				a.Observaciones = string(obs)
			}
		}
		resultado = append(resultado, a)
	}
	return resultado, filas.Err()
}

func ActualizarAcceso(contexto context.Context, ejecutor cockroach.EjecutorSql, claves *cripto.ClavesCifrado, a *AccesoGuardado, actualizadoPor uuid.UUID) error {
	usuario := strings.TrimSpace(a.UsuarioExterno)
	if usuario == "" {
		return ErrUsuarioRequerido
	}
	if a.SistemaDestinoId == uuid.Nil {
		return ErrSistemaRequerido
	}
	clave := claves.ClaveBoveda()
	usuarioCifrado, err := cripto.CifrarConAesGcm(clave, []byte(usuario))
	if err != nil {
		return err
	}
	usuarioHash := cripto.HmacSha256(clave, usuario)
	var observacionesCifradas any = nil
	if obs := strings.TrimSpace(a.Observaciones); obs != "" {
		oc, err := cripto.CifrarConAesGcm(clave, []byte(obs))
		if err != nil {
			return err
		}
		observacionesCifradas = oc
	}
	var passwordCifrada any = nil
	if a.PasswordPlana != "" {
		pc, err := cripto.CifrarConAesGcm(clave, []byte(a.PasswordPlana))
		if err != nil {
			return err
		}
		passwordCifrada = pc
	}
	tipo := strings.TrimSpace(a.Tipo)
	if tipo == "" {
		tipo = TipoAccesoWeb
	}
	if !tipoValido(tipo) {
		return errors.New("tipo inválido (use WEB, ESCRITORIO, FTP u OTRO)")
	}
	var puerto any = nil
	if a.Puerto != nil {
		if *a.Puerto < 1 || *a.Puerto > 32767 {
			return errors.New("puerto fuera de rango (1..65535)")
		}
		puerto = *a.Puerto
	}
	tag, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET titulo = $2,
		    sistema_destino_id = $3,
		    usuario_externo = $4,
		    usuario_externo_hash = $5,
		    password_cifrada = CASE WHEN $6::BYTES IS NULL THEN password_cifrada ELSE $6 END,
		    observaciones = $7,
		    tipo = $8,
		    puerto = $9,
		    actualizado_en = now(),
		    actualizado_por = $10
		WHERE id = $1 AND estado != 'ELIMINADO'
	`,
		a.Id,
		strings.TrimSpace(a.Titulo),
		a.SistemaDestinoId,
		usuarioCifrado,
		usuarioHash,
		passwordCifrada,
		observacionesCifradas,
		tipo,
		puerto,
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

func DesactivarAcceso(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, actualizadoPor uuid.UUID) error {
	tag, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET estado = 'REVOCADO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado = 'ACTIVO'
	`, id, actualizadoPor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAccesoNoEncontrado
	}
	return nil
}

func ReactivarAcceso(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, actualizadoPor uuid.UUID) error {
	tag, err := ejecutor.Exec(contexto, `
		UPDATE acceso_guardado
		SET estado = 'ACTIVO',
		    actualizado_en = now(),
		    actualizado_por = $2
		WHERE id = $1 AND estado = 'REVOCADO'
	`, id, actualizadoPor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAccesoNoEncontrado
	}
	return nil
}
