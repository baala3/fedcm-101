package sp

import "net/http"

type indexPageData struct {
	IdPConfigURL string
	ClientID     string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if currentClaims(r) != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	_ = s.indexTmpl.Execute(w, indexPageData{
		IdPConfigURL: IdPConfigURL,
		ClientID:     SPOrigin,
	})
}

type profilePageData struct {
	Name         string
	GivenName    string
	Email        string
	Picture      string
	IdPConfigURL string
	ClientID     string
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r)
	if claims == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	_ = s.profileTmpl.Execute(w, profilePageData{
		Name:         claims.Name,
		GivenName:    claims.GivenName,
		Email:        claims.Email,
		Picture:      claims.Picture,
		IdPConfigURL: IdPConfigURL,
		ClientID:     SPOrigin,
	})
}
