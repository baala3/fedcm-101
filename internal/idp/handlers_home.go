package idp

import "net/http"

type homePageData struct {
	Name      string
	Email     string
	GivenName string
	Picture   string
	Grants    []string
	RPOrigin  string
}

// handleHome is the IdP's own account page: shows who's signed in and which
// RPs they've granted consent to, with a self-service revoke button per RP
// (distinct from the RP-initiated IdentityCredential.disconnect() flow).
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	user := s.currentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	grants, err := s.store.ClientsGrantedFor(user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = s.homeTmpl.Execute(w, homePageData{
		Name:      user.Name,
		Email:     user.Email,
		GivenName: user.GivenName,
		Picture:   user.Picture,
		Grants:    grants,
		RPOrigin:  RPOrigin,
	})
}

// handleRevoke lets the signed-in user drop a consent grant from the IdP
// side, e.g. without visiting the RP at all.
func (s *Server) handleRevoke(w http.ResponseWriter, r *http.Request) {
	user := s.currentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	clientID := r.PostForm.Get("client_id")
	if _, err := s.store.RemoveGrant(user.ID, clientID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
