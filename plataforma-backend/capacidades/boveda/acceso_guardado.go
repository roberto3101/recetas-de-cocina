package boveda

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
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

// Opciones para listar accesos. Todos los campos son opcionales con defaults
// sanos. La validación de valores permitidos se hace acá (whitelist) para
// evitar SQL injection cuando construimos ORDER BY dinámico.
type OpcionesListarAccesos struct {
	// OrdenarPor: "creado_en" (default) o "titulo". Cualquier otra cosa se
	// fuerza a "creado_en" para no romper SQL.
	OrdenarPor string
	// Direccion: "asc" o "desc" (default desc).
	Direccion string
	// Limite máximo de filas por respuesta. 0 = sin paginar. Cap a 200.
	Limite int
	// Offset para paginación. Negativo se trata como 0.
	Offset int
	// FiltroEstado: "ACTIVO", "REVOCADO" o vacío (incluye ambos, sin ELIMINADO).
	FiltroEstado string
	// SistemaDestinoId: limita a un sistema específico. Si vacío o UUID
	// inválido, no filtra. La validación de UUID la hace el caller (handler).
	SistemaDestinoId string
	// Busqueda: texto libre, case-insensitive. Busca en titulo del acceso,
	// nombre del sistema y url_acceso del sistema. NO busca en
	// usuario_externo porque ese campo está cifrado (AES-GCM).
	// Caracteres % y _ se escapan para que no actúen como wildcards.
	Busqueda string
}

// escaparPatronLike escapa los wildcards de SQL LIKE (% y _) para que
// el texto del usuario se busque literalmente, no como patrón.
// Ejemplo: "100%" buscaría todo si no escapamos; con escape busca el
// substring literal "100%".
func escaparPatronLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

// columnaOrdenValida traduce el nombre lógico que viene del cliente al
// nombre real de la columna. Whitelist explícita: cualquier otra cosa
// devuelve la columna por defecto. Esto cierra la puerta a SQL injection
// vía concatenación.
func columnaOrdenValida(nombre string) string {
	switch nombre {
	case "titulo":
		return "titulo"
	case "creado_en":
		return "creado_en"
	default:
		return "creado_en"
	}
}

// direccionValida es ASC o DESC; cualquier otra cosa cae a DESC.
func direccionValida(d string) string {
	switch strings.ToUpper(d) {
	case "ASC":
		return "ASC"
	case "DESC":
		return "DESC"
	default:
		return "DESC"
	}
}

// ContarAccesos devuelve el total que coincide con los filtros (sin paginar).
// Necesario para que el frontend sepa cuántas páginas hay. Aplica los mismos
// filtros que ListarAccesos (estado, sistema, búsqueda de texto) para que
// el total refleje exactamente lo que verá el usuario.
func ContarAccesos(contexto context.Context, ejecutor cockroach.EjecutorSql, opts OpcionesListarAccesos) (int, error) {
	where, args := construirWherePaginado(opts)
	consulta := `
		SELECT count(*) FROM acceso_guardado a
		LEFT JOIN sistema_destino s ON s.id = a.sistema_destino_id
		` + where
	var total int
	err := ejecutor.QueryRow(contexto, consulta, args...).Scan(&total)
	return total, err
}

// construirWherePaginado arma la cláusula WHERE compartida entre ListarAccesos
// y ContarAccesos. Devuelve el fragmento WHERE (con sus placeholders ya
// numerados) y la lista de args en el mismo orden. Centralizado para que
// nunca se desincronicen los filtros entre count y query principal.
func construirWherePaginado(opts OpcionesListarAccesos) (string, []any) {
	condiciones := []string{}
	args := []any{}

	// Estado: ACTIVO/REVOCADO explícito; si no, excluye ELIMINADO.
	if opts.FiltroEstado == "ACTIVO" || opts.FiltroEstado == "REVOCADO" {
		args = append(args, opts.FiltroEstado)
		condiciones = append(condiciones, "a.estado = $"+strconv.Itoa(len(args)))
	} else {
		condiciones = append(condiciones, "a.estado != 'ELIMINADO'")
	}

	// Sistema específico (UUID ya validado por el handler).
	if opts.SistemaDestinoId != "" {
		args = append(args, opts.SistemaDestinoId)
		condiciones = append(condiciones, "a.sistema_destino_id = $"+strconv.Itoa(len(args)))
	}

	// Búsqueda de texto: matchea contra titulo del acceso, nombre del
	// sistema y url_acceso del sistema. Todo case-insensitive (ILIKE).
	if q := strings.TrimSpace(opts.Busqueda); q != "" {
		patron := "%" + escaparPatronLike(strings.ToLower(q)) + "%"
		args = append(args, patron)
		idx := strconv.Itoa(len(args))
		condiciones = append(condiciones,
			"(lower(a.titulo) ILIKE $"+idx+
				" OR lower(s.nombre) ILIKE $"+idx+
				" OR lower(s.url_acceso) ILIKE $"+idx+")")
	}

	return "WHERE " + strings.Join(condiciones, " AND "), args
}

