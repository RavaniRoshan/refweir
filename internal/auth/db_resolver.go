package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"refweir.dev/internal/store"
	"refweir.dev/pkg/scopes"
)

// ScopedToken represents an authenticated token with explicit scopes and subject.
type ScopedToken struct {
	tokenType string
	subject   string
	scopes    []string
}

func (s ScopedToken) Type() string     { return s.tokenType }
func (s ScopedToken) Subject() string  { return s.subject }
func (s ScopedToken) Scopes() []string { return s.scopes }

// NewDatabaseResolver returns an auth.Resolver backed by the local store.
// It verifies bearer tokens against stored access_tokens (and app JWTs) rather
// than trusting unverified arbitrary strings.
func NewDatabaseResolver(db *store.DB) Resolver {
	return func(r *http.Request) (Token, error) {
		raw, ok := bearerCredential(r)
		if !ok {
			if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
				sessHash := HashCredential(cookie.Value)
				sess, user, err := db.GetSessionByHash(r.Context(), sessHash)
				if err == nil {
					exp, err := time.Parse(time.RFC3339Nano, sess.ExpiresAt)
					if err == nil && time.Now().After(exp) {
						return nil, errors.New("auth: session expired")
					}
					if user.Status == "suspended" {
						return nil, errors.New("auth: user account is suspended")
					}
					return ScopedToken{
						tokenType: TypeUser,
						subject:   user.ID,
						scopes:    scopes.All(),
					}, nil
				}
			}
			return nil, errMissingToken
		}

		// Exchangeable keys must be exchanged first
		if strings.HasPrefix(raw, "key_") || strings.HasPrefix(raw, "crsr_") {
			return nil, errMustExchange
		}

		// App token: Ed25519 compact JWT
		if appTok, err := ParseAppToken(raw); err == nil && strings.HasPrefix(appTok.Subject(), "app_") {
			// Cryptographically verify app token signature against registered app signing keys
			app, err := db.GetApp(r.Context(), appTok.Subject())
			if err != nil {
				return nil, errors.New("auth: app not found")
			}
			keys, _ := db.ListAppSigningKeys(r.Context(), app.ID)
			var candidateKeyStrings []string
			if app.SigningKey != "" {
				candidateKeyStrings = append(candidateKeyStrings, app.SigningKey)
			}
			for _, k := range keys {
				if k.PublicKey != "" {
					candidateKeyStrings = append(candidateKeyStrings, k.PublicKey)
				}
			}
			if len(candidateKeyStrings) == 0 {
				return nil, errors.New("auth: app has no registered signing key")
			}
			var verified bool
			for _, keyStr := range candidateKeyStrings {
				pubKey, err := ParseEd25519PublicKey(keyStr)
				if err != nil {
					continue
				}
				if err := VerifyAppTokenSignature(raw, pubKey); err == nil {
					verified = true
					break
				}
			}
			if !verified {
				return nil, errors.New("auth: invalid app token signature")
			}
			return appTok, nil
		}

		// Look up in database access_tokens
		tokHash := HashCredential(raw)
		tokRecord, err := db.GetAccessTokenByHash(r.Context(), tokHash)
		if err == nil {
			if tokRecord.RevokedAt.Valid && tokRecord.RevokedAt.String != "" {
				return nil, errors.New("auth: token has been revoked")
			}
			exp, err := time.Parse(time.RFC3339, tokRecord.ExpiresAt)
			if err == nil && time.Now().After(exp) {
				return nil, errors.New("auth: token has expired")
			}
			var granted []string
			if tokRecord.Scopes == "*" {
				granted = scopes.All()
			} else {
				granted = strings.Fields(tokRecord.Scopes)
			}
			return ScopedToken{
				tokenType: tokRecord.TokenType,
				subject:   tokRecord.UserID,
				scopes:    granted,
			}, nil
		}

		return nil, errors.New("auth: unrecognized or invalid token")
	}
}
