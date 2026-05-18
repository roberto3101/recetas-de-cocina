-- =============================================================================
-- Dominio: Catalogo de Sistemas
-- Sistema: sistemas_unificados
-- Motor:   CockroachDB
--
-- Catálogo de sistemas externos a los que el operador necesita acceder.
-- Cada sistema declara su URL de login y los nombres de los campos del form,
-- para que el bookmarklet / autofill sepa qué inputs rellenar.
-- =============================================================================

USE sistemas_unificados;

-- -----------------------------------------------------------------------------
-- Tabla: sistema_destino
-- Catálogo de URLs/sistemas. Sin conexión a BD externa.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sistema_destino (
    id                      UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    codigo                  STRING      NOT NULL,
    nombre                  STRING      NOT NULL,
    url_acceso              STRING      NOT NULL,
    url_login               STRING      NOT NULL DEFAULT '',
    nombre_campo_usuario    STRING      NOT NULL DEFAULT 'correo_electronico',
    nombre_campo_password   STRING      NOT NULL DEFAULT 'password',
    metodo_login            STRING      NOT NULL DEFAULT 'POST'
                            CHECK (metodo_login IN ('POST','GET')),
    estado                  STRING      NOT NULL DEFAULT 'ACTIVO'
                            CHECK (estado IN ('ACTIVO','INACTIVO','ELIMINADO')),
    creado_en               TIMESTAMPTZ NOT NULL DEFAULT now(),
    creado_por              UUID        REFERENCES usuario (id) ON DELETE RESTRICT,
    actualizado_en          TIMESTAMPTZ,
    actualizado_por         UUID        REFERENCES usuario (id) ON DELETE RESTRICT,
    CHECK (length(codigo) >= 2)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sistema_destino_codigo
    ON sistema_destino (lower(codigo))
    WHERE estado != 'ELIMINADO';

CREATE INDEX IF NOT EXISTS idx_sistema_destino_estado
    ON sistema_destino (estado);
