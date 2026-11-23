-- ============================================================
-- Sistema de Pesaje - Esquema de Base de Datos SQLite
-- ------------------------------------------------------------
-- Versionado como recurso (pesaje/src/main/resources/db/schema.sql).
-- Los nombres de tablas y columnas son EXACTAMENTE los del sistema
-- original (lazarus-pesaje/src/database/schema.sql): el escritorio es
-- cliente de la web y ambos extremos deben compartir el mismo modelo.
--
-- El archivo se aplica sentencia por sentencia al abrir la base; la
-- version del esquema se lleva en PRAGMA user_version (ver Database).
-- Las columnas que se agreguen despues deben entrar como migracion
-- idempotente (ALTER TABLE ... ADD COLUMN) y no solo como CREATE TABLE,
-- para que las bases ya instaladas no pierdan datos.
--
-- El seed de datos (usuario administrador) lo hace Database con una
-- contrasena generada en el primer arranque.
-- ============================================================

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS personas (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    nombre              TEXT NOT NULL,
    apellido_paterno    TEXT,
    apellido_materno    TEXT,
    ci                  TEXT,
    telefono            TEXT,
    correo              TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS usuarios (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    persona_id          INTEGER NOT NULL UNIQUE REFERENCES personas(id) ON DELETE CASCADE,
    email               TEXT NOT NULL UNIQUE,
    password_hash       TEXT,
    rol                 TEXT NOT NULL DEFAULT 'usuario',
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    ultimo_login        INTEGER,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_usuarios_persona ON usuarios(persona_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_usuarios_email ON usuarios(email);

CREATE TABLE IF NOT EXISTS empresas (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    nombre_empresa      TEXT NOT NULL,
    actividad_economica TEXT,
    correo_electronico  TEXT,
    telefono            TEXT,
    logo                TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS proveedores (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    persona_id          INTEGER NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    nombre_empresa      TEXT,
    descripcion         TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    version             INTEGER NOT NULL DEFAULT 0,
    pendiente_push      INTEGER NOT NULL DEFAULT 0,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS choferes (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    persona_id          INTEGER NOT NULL UNIQUE REFERENCES personas(id) ON DELETE CASCADE,
    licencia            TEXT,
    telefono            TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    version             INTEGER NOT NULL DEFAULT 0,
    pendiente_push      INTEGER NOT NULL DEFAULT 0,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_choferes_persona ON choferes(persona_id);

CREATE TABLE IF NOT EXISTS vehiculos (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    placa               TEXT NOT NULL,
    tipo_vehiculo       TEXT,
    tara                INTEGER NOT NULL DEFAULT 0,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    version              INTEGER NOT NULL DEFAULT 0,
    pendiente_push       INTEGER NOT NULL DEFAULT 0,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS vehiculo_chofer (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    vehiculo_id         INTEGER REFERENCES vehiculos(id) ON DELETE CASCADE,
    chofer_id           INTEGER NOT NULL REFERENCES choferes(id) ON DELETE CASCADE,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS bodegas (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    nombre              TEXT NOT NULL,
    descripcion         TEXT,
    ubicacion           TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS productos (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    nombre              TEXT NOT NULL,
    descripcion         TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    version              INTEGER NOT NULL DEFAULT 0,
    pendiente_push       INTEGER NOT NULL DEFAULT 0,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS origenes (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    nombre              TEXT NOT NULL,
    descripcion         TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    version              INTEGER NOT NULL DEFAULT 0,
    pendiente_push       INTEGER NOT NULL DEFAULT 0,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS destinos (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    nombre              TEXT NOT NULL,
    descripcion         TEXT,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    version              INTEGER NOT NULL DEFAULT 0,
    pendiente_push       INTEGER NOT NULL DEFAULT 0,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS pesajes (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    guia                TEXT,
    lote                TEXT,
    vehiculo_id         INTEGER REFERENCES vehiculos(id) ON DELETE SET NULL,
    chofer_id           INTEGER REFERENCES choferes(id) ON DELETE SET NULL,
    proveedor_id        INTEGER REFERENCES proveedores(id) ON DELETE SET NULL,
    producto_id         INTEGER REFERENCES productos(id) ON DELETE SET NULL,
    id_origen           INTEGER REFERENCES origenes(id) ON DELETE SET NULL,
    id_destino          INTEGER REFERENCES destinos(id) ON DELETE SET NULL,
    peso_bruto          INTEGER,
    tara                INTEGER,
    peso_neto           INTEGER,
    costo_bs            INTEGER,
    flete_bs_pendiente  INTEGER,
    pesador_id          INTEGER REFERENCES personas(id) ON DELETE SET NULL,
    estado              TEXT NOT NULL DEFAULT 'ACTIVO',
    estado_balanza      TEXT NOT NULL DEFAULT 'EN_PROCESO',
    sincronizado        INTEGER NOT NULL DEFAULT 0,
    sync_event_id       TEXT,
    origen              TEXT NOT NULL DEFAULT 'ESCRITORIO',
    web_id              INTEGER,
    sync_error          TEXT,
    usuario_creacion    INTEGER,
    usuario_modificacion INTEGER,
    fecha_creacion      TEXT,
    fecha_modificacion  TEXT
);

CREATE INDEX IF NOT EXISTS idx_pesajes_fecha ON pesajes(fecha_creacion);
CREATE INDEX IF NOT EXISTS idx_pesajes_vehiculo ON pesajes(vehiculo_id);

CREATE TABLE IF NOT EXISTS boleta_config (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    salida              TEXT NOT NULL DEFAULT 'Salida a Potosi',
    direccion           TEXT NOT NULL DEFAULT 'Tarija km 2',
    celular1            TEXT NOT NULL DEFAULT '2782323',
    celular2            TEXT NOT NULL DEFAULT '1234343',
    ciudad              TEXT NOT NULL DEFAULT 'POTOSI - BOLIVIA',
    titulo_superior     TEXT NOT NULL DEFAULT 'BALANZA DE PESAJE DIGITAL',
    marca               TEXT NOT NULL DEFAULT 'PRIMAVERA',
    titulo_documento    TEXT NOT NULL DEFAULT 'BOLETA DE PESAJE',
    acreditacion        TEXT,
    usuario_modificacion INTEGER,
    fecha_modificacion  TEXT
);

CREATE TABLE IF NOT EXISTS config_balanza (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    puerto_com          TEXT NOT NULL DEFAULT 'COM1',
    baudrate            INTEGER NOT NULL DEFAULT 9600,
    databits            INTEGER NOT NULL DEFAULT 8,
    paridad             TEXT NOT NULL DEFAULT 'N',
    stopbits            INTEGER NOT NULL DEFAULT 1,
    flowcontrol         TEXT NOT NULL DEFAULT 'None',
    timeout_ms          INTEGER NOT NULL DEFAULT 1000,
    metodo_lectura      TEXT NOT NULL DEFAULT 'AUTO',
    posicion_inicio     INTEGER NOT NULL DEFAULT 8,
    posicion_longitud   INTEGER NOT NULL DEFAULT 5,
    usuario_modificacion INTEGER,
    fecha_modificacion  TEXT
);

-- Cola de salida hacia la web. client_event_id es la clave de
-- idempotencia que usa el servidor: reenviar el mismo evento nunca
-- duplica el pesaje.
CREATE TABLE IF NOT EXISTS sync_outbox (
    id              INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    client_event_id TEXT NOT NULL UNIQUE,
    recurso         TEXT NOT NULL,
    entidad_id      INTEGER,
    payload         TEXT,
    estado          TEXT NOT NULL DEFAULT 'PENDIENTE',
    intentos        INTEGER NOT NULL DEFAULT 0,
    ultimo_error    TEXT,
    proximo_intento TEXT,
    web_id          INTEGER,
    creado_en       TEXT NOT NULL,
    actualizado_en  TEXT
);

CREATE INDEX IF NOT EXISTS idx_sync_outbox_pendientes ON sync_outbox(estado, proximo_intento);

-- Cursor de pull: hasta donde bajamos cada recurso del servidor.
CREATE TABLE IF NOT EXISTS sync_cursor (
    recurso     TEXT PRIMARY KEY NOT NULL,
    cursor      TEXT,
    ultima_sync TEXT
);

-- Conflictos de LWW: el servidor tenia una version mas nueva.
CREATE TABLE IF NOT EXISTS sync_conflictos (
    id            INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    recurso       TEXT NOT NULL,
    entidad_id    INTEGER,
    cliente_json  TEXT,
    servidor_json TEXT,
    detectado_en  TEXT NOT NULL
);

