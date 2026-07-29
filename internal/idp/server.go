package idp

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	// IdPOrigin is this server's own origin, used as the JWT issuer and in
	// the well-known/config responses.
	IdPOrigin = "http://localhost:8080"
	// RPOrigin is the only client this demo IdP knows about. A real IdP
	// would look up allowed origins per client_id instead of hardcoding one.
	RPOrigin = "http://localhost:8081"
)

// Server holds the IdP's dependencies: the router, the sqlite-backed store,
// and in-memory sessions.
type Server struct {
	router   *chi.Mux
	store    *Store
	sessions *sessionStore
}

func NewServer(db *sql.DB) *Server {
	srv := &Server{
		store:    NewStore(db),
		sessions: newSessionStore(),
	}
	srv.routes()
	return srv
}

func (s *Server) routes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/static/icon.svg", s.handleIcon)

	r.Get("/.well-known/web-identity", s.handleWellKnown)
	r.Get("/fedcm.json", s.handleConfig)

	s.router = r
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("IdP listening on %s (%s)", addr, IdPOrigin)
	return http.ListenAndServe(addr, s)
}
