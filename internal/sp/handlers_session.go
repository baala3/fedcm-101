package sp

import (
	"encoding/json"
	"net/http"

	"fedcm-demo/internal/jwtutil"
)

type createSessionRequest struct {
	Token string `json:"token"`
}

// handleCreateSession is called by the page's own JS (same-origin fetch)
// right after navigator.credentials.get() resolves with an IdentityCredential.
// It verifies the id token the IdP minted and, if valid, opens an SP-side
// session for it.
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	claims, err := jwtutil.Verify(req.Token, SPOrigin)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	setSessionCookie(w, req.Token)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"name": claims.Name})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleDisconnected is called after the page's JS successfully invokes
// IdentityCredential.disconnect() against the IdP, so the SP also drops its
// local session for the account that just got unlinked.
func (s *Server) handleDisconnected(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
