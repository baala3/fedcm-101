package idp

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type account struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	GivenName       string   `json:"given_name"`
	Email           string   `json:"email"`
	Picture         string   `json:"picture,omitempty"`
	ApprovedClients []string `json:"approved_clients,omitempty"`
}

// handleAccounts serves the credentialed accounts_endpoint. The browser
// sends this request with the IdP's cookies attached and expects a 401 when
// nobody is signed in (which the Login Status API mismatch handling relies
// on to clear a stale "logged-in" hint).
func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" && dest != "webidentity" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	user := s.currentUser(r)
	if user == nil {
		http.Error(w, "not signed in", http.StatusUnauthorized)
		return
	}

	approved, err := s.store.ClientsGrantedFor(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accounts": []account{
			{
				ID:              strconv.FormatInt(user.ID, 10),
				Name:            user.Name,
				GivenName:       user.GivenName,
				Email:           user.Email,
				Picture:         user.Picture,
				ApprovedClients: approved,
			},
		},
	})
}
