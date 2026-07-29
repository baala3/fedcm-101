// Package jwtutil mints and verifies the FedCM id token exchanged between
// the demo IdP and SP. It uses a hardcoded shared HMAC secret because this
// is a local-only demo — a real IdP would sign with an asymmetric key and
// publish a JWKS instead.
package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DemoSharedSecret is used to sign/verify id tokens. Never do this in
// production — it only works here because the IdP and SP are the same
// trusted demo codebase.
const DemoSharedSecret = "fedcm-demo-shared-secret-do-not-use-in-prod"

const (
	IssuerIdP = "http://localhost:8080"
)

// Claims is the id token payload returned to the RP from the IdP's
// id_assertion_endpoint.
type Claims struct {
	jwt.RegisteredClaims
	Email     string `json:"email"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	Picture   string `json:"picture,omitempty"`
}

// Mint signs an id token for the given account, scoped to clientID (the RP
// origin acting as FedCM client_id) and the given nonce (if the RP supplied
// one in its navigator.credentials.get call).
func Mint(accountID, clientID, nonce, email, name, givenName, picture string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    IssuerIdP,
			Subject:   accountID,
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
			ID:        nonce,
		},
		Email:     email,
		Name:      name,
		GivenName: givenName,
		Picture:   picture,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(DemoSharedSecret))
}

// Verify checks the token's signature/expiry and that it was issued for
// clientID, returning the parsed claims.
func Verify(tokenString, clientID string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(DemoSharedSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	aud, err := claims.GetAudience()
	if err != nil || len(aud) == 0 || aud[0] != clientID {
		return nil, errors.New("token audience does not match client_id")
	}
	return claims, nil
}
