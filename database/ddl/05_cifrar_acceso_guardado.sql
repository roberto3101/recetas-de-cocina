-- =============================================================================
-- Cifrado de campos sensibles en acceso_guardado
-- Limpia y aplica nuevo schema con BYTES cifrados.
-- (CockroachDB no soporta ALTER COLUMN TYPE de STRING→BYTES directo; usamos
-- drop+add.)
-- =============================================================================

USE sistemas_unificados;

-- Vaciar tabla (si tiene datos viejos sin cifrar)
DELETE FROM acceso_guardado;

DROP INDEX IF EXISTS idx_acceso_guardado_unico CASCADE;

ALTER TABLE acceso_guardado DROP COLUMN IF EXISTS usuario_externo;
ALTER TABLE acceso_guardado DROP COLUMN IF EXISTS observaciones;

ALTER TABLE acceso_guardado ADD COLUMN usuario_externo BYTES NOT NULL DEFAULT '\x00'::BYTES;
ALTER TABLE acceso_guardado ALTER COLUMN usuario_externo DROP DEFAULT;
ALTER TABLE acceso_guardado ADD COLUMN observaciones BYTES;
ALTER TABLE acceso_guardado ADD COLUMN usuario_externo_hash BYTES NOT NULL DEFAULT '\x00'::BYTES;
ALTER TABLE acceso_guardado ALTER COLUMN usuario_externo_hash DROP DEFAULT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_acceso_guardado_unico
    ON acceso_guardado (sistema_destino_id, usuario_externo_hash)
    WHERE estado = 'ACTIVO';
