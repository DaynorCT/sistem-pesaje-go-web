// Uso: go run ./cmd/set-admin-pw [contraseña]  (default admin123)
package main

import (
	"database/sql"
	"fmt"
	"os"

	"sistema-pesaje-web/internal/clave"
	"sistema-pesaje-web/internal/config"

	_ "modernc.org/sqlite"
)

func main() {
	pwd := "admin123"
	if len(os.Args) > 1 {
		pwd = os.Args[1]
	}
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "./data/servidor.db"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	r, err := db.Exec(
		`UPDATE usuarios SET password_hash = ? WHERE email = ?`,
		clave.Hash(pwd), config.EmailAdministrador,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	n, _ := r.RowsAffected()
	fmt.Printf("Contraseña de %s actualizada (%d fila).\n", config.EmailAdministrador, n)
}