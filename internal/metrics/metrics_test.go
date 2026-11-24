package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsRegistry(t *testing.T) {
	reg := NewRegistry()
	reg.IncHTTPRequests()
	reg.IncGitTransfers()
	reg.IncMCPCalls()
	reg.IncSecretPushBlocks()

	if reg.HTTPRequestsTotal() != 1 {
		t.Errorf("expected 1 http request, got %d", reg.HTTPRequestsTotal())
	}
	if reg.GitTransfersTotal() != 1 {
		t.Errorf("expected 1 git transfer, got %d", reg.GitTransfersTotal())
	}
	if reg.MCPCallsTotal() != 1 {
		t.Errorf("expected 1 mcp call, got %d", reg.MCPCallsTotal())
	}
	if reg.SecretPushBlocksTotal() != 1 {
		t.Errorf("expected 1 secret push block, got %d", reg.SecretPushBlocksTotal())
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	reg.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"refweir_uptime_seconds",
		"refweir_http_requests_total 1",
		"refweir_git_transfers_total 1",
		"refweir_mcp_tool_calls_total 1",
		"refweir_secret_push_blocks_total 1",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected metric output to contain %q, got: %s", expected, body)
		}
	}
}

func TestHistoricalMetricsRetention(t *testing.T) {
	reg := NewRegistry()
	history := reg.History()
	if len(history) == 0 {
		t.Fatalf("expected seeded history, got 0 samples")
	}

	initialLen := len(history)
	sample := reg.SampleNow(42.5, 30, 1.2)
	if sample.AllocatedMemoryMb != 42.5 {
		t.Errorf("expected allocatedMemoryMb = 42.5, got %f", sample.AllocatedMemoryMb)
	}
	if sample.Goroutines != 30 {
		t.Errorf("expected goroutines = 30, got %d", sample.Goroutines)
	}

	newHistory := reg.History()
	if len(newHistory) != initialLen+1 {
		t.Errorf("expected %d samples, got %d", initialLen+1, len(newHistory))
	}

	// Test periodic sampler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reg.StartPeriodicSampler(ctx, 20*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	cancel()

	sampledHistory := reg.History()
	if len(sampledHistory) <= len(newHistory) {
		t.Errorf("expected periodic sampler to add samples, had %d, now %d", len(newHistory), len(sampledHistory))
	}
}
