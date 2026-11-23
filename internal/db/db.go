package db

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sistema-pesaje-web/internal/clave"
	"sistema-pesaje-web/internal/config"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql migrations.sql
var scripts embed.FS

const alfabetoClave = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

// Open abre o crea la base del servidor web.
func Open(path string) (*sql.DB, error) {
	path = filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}
	existia := fileExists(path)
	if err := applyScripts(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateColumns(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	adminPwd, err := seedAdmin(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if !existia && adminPwd != "" {
		fmt.Fprintf(os.Stderr, "[servidor] Admin creado (%s). Contraseña inicial: %s\n", config.EmailAdministrador, adminPwd)
	}
	return db, nil
}

func applyScripts(db *sql.DB) error {
	for _, name := range []string{"schema.sql", "migrations.sql"} {
		raw, err := scripts.ReadFile(name)
		if err != nil {
			return err
		}
		for _, stmt := range splitStatements(string(raw)) {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}

var columnasSync = []struct{ tabla, columna, tipo string }{
	{"vehiculos", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"choferes", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"proveedores", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"productos", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"origenes", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"destinos", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"pesajes", "sync_seq", "INTEGER NOT NULL DEFAULT 0"},
	{"pesajes", "client_event_id", "TEXT"},
	{"pesajes", "origen", "TEXT"},
	{"pesajes", "terminal_id", "TEXT"},
}

func migrateColumns(db *sql.DB) error {
	for _, c := range columnasSync {
		if err := addColumnIfMissing(db, c.tabla, c.columna, c.tipo); err != nil {
			return err
		}
	}
	_, _ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_pesajes_client_event_id ON pesajes(client_event_id)`)
	return nil
}

func addColumnIfMissing(db *sql.DB, tabla, columna, tipo string) error {
	rows, err := db.Query(`PRAGMA table_info(` + tabla + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == columna {
			return nil
		}
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tabla, columna, tipo))
	return err
}

func seedAdmin(db *sql.DB) (string, error) {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM usuarios WHERE rol = ?`, "administrador").Scan(&n); err != nil {
		return "", err
	}
	if n > 0 {
		return "", nil
	}
	pwd := randomPassword(12)
	ahora := time.Now().Format("2006-01-02 15:04:05")
	res, err := db.Exec(
		`INSERT INTO personas (nombre, apellido_paterno, correo, estado, fecha_creacion) VALUES (?, ?, ?, ?, ?)`,
		"Administrador", "Sistema", config.EmailAdministrador, "ACTIVO", ahora,
	)
	if err != nil {
		return "", err
	}
	personaID, _ := res.LastInsertId()
	_, err = db.Exec(
		`INSERT INTO usuarios (persona_id, email, password_hash, rol, estado, fecha_creacion) VALUES (?, ?, ?, ?, ?, ?)`,
		personaID, config.EmailAdministrador, clave.Hash(pwd), "administrador", "ACTIVO", ahora,
	)
	return pwd, err
}

func randomPassword(n int) string {
	b := make([]byte, n)
	for i := range b {
		var rb [1]byte
		_, _ = rand.Read(rb[:])
		b[i] = alfabetoClave[int(rb[0])%len(alfabetoClave)]
	}
	return string(b)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func SecuenciaActual(db *sql.DB) (int64, error) {
	var v int64
	err := db.QueryRow(`SELECT valor FROM sync_seq_contador WHERE id = 1`).Scan(&v)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return v, err
}

func NextSeq(tx *sql.Tx) (int64, error) {
	_, _ = tx.Exec(`INSERT OR IGNORE INTO sync_seq_contador (id, valor) VALUES (1, 0)`)
	_, err := tx.Exec(`UPDATE sync_seq_contador SET valor = valor + 1 WHERE id = 1`)
	if err != nil {
		return 0, err
	}
	var v int64
	err = tx.QueryRow(`SELECT valor FROM sync_seq_contador WHERE id = 1`).Scan(&v)
	return v, err
}

func Ahora() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05")[:19]
}