// ListarAccesos devuelve los accesos paginados y ordenados según opts.
// Descifra usuario_externo y observaciones. NO descifra password.
//
// Importante: filas cuyo descifrado falla (KEK distinta) se omiten — pero
// se descuentan del total que se está devolviendo. El ContarAccesos cuenta
// la realidad de la BD, así que el "total" puede ser mayor que la suma de
// los items visibles si hay residuos con KEK vieja. Es el comportamiento
// menos sorprendente para el usuario final.
func ListarAccesos(contexto context.Context, ejecutor cockroach.EjecutorSql, claves *cripto.ClavesCifrado, opts OpcionesListarAccesos) ([]AccesoGuardado, error) {
	columna := columnaOrdenValida(opts.OrdenarPor)
	direccion := direccionValida(opts.Direccion)
	where, args := construirWherePaginado(opts)

	// Tie-breaker: si dos filas tienen el mismo valor en la columna primaria
	// (ej. dos accesos creados el mismo segundo), ordenamos por id para que
	// la paginación sea estable y no devuelva filas duplicadas entre páginas.
	limitClause := ""
	if opts.Limite > 0 {
		lim := opts.Limite
		if lim > 200 {
			lim = 200
		}
		off := opts.Offset
		if off < 0 {
			off = 0
		}
		limitClause = " LIMIT " + intToStr(lim) + " OFFSET " + intToStr(off)
	}

	// LEFT JOIN con sistema_destino para que la búsqueda de texto pueda
	// matchear contra nombre/url del sistema. Si el sistema no existe
	// (huérfano por eliminación) el LEFT JOIN deja s.* como NULL y los
	// matches en s.nombre/url no aplican — pero el acceso sigue listándose.
	consulta := `
		SELECT a.id, a.titulo, a.sistema_destino_id, a.usuario_externo,
		       coalesce(a.observaciones, ''::BYTES), a.tipo, a.puerto, a.estado, a.creado_en
		FROM acceso_guardado a
		LEFT JOIN sistema_destino s ON s.id = a.sistema_destino_id
		` + where + `
		ORDER BY a.` + columna + ` ` + direccion + `, a.id ASC` + limitClause

	filas, err := ejecutor.Query(contexto, consulta, args...)
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

// intToStr usa strconv para evitar import circular o sprintf overhead.
func intToStr(n int) string { return strconv.Itoa(n) }

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

// EliminarAccesoPermanente borra físicamente la fila de la BD. A diferencia
// de DesactivarAcceso (estado=REVOCADO, recuperable), aquí ya no queda nada
// en la tabla — ni el ciphertext de la password ni del usuario. Es el
// "delete" real para casos en que el usuario quiere asegurar que el dato
// desaparezca del sistema (ej: rotación de credenciales por compromiso).
//
// Devuelve ErrAccesoNoEncontrado si el id no existe. NO discrimina por
// estado: podés borrar desde ACTIVO, REVOCADO o el caso raro de ELIMINADO
// (sin efecto si ya no estaba).
func EliminarAccesoPermanente(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, eliminadoPor uuid.UUID) error {
	tag, err := ejecutor.Exec(contexto, `
		DELETE FROM acceso_guardado WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAccesoNoEncontrado
	}
	// eliminadoPor no se persiste en la fila (ya no existe), pero el caller
	// debe escribir el audit_log con este UUID. Lo aceptamos por consistencia
	// con DesactivarAcceso y porque el handler lo necesita.
	_ = eliminadoPor
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
