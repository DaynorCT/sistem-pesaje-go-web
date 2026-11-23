CREATE TABLE IF NOT EXISTS sync_seq_contador (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    valor INTEGER NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO sync_seq_contador (id, valor) VALUES (1, 0);

CREATE TABLE IF NOT EXISTS peso_en_vivo (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    peso REAL NOT NULL DEFAULT 0,
    fecha TEXT
);

CREATE TABLE IF NOT EXISTS sync_ops (
    client_event_id TEXT PRIMARY KEY,
    recurso TEXT NOT NULL,
    entidad_id INTEGER,
    resultado TEXT,
    aplicado_en TEXT
);
