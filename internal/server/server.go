package server

import (
	"database/sql"
	"net/http"
	"os"
	"path"
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
	inicio := &api.InicioHandler{DB: db, JWTSecret: cfg.JWTSecret}
	empresas := &api.EmpresasHandler{DB: db, JWTSecret: cfg.JWTSecret}
	usuarios := &api.UsuariosHandler{DB: db, JWTSecret: cfg.JWTSecret}
	choferes := &api.ChoferesHandler{DB: db, JWTSecret: cfg.JWTSecret}
	proveedores := &api.ProveedoresHandler{DB: db, JWTSecret: cfg.JWTSecret}
	vehiculos := &api.VehiculosHandler{DB: db, JWTSecret: cfg.JWTSecret}
	productos := &api.CatalogoHandler{DB: db, JWTSecret: cfg.JWTSecret, Tabla: "productos", Label: "Producto"}
	origenes := &api.CatalogoHandler{DB: db, JWTSecret: cfg.JWTSecret, Tabla: "origenes", Label: "Origen"}
	destinos := &api.CatalogoHandler{DB: db, JWTSecret: cfg.JWTSecret, Tabla: "destinos", Label: "Destino"}
	boletaCfg := &api.BoletaConfigHandler{DB: db, JWTSecret: cfg.JWTSecret}
	reportesH := &api.ReportesHandler{DB: db, JWTSecret: cfg.JWTSecret}

	s.mux.HandleFunc("/api/auth/token", authH.Token)
	s.mux.HandleFunc("/api/auth/login", authH.Login)
	s.mux.HandleFunc("/api/auth/refresh", authH.Refresh)
	s.mux.HandleFunc("GET /api/pesajes", pesajes.ListHTTP)
	s.mux.HandleFunc("POST /api/pesajes", pesajes.CreateHTTP)
	s.mux.HandleFunc("GET /api/pesajes/ultimo-peso", ultimo.ServeHTTP)
	s.mux.HandleFunc("GET /api/pesajes/opciones", pesajes.Opciones)
	s.mux.HandleFunc("GET /api/pesajes/grilla", pesajes.Grilla)
	s.mux.HandleFunc("GET /api/pesajes/lookup/vehiculo/{id}", pesajes.LookupVehiculo)
	s.mux.HandleFunc("GET /api/pesajes/lookup/chofer/{id}", pesajes.LookupChofer)
	s.mux.HandleFunc("GET /api/pesajes/{id}", pesajes.GetByID)
	s.mux.HandleFunc("PUT /api/pesajes/{id}", pesajes.UpdateByID)
	s.mux.HandleFunc("POST /api/pesajes/{id}/finalizar", pesajes.FinalizarByID)
	s.mux.HandleFunc("POST /api/pesajes/{id}/estado", pesajes.CambiarEstadoByID)
	s.mux.HandleFunc("GET /api/pesajes/{id}/boleta", pesajes.BoletaPDF)
	s.mux.HandleFunc("/api/config/boleta", boletaCfg.ServeHTTP)
	s.mux.HandleFunc("/api/empresas", empresas.ServeHTTP)
	s.mux.HandleFunc("/api/usuarios", usuarios.ServeHTTP)
	s.mux.HandleFunc("/api/choferes", choferes.ServeHTTP)
	s.mux.HandleFunc("/api/proveedores", proveedores.ServeHTTP)
	s.mux.HandleFunc("/api/vehiculos", vehiculos.ServeHTTP)
	s.mux.HandleFunc("/api/productos", productos.ServeHTTP)
	s.mux.HandleFunc("/api/origenes", origenes.ServeHTTP)
	s.mux.HandleFunc("/api/destinos", destinos.ServeHTTP)
	s.mux.HandleFunc("/api/empresas/branding", empresas.Branding)
	s.mux.HandleFunc("/api/sync/", sync.ServeHTTP)

	s.mux.HandleFunc("/api/reportes", reportesH.ServeHTTP)
	s.mux.HandleFunc("/api/reportes/pdf", reportesH.PDF)
	s.mux.HandleFunc("/api/inicio/resumen", inicio.Resumen)
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
	root = filepath.Clean(root)
	fs := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if rel == "" || rel == "." || strings.Contains(rel, "..") {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}
		p := filepath.Join(root, filepath.FromSlash(rel))
		if !strings.HasPrefix(filepath.Clean(p)+string(os.PathSeparator), root+string(os.PathSeparator)) &&
			filepath.Clean(p) != root {
			http.NotFound(w, r)
			return
		}
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}
