package idp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fedcm-demo/internal/jwtutil"
)

// handleAssertion serves the credentialed id_assertion_endpoint. The browser
// POSTs here (form-encoded) after the user picks an account in the FedCM
// chooser. We verify the request actually came from the browser's FedCM
// machinery and from the RP origin we know about, mint a signed id token,
// and record a grant so future sign-ins can skip the consent dialog.
func (s *Server) handleAssertion(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" && dest != "webidentity" {
		writeAssertionError(w, http.StatusForbidden, "unauthorized_client", "")
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != RPOrigin {
		writeAssertionError(w, http.StatusForbidden, "unauthorized_client", "")
		return
	}

	user := s.currentUser(r)
	if user == nil {
		writeAssertionError(w, http.StatusUnauthorized, "access_denied", "")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAssertionError(w, http.StatusBadRequest, "invalid_request", "")
		return
	}
	accountID := r.PostForm.Get("account_id")
	clientID := r.PostForm.Get("client_id")
	nonce := r.PostForm.Get("nonce")

	if accountID != strconv.FormatInt(user.ID, 10) {
		writeAssertionError(w, http.StatusForbidden, "access_denied", "")
		return
	}
	if clientID != RPOrigin {
		writeAssertionError(w, http.StatusForbidden, "unauthorized_client", "")
		return
	}

	token, err := jwtutil.Mint(accountID, clientID, nonce, user.Email, user.Name, user.GivenName, user.Picture)
	if err != nil {
		writeAssertionError(w, http.StatusInternalServerError, "server_error", "")
		return
	}

	if err := s.store.AddGrant(user.ID, clientID); err != nil {
		writeAssertionError(w, http.StatusInternalServerError, "server_error", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func writeAssertionError(w http.ResponseWriter, status int, code, url string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "url": url},
	})
}
