-- =============================================================================
-- Dominio: Recuperación de Contraseña ("Receta Perdida")
-- Sistema: sistemas_unificados
-- Motor:   CockroachDB
--
-- Token de un solo uso con expiración corta (30 min). El backend nunca
-- almacena el token en claro: guarda sha256(token) y compara igual que como
-- se hace con los tokens de sesión. El link a enviar al usuario contiene el
-- token plano que solo existe en memoria al momento de generarlo + en el
-- log/correo. Después de eso es irrecuperable.
--
-- Endpoint disfrazado:
--   POST /buscar/receta-perdida        → solicitar
--   GET  /buscar/validar-receta/{cod}  → validar antes del form
--   POST /buscar/preparar-receta       → consumir + cambiar password
-- =============================================================================

USE sistemas_unificados;

CREATE TABLE IF NOT EXISTS token_recuperacion (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    usuario_id      UUID        NOT NULL REFERENCES usuario (id) ON DELETE RESTRICT,
    token_hash      STRING      NOT NULL,
    ip_origen       STRING,
    agente_usuario  STRING,
    creado_en       TIMESTAMPTZ NOT NULL DEFAULT now(),
    expira_en       TIMESTAMPTZ NOT NULL,
    consumido_en    TIMESTAMPTZ,
    CHECK (expira_en > creado_en)
);

-- Lookup principal: por hash, solo tokens vigentes.
CREATE UNIQUE INDEX IF NOT EXISTS idx_token_recuperacion_hash
    ON token_recuperacion (token_hash);

-- Para limpiar / invalidar tokens previos al solicitar uno nuevo.
CREATE INDEX IF NOT EXISTS idx_token_recuperacion_usuario_vigente
    ON token_recuperacion (usuario_id)
    WHERE consumido_en IS NULL;

-- Para job futuro que purgue tokens expirados.
CREATE INDEX IF NOT EXISTS idx_token_recuperacion_expiracion
    ON token_recuperacion (expira_en)
    WHERE consumido_en IS NULL;
