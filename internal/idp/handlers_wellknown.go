package idp

import "net/http"

// handleWellKnown serves /.well-known/web-identity, which the browser
// fetches (without credentials) to confirm this origin actually owns the
// config URL it was given by the RP. Must not redirect.
func (s *Server) handleWellKnown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{
		"provider_urls": ["` + IdPOrigin + `/fedcm.json"],
		"accounts_endpoint": "` + IdPOrigin + `/fedcm/accounts",
		"login_url": "` + IdPOrigin + `/login"
	}`))
}

// handleIcon serves the small badge shown next to accounts in the FedCM
// chooser UI (referenced from the config's branding.icons).
func (s *Server) handleIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"><rect width="32" height="32" rx="6" fill="#1a73e8"/><text x="16" y="21" font-size="16" text-anchor="middle" fill="#fff" font-family="sans-serif">ID</text></svg>`))
}
