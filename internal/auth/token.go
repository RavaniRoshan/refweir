// Package auth implements the Origin token types and the API key exchange flow.
//
// Origin's auth model has three token types:
//   - app: an Ed25519 JWT issued to an installed app
//   - installation: an "oit_…" token with a ≤15 minute expiry
//   - user: an OAuth user credential used directly as a bearer token
//
// Raw "key_…" or "crsr_…" API keys are not bearer tokens; they must first be
// exchanged via ExchangeAPIKey.
package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	TypeApp          = "app"
	TypeInstallation = "installation"
	TypeUser         = "user"

	// TypeExchangeable marks raw API keys ("key_…"/"crsr_…") that must be
	// exchanged before they can be used as bearer tokens.
	TypeExchangeable = "exchangeable"
)

// InstallTokenMaxLifetime is the maximum lifetime Origin grants an
// installation token.
const InstallTokenMaxLifetime = 15 * time.Minute

// Token is the common interface implemented by every Origin token type.
type Token interface {
	// Type returns the Origin token type: TypeApp, TypeInstallation, or TypeUser.
	Type() string
	// Scopes returns the scopes granted to the token, in order.
	Scopes() []string
	// Subject returns the principal the token acts for.
	Subject() string
}

// AppToken is an Ed25519 JWT issued to an installed app.
type AppToken struct {
	raw     string
	scopes  []string
	subject string
}

// ParseAppToken parses a compact JWT's payload claims. Signature validation
// must be performed against the app's enrolled public key via VerifyAppTokenSignature.
func ParseAppToken(raw string) (AppToken, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return AppToken{}, errors.New("auth: not a compact JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return AppToken{}, fmt.Errorf("auth: app token payload: %w", err)
		}
	}
	var claims struct {
		Sub   string `json:"sub"`
		Scope string `json:"scope"`
		Exp   int64  `json:"exp,omitempty"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return AppToken{}, fmt.Errorf("auth: app token claims: %w", err)
	}
	if claims.Exp != 0 && time.Now().Unix() > claims.Exp {
		return AppToken{}, errors.New("auth: app token has expired")
	}
	return AppToken{raw: raw, scopes: strings.Fields(claims.Scope), subject: claims.Sub}, nil
}

// VerifyAppTokenSignature checks that raw is a compact JWT signed by pubKey.
func VerifyAppTokenSignature(raw string, pubKey ed25519.PublicKey) error {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return errors.New("auth: not a compact JWT")
	}
	signedContent := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		sig, err = base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			return fmt.Errorf("auth: invalid signature encoding: %w", err)
		}
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return errors.New("auth: invalid ed25519 public key size")
	}
	if !ed25519.Verify(pubKey, []byte(signedContent), sig) {
		return errors.New("auth: invalid app token signature")
	}
	return nil
}

// ParseEd25519PublicKey attempts to parse a raw string representation of an Ed25519 public key
// (raw bytes, hex-encoded 64 chars, base64url/base64).
func ParseEd25519PublicKey(raw string) (ed25519.PublicKey, error) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) == ed25519.PublicKeySize {
		return ed25519.PublicKey([]byte(trimmed)), nil
	}
	// Try Hex (64 chars)
	if len(trimmed) == ed25519.PublicKeySize*2 {
		if b, err := hex.DecodeString(trimmed); err == nil && len(b) == ed25519.PublicKeySize {
			return ed25519.PublicKey(b), nil
		}
	}
	// Try Base64 RawURL
	if b, err := base64.RawURLEncoding.DecodeString(trimmed); err == nil && len(b) == ed25519.PublicKeySize {
		return ed25519.PublicKey(b), nil
	}
	// Try Base64 Std
	if b, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(b) == ed25519.PublicKeySize {
		return ed25519.PublicKey(b), nil
	}
	return nil, fmt.Errorf("auth: unsupported ed25519 public key format (length %d)", len(trimmed))
}

func (t AppToken) Type() string     { return TypeApp }
func (t AppToken) Scopes() []string { return t.scopes }
func (t AppToken) Subject() string  { return t.subject }

// InstallToken is an installation-scoped Origin token ("oit_…"). Origin
// bounds its lifetime to at most 15 minutes.
type InstallToken struct {
	raw      string
	issuedAt time.Time
}

// NewInstallToken wraps an "oit_…" token, marking its issuance time.
func NewInstallToken(raw string) InstallToken {
	return InstallToken{raw: raw, issuedAt: time.Now()}
}

func (t InstallToken) Type() string     { return TypeInstallation }
func (t InstallToken) Scopes() []string { return nil }
func (t InstallToken) Subject() string  { return t.raw }

// ExpiresAt returns the latest point at which Origin accepts the token.
func (t InstallToken) ExpiresAt() time.Time { return t.issuedAt.Add(InstallTokenMaxLifetime) }

// Expired reports whether the token's bounded lifetime has elapsed.
func (t InstallToken) Expired() bool { return time.Now().After(t.ExpiresAt()) }

// UserToken is an Origin OAuth user credential. Non-prefixed bearer tokens
// that are not compact JWTs are used directly as user tokens.
type UserToken struct {
	raw string
}

// NewUserToken wraps a raw user credential for direct bearer use.
func NewUserToken(raw string) UserToken { return UserToken{raw: raw} }

func (t UserToken) Type() string     { return TypeUser }
func (t UserToken) Scopes() []string { return nil }
func (t UserToken) Subject() string  { return t.raw }

// DetectType classifies a raw bearer credential by prefix, matching Origin:
// "oit_…" tokens are installation tokens, "key_…"/"crsr_…" are raw API keys
// that must be exchanged before use, and anything else is a user token.
func DetectType(raw string) string {
	switch {
	case strings.HasPrefix(raw, "oit_"):
		return TypeInstallation
	case strings.HasPrefix(raw, "key_"), strings.HasPrefix(raw, "crsr_"):
		return TypeExchangeable
	default:
		return TypeUser
	}
}
