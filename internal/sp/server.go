package sp

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	// SPOrigin is this server's own origin, used as the FedCM client_id.
	SPOrigin = "http://localhost:8081"
	// IdPConfigURL is the well-known IdP config this demo RP is wired to.
	IdPConfigURL = "http://localhost:8080/fedcm.json"
)

// Server holds the SP's dependencies: the router and parsed HTML templates.
type Server struct {
	router      *chi.Mux
	indexTmpl   *template.Template
	profileTmpl *template.Template
}

func NewServer() *Server {
	srv := &Server{
		indexTmpl:   template.Must(template.ParseFiles("internal/sp/templates/index.html")),
		profileTmpl: template.Must(template.ParseFiles("internal/sp/templates/profile.html")),
	}
	srv.routes()
	return srv
}

func (s *Server) routes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fileServer := http.FileServer(http.Dir("internal/sp/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", s.handleIndex)
	r.Get("/profile", s.handleProfile)

	s.router = r
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("SP listening on %s (%s)", addr, SPOrigin)
	return http.ListenAndServe(addr, s)
}
