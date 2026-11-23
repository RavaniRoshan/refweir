package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"refweir.dev/internal/store"
	"refweir.dev/pkg/scopes"
)

func setupTestStore(t *testing.T) (*store.DB, string) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userID := "user_admin1"
	if _, err := db.UpsertUserByEmail(t.Context(), "admin@refweir.dev", func() string { return userID }); err != nil {
		t.Fatalf("UpsertUserByEmail: %v", err)
	}
	return db, userID
}

func TestServerExchangeAndDBResolver(t *testing.T) {
	db, userID := setupTestStore(t)
	handler := ExchangeHandler(db)

	rawKey, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey: %v", err)
	}

	apiKey := store.APIKey{
		ID:        "key_test1",
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      "bootstrap-key",
		Scopes:    scopes.NamespaceRepositoriesRead + " " + scopes.NamespaceRepositoriesCreate,
	}
	if err := db.CreateAPIKey(t.Context(), apiKey); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	// 1. Missing key
	req := httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing key status = %d, want 401", rec.Code)
	}

	// 2. Invalid prefix
	req = httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer invalidprefix_123")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid prefix status = %d, want 400", rec.Code)
	}

	// 3. Unknown key
	req = httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer key_unknown1234567890abcdef")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown key status = %d, want 401", rec.Code)
	}

	// 4. Valid exchange
	req = httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid exchange status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp exchangeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal exchange response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected non-empty accessToken in response")
	}

	// 5. Verify DatabaseResolver resolves the exchanged access token
	resolver := NewDatabaseResolver(db)
	authedReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	authedReq.Header.Set("Authorization", "Bearer "+resp.AccessToken)

	tok, err := resolver(authedReq)
	if err != nil {
		t.Fatalf("resolver failed on valid token: %v", err)
	}
	if tok.Subject() != userID {
		t.Fatalf("token subject = %s, want %s", tok.Subject(), userID)
	}
	if tok.Type() != TypeUser {
		t.Fatalf("token type = %s, want user", tok.Type())
	}
	if len(tok.Scopes()) != 2 {
		t.Fatalf("expected 2 scopes, got %v", tok.Scopes())
	}

	// 6. DatabaseResolver rejects raw key directly (must exchange)
	rawKeyReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	rawKeyReq.Header.Set("Authorization", "Bearer "+rawKey)
	_, err = resolver(rawKeyReq)
	if err != errMustExchange {
		t.Fatalf("expected errMustExchange for raw key, got %v", err)
	}

	// 7. DatabaseResolver rejects arbitrary unrecognized token
	badReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	badReq.Header.Set("Authorization", "Bearer totally_fake_token")
	_, err = resolver(badReq)
	if err == nil {
		t.Fatal("expected error for fake token, got nil")
	}
}

func TestExchangeRevokedAndExpiredKeys(t *testing.T) {
	db, userID := setupTestStore(t)
	handler := ExchangeHandler(db)

	rawKey, prefix, hash, _ := GenerateAPIKey()
	apiKey := store.APIKey{
		ID:        "key_revoked",
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      "revoked-key",
		Scopes:    "*",
		RevokedAt: sql.NullString{String: time.Now().UTC().Format(time.RFC3339), Valid: true},
	}
	if err := db.CreateAPIKey(t.Context(), apiKey); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked key status = %d, want 401", rec.Code)
	}
}

func TestDatabaseResolverAppTokenVerification(t *testing.T) {
	db, _ := setupTestStore(t)
	ctx := t.Context()

	// 1. Generate app keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	pubKeyB64 := base64.RawURLEncoding.EncodeToString(pub)

	// 2. Register namespace and app in database with the public signing key
	_ = db.CreateNamespace(ctx, store.Namespace{
		ID:   "ns_default",
		Slug: "default",
		Type: "team",
		Name: "Default Namespace",
	})
	app := store.App{
		ID:          "app_sec_audit",
		NamespaceID: "ns_default",
		DisplayName: "Security Auditor App",
		SigningKey:  pubKeyB64,
	}
	if _, err := db.CreateApp(ctx, app); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}

	resolver := NewDatabaseResolver(db)

	// 3. Create valid, authentically signed App JWT
	payload, _ := json.Marshal(map[string]any{
		"sub":   app.ID,
		"scope": "app:metadata:read repo:read",
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	})
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA"}`))
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	signedContent := header + "." + payloadEnc

	sig := ed25519.Sign(priv, []byte(signedContent))
	validJWT := signedContent + "." + base64.RawURLEncoding.EncodeToString(sig)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repos", nil)
	req.Header.Set("Authorization", "Bearer "+validJWT)

	tok, err := resolver(req)
	if err != nil {
		t.Fatalf("resolver failed on authentically signed App JWT: %v", err)
	}
	if tok.Type() != TypeApp {
		t.Errorf("tok.Type() = %q, want %q", tok.Type(), TypeApp)
	}
	if tok.Subject() != app.ID {
		t.Errorf("tok.Subject() = %q, want %q", tok.Subject(), app.ID)
	}

	// 4. Test forged App JWT (signed with an attacker's rogue key)
	_, roguePriv, _ := ed25519.GenerateKey(rand.Reader)
	rogueSig := ed25519.Sign(roguePriv, []byte(signedContent))
	forgedJWT := signedContent + "." + base64.RawURLEncoding.EncodeToString(rogueSig)

	forgedReq := httptest.NewRequest(http.MethodGet, "/api/v1/repos", nil)
	forgedReq.Header.Set("Authorization", "Bearer "+forgedJWT)

	_, err = resolver(forgedReq)
	if err == nil {
		t.Fatalf("resolver accepted forged App JWT, expected error")
	}

	// 5. Test App JWT for non-existent app
	nonExistentPayload, _ := json.Marshal(map[string]any{
		"sub":   "app_does_not_exist_999",
		"scope": "repo:read",
	})
	nonExistentContent := header + "." + base64.RawURLEncoding.EncodeToString(nonExistentPayload)
	nonExistentSig := ed25519.Sign(priv, []byte(nonExistentContent))
	nonExistentJWT := nonExistentContent + "." + base64.RawURLEncoding.EncodeToString(nonExistentSig)

	nonExistentReq := httptest.NewRequest(http.MethodGet, "/api/v1/repos", nil)
	nonExistentReq.Header.Set("Authorization", "Bearer "+nonExistentJWT)

	_, err = resolver(nonExistentReq)
	if err == nil {
		t.Fatalf("resolver accepted JWT for non-existent app, expected error")
	}
}
