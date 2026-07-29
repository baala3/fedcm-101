package idp

import "net/http"

type loginPageData struct {
	Error       string
	RedirectFmt string
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if s.currentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	_ = s.loginTmpl.Execute(w, loginPageData{})
}

// handleLoginSubmit authenticates the demo user, opens a session, and tells
// the browser (via the Login Status API's Set-Login header) that this user
// is now logged in to the IdP. That header is what lets FedCM know it's
// worth trying the accounts_endpoint on this RP's behalf.
func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	user, err := s.store.Authenticate(username, password)
	if err != nil {
		_ = s.loginTmpl.Execute(w, loginPageData{Error: "Invalid username or password"})
		return
	}

	token := s.sessions.create(user.ID)
	setSessionCookie(w, token)
	w.Header().Set("Set-Login", "logged-in")

	// This page is normally loaded inside the popup/tab the browser opens
	// for the FedCM login_url flow. IdentityProvider.close() tells the
	// browser the IdP is done, so it closes the popup and retries the
	// accounts_endpoint automatically. window.IdentityProvider exists on
	// every page (not just inside that popup), so close() silently no-ops
	// when this is loaded as a normal top-level navigation (e.g. someone
	// visiting /login directly) instead of throwing — we can't feature-
	// detect our way out of that, so we always arm a fallback redirect and
	// only rely on it if the popup-close didn't actually tear the page down.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html><head><style>
  body { font-family: system-ui, sans-serif; display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; margin: 0; color: #5f6368; }
  .spinner { width: 28px; height: 28px; border: 3px solid #e0e0e0; border-top-color: #1a73e8; border-radius: 50%; animation: spin 0.8s linear infinite; margin-bottom: 12px; }
  @keyframes spin { to { transform: rotate(360deg); } }
</style></head>
<body>
  <div class="spinner"></div>
  <p>Signed in. Redirecting…</p>
  <script>
    var fallback = setTimeout(function () { window.location.href = "/"; }, 1200);
    try {
      if (window.IdentityProvider && IdentityProvider.close) {
        IdentityProvider.close();
      } else {
        clearTimeout(fallback);
        window.location.href = "/";
      }
    } catch (e) {
      clearTimeout(fallback);
      window.location.href = "/";
    }
  </script>
</body></html>`))
}

// handleLogout clears the IdP session and tells the browser (again via the
// Login Status API) that nobody is logged in here anymore.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		s.sessions.delete(c.Value)
	}
	clearSessionCookie(w)
	w.Header().Set("Set-Login", "logged-out")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
