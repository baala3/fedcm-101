package idp

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// handleDisconnect serves the credentialed disconnect_endpoint, triggered by
// the RP calling IdentityCredential.disconnect(). It revokes the RP's
// consent grant so future sign-ins require fresh consent again.
func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" && dest != "webidentity" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != RPOrigin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	user := s.currentUser(r)
	if user == nil {
		http.Error(w, "not signed in", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	clientID := r.PostForm.Get("client_id")
	if clientID != RPOrigin {
		http.Error(w, "unknown client", http.StatusBadRequest)
		return
	}

	removed, err := s.store.RemoveGrant(user.ID, clientID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !removed {
		http.Error(w, "no such grant", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"account_id": strconv.FormatInt(user.ID, 10)})
}
