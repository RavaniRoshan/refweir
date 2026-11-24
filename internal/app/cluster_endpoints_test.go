package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"refweir.dev/internal/auth"
	"refweir.dev/internal/config"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func TestCluster_Endpoints_AuthorizationAndExecution(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer application.Close()

	handler := application.Handler()
	ctx := t.Context()

	// 1. Create an admin user and member user
	adminUser := store.User{
		ID:    idgen.NewUser(),
		Email: "admin@example.com",
		Login: "admin-sarah",
		Role:  "admin",
	}
	if err := application.Store.CreateUserWithPassword(ctx, adminUser); err != nil {
		t.Fatalf("CreateUserWithPassword admin: %v", err)
	}

	memberUser := store.User{
		ID:    idgen.NewUser(),
		Email: "member@example.com",
		Login: "bob-dev",
		Role:  "member",
	}
	if err := application.Store.CreateUserWithPassword(ctx, memberUser); err != nil {
		t.Fatalf("CreateUserWithPassword member: %v", err)
	}

	// Create session for admin
	rawAdminToken, adminHash, _ := auth.GenerateSessionToken()
	_ = application.Store.CreateSession(ctx, store.Session{
		ID:        idgen.New("sess"),
		UserID:    adminUser.ID,
		TokenHash: adminHash,
		ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339),
	})

	// Create session for member
	rawMemberToken, memberHash, _ := auth.GenerateSessionToken()
	_ = application.Store.CreateSession(ctx, store.Session{
		ID:        idgen.New("sess"),
		UserID:    memberUser.ID,
		TokenHash: memberHash,
		ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339),
	})

	// 2. Unauthenticated calls to cluster mutation endpoints must fail with 401
	for _, ep := range []string{"/api/v1/cluster/failover/promote", "/api/v1/cluster/checkpoint", "/api/v1/cluster/snapshot"} {
		req := httptest.NewRequest(http.MethodPost, ep, bytes.NewReader([]byte("{}")))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: expected 401 Unauthorized for anonymous caller, got %d", ep, rec.Code)
		}
	}

	// 3. Member calls to cluster mutation endpoints must fail with 403 Forbidden
	for _, ep := range []string{"/api/v1/cluster/failover/promote", "/api/v1/cluster/checkpoint", "/api/v1/cluster/snapshot"} {
		req := httptest.NewRequest(http.MethodPost, ep, bytes.NewReader([]byte("{}")))
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawMemberToken})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s: expected 403 Forbidden for non-admin caller, got %d", ep, rec.Code)
		}
	}

	// 4. Admin call to GET /api/v1/cluster/replication succeeds
	reqRep := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/replication", nil)
	reqRep.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawAdminToken})
	recRep := httptest.NewRecorder()
	handler.ServeHTTP(recRep, reqRep)
	if recRep.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/cluster/replication: expected 200, got %d", recRep.Code)
	}
	var repStatus map[string]any
	if err := json.Unmarshal(recRep.Body.Bytes(), &repStatus); err != nil {
		t.Fatalf("failed to parse replication status: %v", err)
	}
	if repStatus["role"] != "primary" {
		t.Errorf("expected role primary, got %v", repStatus["role"])
	}

	// 5. Admin call to POST /api/v1/cluster/checkpoint succeeds
	reqChk := httptest.NewRequest(http.MethodPost, "/api/v1/cluster/checkpoint", nil)
	reqChk.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawAdminToken})
	recChk := httptest.NewRecorder()
	handler.ServeHTTP(recChk, reqChk)
	if recChk.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/cluster/checkpoint: expected 200, got %d: %s", recChk.Code, recChk.Body.String())
	}

	// 6. Admin call to POST /api/v1/cluster/snapshot succeeds (201 Created)
	reqSnap := httptest.NewRequest(http.MethodPost, "/api/v1/cluster/snapshot", nil)
	reqSnap.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawAdminToken})
	recSnap := httptest.NewRecorder()
	handler.ServeHTTP(recSnap, reqSnap)
	if recSnap.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/cluster/snapshot: expected 201 Created, got %d: %s", recSnap.Code, recSnap.Body.String())
	}
	var snapMeta map[string]any
	if err := json.Unmarshal(recSnap.Body.Bytes(), &snapMeta); err != nil {
		t.Fatalf("failed to decode snapshot metadata: %v", err)
	}
	if snapMeta["filename"] == "" || snapMeta["checksumSha256"] == "" {
		t.Errorf("incomplete snapshot metadata: %+v", snapMeta)
	}
}
