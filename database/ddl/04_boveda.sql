-- =============================================================================
-- Dominio: Boveda
-- Sistema: sistemas_unificados
-- Motor:   CockroachDB
--
-- Cada fila es un acceso guardado por el operador para un sistema_destino:
-- titulo, usuario que va al form externo, password cifrada AES-256-GCM.
-- =============================================================================

USE sistemas_unificados;

CREATE TABLE IF NOT EXISTS acceso_guardado (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    titulo              STRING      NOT NULL DEFAULT '',
    sistema_destino_id  UUID        NOT NULL REFERENCES sistema_destino (id) ON DELETE RESTRICT,
    usuario_externo     STRING      NOT NULL,
    password_cifrada    BYTES       NOT NULL,
    observaciones       STRING,
    estado              STRING      NOT NULL DEFAULT 'ACTIVO'
                        CHECK (estado IN ('ACTIVO','REVOCADO','ELIMINADO')),
    creado_en           TIMESTAMPTZ NOT NULL DEFAULT now(),
    creado_por          UUID        REFERENCES usuario (id) ON DELETE RESTRICT,
    actualizado_en      TIMESTAMPTZ,
    actualizado_por     UUID        REFERENCES usuario (id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_acceso_guardado_unico
    ON acceso_guardado (sistema_destino_id, usuario_externo)
    WHERE estado = 'ACTIVO';

CREATE INDEX IF NOT EXISTS idx_acceso_guardado_sistema
    ON acceso_guardado (sistema_destino_id)
    WHERE estado = 'ACTIVO';

CREATE INDEX IF NOT EXISTS idx_acceso_guardado_creador
    ON acceso_guardado (creado_por)
    WHERE estado = 'ACTIVO';
