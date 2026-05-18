-- =============================================================================
-- Reincorporación de campos tipo + puerto del mockup original
-- - tipo: WEB | ESCRITORIO | FTP | OTRO
-- - puerto: opcional (INT2 nullable, 1..65535)
-- =============================================================================

USE sistemas_unificados;

ALTER TABLE acceso_guardado
    ADD COLUMN IF NOT EXISTS tipo STRING NOT NULL DEFAULT 'WEB'
        CHECK (tipo IN ('WEB','ESCRITORIO','FTP','OTRO'));

ALTER TABLE acceso_guardado
    ADD COLUMN IF NOT EXISTS puerto INT2
        CHECK (puerto IS NULL OR (puerto >= 1 AND puerto <= 65535));
