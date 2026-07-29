package sp

import (
	"net/http"

	"fedcm-101/internal/jwtutil"
)

const sessionCookieName = "sp_session"

// currentClaims reads the sp_session cookie (which holds the raw id token
// minted by the IdP) and re-verifies it, so an expired or tampered token
// never appears to be a valid session.
func currentClaims(r *http.Request) *jwtutil.Claims {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}
	claims, err := jwtutil.Verify(c.Value, SPOrigin)
	if err != nil {
		return nil
	}
	return claims
}

func setSessionCookie(w http.ResponseWriter, idToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    idToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
