package idp

import (
	"encoding/json"
	"net/http"
)

type brandingIcon struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
}

type branding struct {
	BackgroundColor string         `json:"background_color"`
	Color           string         `json:"color"`
	Icons           []brandingIcon `json:"icons"`
}

type idpConfig struct {
	AccountsEndpoint       string   `json:"accounts_endpoint"`
	ClientMetadataEndpoint string   `json:"client_metadata_endpoint"`
	IDAssertionEndpoint    string   `json:"id_assertion_endpoint"`
	DisconnectEndpoint     string   `json:"disconnect_endpoint"`
	LoginURL               string   `json:"login_url"`
	Branding               branding `json:"branding"`
}

// handleConfig serves the IdP's FedCM config document. The browser fetches
// this without credentials and expects it to advertise Sec-Fetch-Dest:
// webidentity, which we check defensively even though Chrome always sends it.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" && dest != "webidentity" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	cfg := idpConfig{
		AccountsEndpoint:       IdPOrigin + "/fedcm/accounts",
		ClientMetadataEndpoint: IdPOrigin + "/fedcm/client_metadata",
		IDAssertionEndpoint:    IdPOrigin + "/fedcm/assertion",
		DisconnectEndpoint:     IdPOrigin + "/fedcm/disconnect",
		LoginURL:               IdPOrigin + "/login",
		Branding: branding{
			BackgroundColor: "#1a73e8",
			Color:           "#ffffff",
			Icons: []brandingIcon{
				{URL: IdPOrigin + "/static/icon.svg", Size: 32},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cfg)
}
