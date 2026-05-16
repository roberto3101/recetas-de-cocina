-- =============================================================================
-- Dominio: Identidad
-- Sistema: sistemas_unificados (gestor de credenciales descentralizado)
-- Motor:   CockroachDB
-- =============================================================================

USE sistemas_unificados;

-- -----------------------------------------------------------------------------
-- Tabla: usuario
-- Raíz de Agregado. Identidad permanente del operador del gestor.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS usuario (
    id                            UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    correo_electronico            STRING      NOT NULL,
    password_hash                 STRING      NOT NULL,
    correo_electronico_verificado BOOL        NOT NULL DEFAULT false,
    estado                        STRING      NOT NULL DEFAULT 'PENDIENTE'
                                  CHECK (estado IN ('PENDIENTE', 'ACTIVO', 'INACTIVO', 'BLOQUEADO', 'ELIMINADO')),
    intentos_fallidos             INT8        NOT NULL DEFAULT 0
                                  CHECK (intentos_fallidos >= 0),
    bloqueado_hasta               TIMESTAMPTZ,
    ultimo_inicio_sesion_en       TIMESTAMPTZ,
    creado_en                     TIMESTAMPTZ NOT NULL DEFAULT now(),
    creado_por                    UUID,
    actualizado_en                TIMESTAMPTZ,
    actualizado_por               UUID
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_usuario_correo_electronico
    ON usuario (lower(correo_electronico));

CREATE INDEX IF NOT EXISTS idx_usuario_bloqueado
    ON usuario (bloqueado_hasta)
    WHERE estado = 'BLOQUEADO';

-- -----------------------------------------------------------------------------
-- Tabla: usuario_totp
-- Secret TOTP por operador para segundo factor (Google Authenticator/Authy).
-- Un usuario tiene a lo sumo un secret ACTIVO.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS usuario_totp (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    usuario_id          UUID        NOT NULL REFERENCES usuario (id) ON DELETE RESTRICT,
    secreto_cifrado     STRING      NOT NULL,
    codigos_respaldo    STRING[]    NOT NULL DEFAULT ARRAY[]::STRING[],
    estado              STRING      NOT NULL DEFAULT 'PENDIENTE'
                        CHECK (estado IN ('PENDIENTE', 'ACTIVO', 'REVOCADO')),
    activado_en         TIMESTAMPTZ,
    revocado_en         TIMESTAMPTZ,
    creado_en           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_usuario_totp_activo
    ON usuario_totp (usuario_id)
    WHERE estado = 'ACTIVO';

-- -----------------------------------------------------------------------------
-- Tabla: sesion_global
-- Raíz de Agregado. Manifestación operativa del usuario.
-- Multi-sesión por dispositivo permitido (cada pestaña una sesión).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sesion_global (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    usuario_id          UUID        NOT NULL REFERENCES usuario (id) ON DELETE RESTRICT,
    token_hash          STRING      NOT NULL,
    refresh_token_hash  STRING      NOT NULL,
    dispositivo_id      STRING      NOT NULL,
    ip_origen           STRING      NOT NULL,
    agente_usuario      STRING,
    segundo_factor_validado BOOL    NOT NULL DEFAULT false,
    emitida_en          TIMESTAMPTZ NOT NULL DEFAULT now(),
    expira_en           TIMESTAMPTZ NOT NULL,
    estado              STRING      NOT NULL DEFAULT 'ACTIVA'
                        CHECK (estado IN ('ACTIVA','EXPIRADA','REVOCADA','INVALIDADA','PENDIENTE_SEGUNDO_FACTOR')),
    ultimo_acceso_en    TIMESTAMPTZ,
    revocado_en         TIMESTAMPTZ,
    revocado_por        UUID        REFERENCES usuario (id) ON DELETE SET NULL,
    motivo_revocacion   STRING,
    CHECK (expira_en > emitida_en)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sesion_token
    ON sesion_global (token_hash);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sesion_refresh
    ON sesion_global (refresh_token_hash);

CREATE INDEX IF NOT EXISTS idx_sesion_usuario_activa
    ON sesion_global (usuario_id)
    WHERE estado = 'ACTIVA';

CREATE INDEX IF NOT EXISTS idx_sesion_expiracion
    ON sesion_global (expira_en)
    WHERE estado = 'ACTIVA';

-- -----------------------------------------------------------------------------
-- Tabla: auditoria_accion
-- Dominio Transversal. Registro inmutable de toda acción del sistema.
-- Todos los dominios escriben aquí. Nadie modifica filas existentes.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS auditoria_accion (
    id               UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    usuario_id       UUID        REFERENCES usuario (id) ON DELETE RESTRICT,
    sesion_id        UUID        REFERENCES sesion_global (id) ON DELETE SET NULL,
    modulo           STRING      NOT NULL,
    accion           STRING      NOT NULL,
    entidad          STRING,
    entidad_id       UUID,
    datos_anteriores JSONB,
    datos_nuevos     JSONB,
    detalle          STRING,
    ip_origen        STRING,
    agente_usuario   STRING,
    creado_en        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auditoria_usuario
    ON auditoria_accion (usuario_id);

CREATE INDEX IF NOT EXISTS idx_auditoria_sesion
    ON auditoria_accion (sesion_id)
    WHERE sesion_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_auditoria_entidad
    ON auditoria_accion (entidad, entidad_id);

CREATE INDEX IF NOT EXISTS idx_auditoria_fecha
    ON auditoria_accion (creado_en DESC);

CREATE INDEX IF NOT EXISTS idx_auditoria_modulo_accion
    ON auditoria_accion (modulo, accion);
