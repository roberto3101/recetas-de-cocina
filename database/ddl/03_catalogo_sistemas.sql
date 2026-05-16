-- =============================================================================
-- Dominio: Catalogo de Sistemas
-- Sistema: sistemas_unificados
-- Motor:   CockroachDB
--
-- Registro de los sistemas externos que el gestor conoce. Cada sistema declara:
--   - su URL de login
--   - el motor y conexión de lectura para consumir sus usuarios
--   - el adaptador en código que sabe cómo provisionar nuevos usuarios
--   - el algoritmo de hash que usa para password en su tabla destino
-- =============================================================================

USE sistemas_unificados;

-- -----------------------------------------------------------------------------
-- Tabla: sistema_destino
-- Raíz de Agregado. Catálogo de sistemas externos integrados.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sistema_destino (
    id                      UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    codigo                  STRING      NOT NULL,
    nombre                  STRING      NOT NULL,
    descripcion             STRING,
    url_acceso              STRING      NOT NULL,
    motor                   STRING      NOT NULL
                            CHECK (motor IN ('COCKROACHDB','POSTGRESQL','MYSQL','MARIADB','API_REST','OTRO')),
    clave_adaptador         STRING      NOT NULL,
    requiere_login_global   BOOL        NOT NULL DEFAULT false,
    soporta_lectura         BOOL        NOT NULL DEFAULT true,
    soporta_autoregistro    BOOL        NOT NULL DEFAULT false,
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_sistema_destino_adaptador
    ON sistema_destino (clave_adaptador)
    WHERE estado != 'ELIMINADO';

CREATE INDEX IF NOT EXISTS idx_sistema_destino_estado
    ON sistema_destino (estado);

-- -----------------------------------------------------------------------------
-- Tabla: conexion_lectura
-- Conexión de la bóveda hacia la BD del sistema externo (solo lectura).
-- Las credenciales del gestor para esa BD se persisten cifradas.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS conexion_lectura (
    id                      UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    sistema_destino_id      UUID        NOT NULL REFERENCES sistema_destino (id) ON DELETE RESTRICT,
    host                    STRING      NOT NULL,
    puerto                  INT8        NOT NULL,
    base_datos              STRING      NOT NULL,
    usuario_db_cifrado      BYTES       NOT NULL,
    password_db_cifrada     BYTES       NOT NULL,
    ssl_modo                STRING      NOT NULL DEFAULT 'require'
                            CHECK (ssl_modo IN ('disable','allow','prefer','require','verify-ca','verify-full')),
    parametros_extra        JSONB,
    estado                  STRING      NOT NULL DEFAULT 'ACTIVO'
                            CHECK (estado IN ('ACTIVO','INACTIVO','ELIMINADO')),
    creado_en               TIMESTAMPTZ NOT NULL DEFAULT now(),
    creado_por              UUID        REFERENCES usuario (id) ON DELETE RESTRICT,
    actualizado_en          TIMESTAMPTZ,
    actualizado_por         UUID        REFERENCES usuario (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_conexion_lectura_sistema
    ON conexion_lectura (sistema_destino_id)
    WHERE estado = 'ACTIVO';

-- -----------------------------------------------------------------------------
-- Tabla: parametro_hash_destino
-- Algoritmo de hash que cada sistema destino espera en su tabla de usuarios.
-- Separado de sistema_destino para permitir múltiples convenciones por sistema
-- (ej: usuarios viejos con bcrypt y nuevos con argon2id).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS parametro_hash_destino (
    id                      UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    sistema_destino_id      UUID        NOT NULL REFERENCES sistema_destino (id) ON DELETE RESTRICT,
    algoritmo               STRING      NOT NULL
                            CHECK (algoritmo IN ('BCRYPT','ARGON2ID','SHA256_HEX','SHA512_HEX','PBKDF2','PLAINTEXT','OTRO')),
    costo                   INT8,
    salt_estrategia         STRING      NOT NULL DEFAULT 'POR_FILA'
                            CHECK (salt_estrategia IN ('POR_FILA','GLOBAL','SIN_SALT')),
    es_default              BOOL        NOT NULL DEFAULT true,
    notas                   STRING,
    creado_en               TIMESTAMPTZ NOT NULL DEFAULT now(),
    creado_por              UUID        REFERENCES usuario (id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_parametro_hash_default
    ON parametro_hash_destino (sistema_destino_id)
    WHERE es_default = true;
