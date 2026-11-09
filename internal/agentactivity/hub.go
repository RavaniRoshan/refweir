package agentactivity

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"refweir.dev/pkg/idgen"
)

// Activity represents an event executed by an autonomous AI coding agent.
type Activity struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	AgentName  string    `json:"agentName"`
	AgentType  string    `json:"agentType"` // "claude-code", "cursor", "codex", "refweir-mcp", "reviewer"
	Action     string    `json:"action"`
	ToolName   string    `json:"toolName,omitempty"`
	Resource   string    `json:"resource"`
	Status     string    `json:"status"` // "success", "failed", "in_progress"
	DurationMs int       `json:"durationMs,omitempty"`
	Details    string    `json:"details,omitempty"`
}

// Hub maintains an in-memory thread-safe bounded buffer of recent agent operations
// and broadcasts live events to active subscribers.
type Hub struct {
	mu          sync.RWMutex
	maxEntries  int
	entries     []Activity
	subscribers map[chan Activity]struct{}
}

// NewHub creates a new agent activity hub with the given capacity limit.
func NewHub(maxEntries int) *Hub {
	if maxEntries <= 0 {
		maxEntries = 200
	}
	return &Hub{
		maxEntries:  maxEntries,
		entries:     make([]Activity, 0, maxEntries),
		subscribers: make(map[chan Activity]struct{}),
	}
}

// DefaultHub is the singleton activity hub used by refweird.
var DefaultHub = NewHub(200)

// Record adds an activity entry to the buffer and fans out to active subscribers.
func (h *Hub) Record(act Activity) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if act.ID == "" {
		act.ID = idgen.New("act")
	}
	if act.Timestamp.IsZero() {
		act.Timestamp = time.Now().UTC()
	}

	h.entries = append(h.entries, act)
	if len(h.entries) > h.maxEntries {
		h.entries = h.entries[len(h.entries)-h.maxEntries:]
	}

	for ch := range h.subscribers {
		select {
		case ch <- act:
		default:
			// Non-blocking drop for slow subscriber to avoid latency spikes
		}
	}
}

// Subscribe registers a new subscriber channel with the given buffer size.
// Returns the receive channel and a cleanup cancel func.
func (h *Hub) Subscribe(bufferSize int) (<-chan Activity, func()) {
	if bufferSize <= 0 {
		bufferSize = 32
	}
	ch := make(chan Activity, bufferSize)

	h.mu.Lock()
	if h.subscribers == nil {
		h.subscribers = make(map[chan Activity]struct{})
	}
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subscribers, ch)
			close(ch)
			h.mu.Unlock()
		})
	}

	return ch, cancel
}

// Recent returns the most recent agent activities up to limit, newest first.
func (h *Hub) Recent(limit int) []Activity {
	h.mu.RLock()
	defer h.mu.RUnlock()

	n := len(h.entries)
	if n == 0 {
		return []Activity{}
	}
	if limit <= 0 || limit > n {
		limit = n
	}

	out := make([]Activity, limit)
	for i := 0; i < limit; i++ {
		out[i] = h.entries[n-1-i]
	}
	return out
}

// Count returns the number of entries in the buffer.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// Clear resets all stored entries.
func (h *Hub) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = make([]Activity, 0, h.maxEntries)
}

// Handler returns an HTTP handler for streaming recent agent operations as a JSON snapshot.
func Handler(h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		limit := 50
		activities := h.Recent(limit)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"activities": activities,
			"count":      len(activities),
		})
	}
}
