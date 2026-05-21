-- =============================================================================
-- Índices para búsqueda y ordenamiento a escala (10K+ accesos)
-- =============================================================================
-- El catálogo actual tiene índices en (sistema_destino_id) y la PK por id,
-- pero ORDER BY creado_en y ORDER BY titulo hacen full scan + sort en memoria.
-- A 10K filas se vuelve lento (~200-300ms en sort). Estos índices llevan
-- el ordenamiento a un seek-and-scan O(log n).
--
-- Notas de diseño:
--   - Incluímos id como segundo campo del índice → tie-breaker estable
--     para la paginación (ORDER BY creado_en DESC, id ASC).
--   - lower(titulo) porque la búsqueda en backend usa ILIKE
--     (case-insensitive). Si indexamos titulo a secas, ILIKE no usaría el
--     índice. Con lower() funcional, la query `WHERE lower(titulo) LIKE ...`
--     sí pega al índice.
--   - WHERE estado != 'ELIMINADO' como índice parcial: los registros
--     borrados lógicos no se listan nunca, no hace falta indexarlos.
-- =============================================================================

USE sistemas_unificados;

-- ORDER BY creado_en DESC (default del listado)
CREATE INDEX IF NOT EXISTS idx_acceso_guardado_creado_en
    ON acceso_guardado (creado_en DESC, id ASC)
    WHERE estado != 'ELIMINADO';

-- ORDER BY titulo + búsqueda ILIKE
CREATE INDEX IF NOT EXISTS idx_acceso_guardado_titulo_lower
    ON acceso_guardado (lower(titulo) ASC, id ASC)
    WHERE estado != 'ELIMINADO';

-- WHERE estado = 'ACTIVO' / 'REVOCADO' (el caso típico del filtro UI)
CREATE INDEX IF NOT EXISTS idx_acceso_guardado_estado_creado
    ON acceso_guardado (estado, creado_en DESC, id ASC);
