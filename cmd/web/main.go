package main

import (
	"log"
	"net/http"

	"sistema-pesaje-web/internal/config"
	"sistema-pesaje-web/internal/db"
	"sistema-pesaje-web/internal/server"
)

func main() {
	cfg := config.Load()
	sqlDB, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("base de datos: %v", err)
	}
	defer sqlDB.Close()

	srv := server.New(cfg, sqlDB)
	log.Printf("servidor en %s (db %s)", cfg.ListenAddr, cfg.DatabasePath)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
