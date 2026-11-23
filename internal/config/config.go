package config

import (
	"os"
	"strconv"
)

const EmailAdministrador = "admin@sistema.com"

type Config struct {
	ListenAddr    string
	DatabasePath  string
	JWTSecret     string
	JWTRefresh    string
	StaticWebDir  string
}

func Load() Config {
	return Config{
		ListenAddr:   envOr("LISTEN_ADDR", ":8080"),
		DatabasePath: envOr("DATABASE_PATH", "./data/servidor.db"),
		JWTSecret:    envOr("JWT_SECRET", "cambiar-jwt-secreto-en-produccion"),
		JWTRefresh:   envOr("JWT_REFRESH_SECRET", "cambiar-refresh-secreto-en-produccion"),
		StaticWebDir: envOr("WEB_DIST", "./web/dist"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func EnvBool(key string) bool {
	v := os.Getenv(key)
	b, _ := strconv.ParseBool(v)
	return b
}
