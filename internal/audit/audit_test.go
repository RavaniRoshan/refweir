package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditLoggerRecordsEvents(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	ev := Event{
		Actor:     "roshan@refweir.dev",
		ActorType: "user",
		Action:    "repo.create_ruleset",
		Resource:  "refweir/refweir.dev:rs-main",
		ClientIP:  "192.168.1.100",
		Severity:  SeverityInfo,
		Details:   "Branch ruleset created",
	}

	if err := logger.Record(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "repo.create_ruleset") {
		t.Errorf("expected output to contain action, got: %s", output)
	}

	var parsed Event
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if parsed.Actor != "roshan@refweir.dev" {
		t.Errorf("expected actor roshan@refweir.dev, got %s", parsed.Actor)
	}
	if parsed.ID == "" {
		t.Errorf("expected auto-generated ID, got empty")
	}
	if parsed.Timestamp.IsZero() {
		t.Errorf("expected auto-generated Timestamp, got zero")
	}
}

func TestAuditLoggerRecent(t *testing.T) {
	logger := NewLogger(nil)
	SetDefaultLogger(logger)

	for i := 1; i <= 5; i++ {
		_ = Log(Event{
			Actor:    "admin",
			Action:   "test.action",
			Resource: "test",
			Details:  "item",
		})
	}

	recent := Recent(3)
	if len(recent) != 3 {
		t.Fatalf("expected 3 events, got %d", len(recent))
	}
}

func TestAuditFileLoggerPersistsAndCloses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	logger, err := NewFileLogger(path)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}
	if err := logger.Record(Event{
		Actor:     "agent:claude",
		ActorType: "agent",
		Action:    "agent.lease_issued",
		Resource:  "repo-1",
		Severity:  SeverityInfo,
		Details:   "durable audit test",
	}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := logger.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Close is idempotent.
	if err := logger.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), string(raw))
	}
	var parsed Event
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("unmarshal persisted event: %v", err)
	}
	if parsed.Action != "agent.lease_issued" || parsed.Actor != "agent:claude" {
		t.Fatalf("unexpected persisted event: %+v", parsed)
	}
	if parsed.ID == "" || parsed.Timestamp.IsZero() {
		t.Fatalf("expected ID and timestamp, got %+v", parsed)
	}
}
