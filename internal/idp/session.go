package idp

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

const sessionCookieName = "idp_session"

// sessionStore is a trivial in-memory session table (token -> user id). A
// demo IdP doesn't need sessions to survive a restart; grants/users are the
// only state that lives in sqlite.
type sessionStore struct {
	mu   sync.RWMutex
	byID map[string]int64
}

func newSessionStore() *sessionStore {
	return &sessionStore{byID: make(map[string]int64)}
}

func (s *sessionStore) create(userID int64) string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	s.mu.Lock()
	s.byID[token] = userID
	s.mu.Unlock()
	return token
}

func (s *sessionStore) lookup(token string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byID[token]
	return id, ok
}

func (s *sessionStore) delete(token string) {
	s.mu.Lock()
	delete(s.byID, token)
	s.mu.Unlock()
}

// currentUser resolves the logged-in user (if any) from the session cookie.
func (srv *Server) currentUser(r *http.Request) *User {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}
	userID, ok := srv.sessions.lookup(c.Value)
	if !ok {
		return nil
	}
	u, err := srv.store.GetUserByID(userID)
	if err != nil {
		return nil
	}
	return u
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
}
