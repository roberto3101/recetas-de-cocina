package recuperacion_password

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sistemas-unificados/persistencia/cockroach"
)

// InsertarToken crea un registro con el hash del token, ignorando el plano
// (el plano solo vive en RAM hasta que termina la petición HTTP).
func InsertarToken(contexto context.Context, ejecutor cockroach.EjecutorSql, t *Token) error {
	return ejecutor.QueryRow(contexto, `
		INSERT INTO token_recuperacion (
			usuario_id, token_hash, ip_origen, agente_usuario, expira_en
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, creado_en
	`, t.UsuarioId, t.TokenHash, t.IpOrigen, t.AgenteUsuario, t.ExpiraEn).Scan(&t.Id, &t.CreadoEn)
}

// ConsultarPorHash retorna el token vigente o no — el caller decide qué hacer.
// El índice unique sobre token_hash garantiza que solo hay uno.
func ConsultarPorHash(contexto context.Context, ejecutor cockroach.EjecutorSql, hash string) (*Token, error) {
	t := &Token{}
	err := ejecutor.QueryRow(contexto, `
		SELECT id, usuario_id, token_hash, ip_origen, agente_usuario,
		       creado_en, expira_en, consumido_en
		FROM token_recuperacion
		WHERE token_hash = $1
	`, hash).Scan(
		&t.Id, &t.UsuarioId, &t.TokenHash, &t.IpOrigen, &t.AgenteUsuario,
		&t.CreadoEn, &t.ExpiraEn, &t.ConsumidoEn,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTokenInvalido
	}
	return t, err
}

// MarcarComoConsumido pone consumido_en=now() solo si aún no estaba consumido,
// y la operación falla (0 filas afectadas) en cualquier otro caso. Esto cierra
// la ventana de race condition: dos peticiones simultáneas con el mismo token
// no pueden ambas cambiar la password.
func MarcarComoConsumido(contexto context.Context, ejecutor cockroach.EjecutorSql, id uuid.UUID, ahora time.Time) error {
	etiqueta, err := ejecutor.Exec(contexto, `
		UPDATE token_recuperacion
		SET consumido_en = $2
		WHERE id = $1 AND consumido_en IS NULL AND expira_en > $2
	`, id, ahora)
	if err != nil {
		return err
	}
	if etiqueta.RowsAffected() == 0 {
		return ErrTokenInvalido
	}
	return nil
}

// InvalidarTokensVigentesDeUsuario marca como consumidos cualquier token
// activo previo del mismo usuario. Se llama al emitir uno nuevo, para que
// solo el más reciente sea válido. Evita que un atacante junte varios links.
func InvalidarTokensVigentesDeUsuario(contexto context.Context, ejecutor cockroach.EjecutorSql, usuarioId uuid.UUID, ahora time.Time) error {
	_, err := ejecutor.Exec(contexto, `
		UPDATE token_recuperacion
		SET consumido_en = $2
		WHERE usuario_id = $1 AND consumido_en IS NULL
	`, usuarioId, ahora)
	return err
}
