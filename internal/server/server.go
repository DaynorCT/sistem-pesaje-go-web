package server

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sistema-pesaje-web/internal/api"
	"sistema-pesaje-web/internal/config"
)

type Server struct {
	cfg config.Config
	db  *sql.DB
	mux *http.ServeMux
}

func New(cfg config.Config, db *sql.DB) *Server {
	s := &Server{cfg: cfg, db: db, mux: http.NewServeMux()}
	secure := os.Getenv("NODE_ENV") == "production"

	authH := &api.AuthHandlers{DB: db, JWTSecret: cfg.JWTSecret, RefreshSecret: cfg.JWTRefresh, CookieSecure: secure}
	pesajes := &api.PesajesHandler{DB: db, JWTSecret: cfg.JWTSecret}
	ultimo := &api.UltimoPesoHandler{DB: db, JWTSecret: cfg.JWTSecret}
	sync := &api.SyncHandler{DB: db, JWTSecret: cfg.JWTSecret}

	s.mux.HandleFunc("/api/auth/token", authH.Token)
	s.mux.HandleFunc("/api/auth/login", authH.Login)
	s.mux.HandleFunc("/api/auth/refresh", authH.Refresh)
	s.mux.HandleFunc("/api/pesajes", pesajes.ServeHTTP)
	s.mux.HandleFunc("/api/pesajes/ultimo-peso", ultimo.ServeHTTP)
	s.mux.HandleFunc("/api/empresas/branding", sync.Branding)
	s.mux.HandleFunc("/api/sync/", sync.ServeHTTP)

	s.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if info, err := os.Stat(cfg.StaticWebDir); err == nil && info.IsDir() {
		s.mux.Handle("/", spaFileServer(cfg.StaticWebDir))
	}
	return s
}

func (s *Server) Handler() http.Handler {
	return withCORS(s.mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cookie")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// spaFileServer entrega la SPA de Vite y hace fallback a index.html.
func spaFileServer(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(root, filepath.Clean(r.URL.Path))
		if strings.HasPrefix(p, root) {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}
