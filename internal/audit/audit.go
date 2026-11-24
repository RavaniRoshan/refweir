// Package audit implements RFC 5424 compliant structured security audit logging for refweir.dev.
package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"refweir.dev/pkg/idgen"
)

// Severity represents the enterprise compliance level of an audit event.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event records a verifiable security, governance, or Git transport action.
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	ActorType string                 `json:"actorType"` // "user", "agent", "system"
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	ClientIP  string                 `json:"clientIp"`
	Severity  Severity               `json:"severity"`
	Details   string                 `json:"details"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Logger emits append-only structured audit records.
type Logger struct {
	mu     sync.RWMutex
	writer io.Writer
	closer io.Closer
	events []Event
	maxBuf int
}

// DefaultLogger is the system-wide default audit logger.
var (
	defaultLoggerMu sync.RWMutex
	DefaultLogger   = NewLogger(os.Stderr)
)

// SetDefaultLogger overrides the global audit logger.
func SetDefaultLogger(l *Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	DefaultLogger = l
}

// Log records an event using the DefaultLogger.
func Log(ev Event) error {
	defaultLoggerMu.RLock()
	l := DefaultLogger
	defaultLoggerMu.RUnlock()
	if l == nil {
		return nil
	}
	return l.Record(ev)
}

// Recent returns the latest recorded events from DefaultLogger.
func Recent(limit int) []Event {
	defaultLoggerMu.RLock()
	l := DefaultLogger
	defaultLoggerMu.RUnlock()
	if l == nil {
		return nil
	}
	return l.Recent(limit)
}

// Filter returns up to limit matching events from DefaultLogger.
func Filter(actor, action string, limit int) []Event {
	defaultLoggerMu.RLock()
	l := DefaultLogger
	defaultLoggerMu.RUnlock()
	if l == nil {
		return nil
	}
	return l.Filter(actor, action, limit)
}

// NewLogger constructs an audit Logger writing to the specified destination.
func NewLogger(w io.Writer) *Logger {
	return &Logger{
		writer: w,
		events: make([]Event, 0, 100),
		maxBuf: 1000,
	}
}

// NewFileLogger initializes or opens an append-only JSON-lines audit log file.
func NewFileLogger(filePath string) (*Logger, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) // #nosec G304 -- path comes from operator configuration or a validated internal identifier
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	l := NewLogger(f)
	l.closer = f
	return l, nil
}

// Close flushes and closes the underlying log file, if the logger owns one.
// It is safe to call on nil or on a logger without a file.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer == nil {
		return nil
	}
	err := l.closer.Close()
	l.closer = nil
	l.writer = nil
	return err
}

// Record outputs an immutable JSON-encoded audit entry and buffers it in memory.
func (l *Logger) Record(ev Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	if ev.ID == "" {
		ev.ID = idgen.New("aud")
	}

	if len(l.events) >= l.maxBuf {
		l.events = l.events[1:]
	}
	l.events = append(l.events, ev)

	if l.writer != nil {
		bytes, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("audit: marshal event: %w", err)
		}
		bytes = append(bytes, '\n')
		_, err = l.writer.Write(bytes)
		return err
	}
	return nil
}

// Recent returns up to limit latest events, ordered newest first.
func (l *Logger) Recent(limit int) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.events)
	if limit <= 0 || limit > n {
		limit = n
	}

	result := make([]Event, limit)
	for i := 0; i < limit; i++ {
		result[i] = l.events[n-1-i]
	}
	return result
}

// Sync flushes pending buffered writes to persistent storage if writer is an *os.File.
func (l *Logger) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if f, ok := l.writer.(*os.File); ok {
		return f.Sync()
	}
	return nil
}

// Filter returns up to limit latest events matching optional actor and action filters, ordered newest first.
func (l *Logger) Filter(actor, action string, limit int) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.events)
	if limit <= 0 {
		limit = 100
	}

	result := make([]Event, 0, min(limit, n))
	for i := n - 1; i >= 0 && len(result) < limit; i-- {
		ev := l.events[i]
		if actor != "" && ev.Actor != actor {
			continue
		}
		if action != "" && ev.Action != action {
			continue
		}
		result = append(result, ev)
	}
	return result
}

// FilterAdvanced returns events matching actor, action, severity, and temporal boundaries (since <= Timestamp <= until).
// Returns results ordered newest-first.
func (l *Logger) FilterAdvanced(actor, action string, severity Severity, since, until time.Time, limit int) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.events)
	if limit <= 0 {
		limit = 1000
	}

	result := make([]Event, 0, min(limit, n))
	for i := n - 1; i >= 0 && len(result) < limit; i-- {
		ev := l.events[i]
		if actor != "" && ev.Actor != actor {
			continue
		}
		if action != "" && ev.Action != action {
			continue
		}
		if severity != "" && ev.Severity != severity {
			continue
		}
		if !since.IsZero() && ev.Timestamp.Before(since) {
			continue
		}
		if !until.IsZero() && ev.Timestamp.After(until) {
			continue
		}
		result = append(result, ev)
	}
	return result
}

// ReadLogFile reads an append-only JSON-lines audit log from disk and returns matching events,
// ordered newest-first.
func ReadLogFile(filePath string, actor, action string, severity Severity, since, until time.Time, limit int) ([]Event, error) {
	f, err := os.Open(filePath) // #nosec G304 -- path comes from operator configuration
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if limit <= 0 {
		limit = 10000
	}

	var events []Event
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		if actor != "" && ev.Actor != actor {
			continue
		}
		if action != "" && ev.Action != action {
			continue
		}
		if severity != "" && ev.Severity != severity {
			continue
		}
		if !since.IsZero() && ev.Timestamp.Before(since) {
			continue
		}
		if !until.IsZero() && ev.Timestamp.After(until) {
			continue
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Reverse to order newest-first
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	if len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

// Export returns events from disk log (if configured via REFWEIR_AUDIT_LOG and file exists)
// or from the DefaultLogger in-memory buffer.
func Export(actor, action string, severity Severity, since, until time.Time, limit int) []Event {
	logPath := os.Getenv("REFWEIR_AUDIT_LOG")
	if logPath != "" {
		if evs, err := ReadLogFile(logPath, actor, action, severity, since, until, limit); err == nil && len(evs) > 0 {
			return evs
		}
	}

	defaultLoggerMu.RLock()
	l := DefaultLogger
	defaultLoggerMu.RUnlock()
	if l == nil {
		return nil
	}
	return l.FilterAdvanced(actor, action, severity, since, until, limit)
}
