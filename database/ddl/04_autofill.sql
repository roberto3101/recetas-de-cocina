-- =============================================================================
-- Migración para soportar autofill HTML POST automático
-- =============================================================================

USE sistemas_unificados;

ALTER TABLE sistema_destino
    ADD COLUMN IF NOT EXISTS url_login STRING NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS nombre_campo_usuario STRING NOT NULL DEFAULT 'usuario',
    ADD COLUMN IF NOT EXISTS nombre_campo_password STRING NOT NULL DEFAULT 'password',
    ADD COLUMN IF NOT EXISTS metodo_login STRING NOT NULL DEFAULT 'POST'
        CHECK (metodo_login IN ('POST','GET'));

-- -----------------------------------------------------------------------------
-- Tabla: credencial_usuario_externo
-- Guarda la contraseña cifrada de cada usuario provisionado en un sistema
-- externo, para poder hacer autofill POST al login del sistema.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS credencial_usuario_externo (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    sistema_destino_id  UUID        NOT NULL REFERENCES sistema_destino (id) ON DELETE RESTRICT,
    id_externo_usuario  STRING      NOT NULL,
    correo_usuario      STRING,
    password_cifrada    BYTES       NOT NULL,
    estado              STRING      NOT NULL DEFAULT 'ACTIVO'
                        CHECK (estado IN ('ACTIVO','REVOCADO','ELIMINADO')),
    creado_en           TIMESTAMPTZ NOT NULL DEFAULT now(),
    creado_por          UUID        REFERENCES usuario (id) ON DELETE RESTRICT,
    actualizado_en      TIMESTAMPTZ,
    actualizado_por     UUID        REFERENCES usuario (id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_credencial_externo_unico
    ON credencial_usuario_externo (sistema_destino_id, id_externo_usuario)
    WHERE estado = 'ACTIVO';

CREATE INDEX IF NOT EXISTS idx_credencial_externo_sistema
    ON credencial_usuario_externo (sistema_destino_id)
    WHERE estado = 'ACTIVO';
