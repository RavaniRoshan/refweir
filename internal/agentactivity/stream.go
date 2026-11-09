package agentactivity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// StreamHandler returns an SSE handler that pushes agent activities to the
// client in real time as they are recorded, with periodic heartbeats.
func StreamHandler(h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Subscribe to real-time broadcasts before taking the snapshot to ensure
		// zero event loss during handshake and initial network flush.
		subCh, cancel := h.Subscribe(64)
		defer cancel()

		// Send initial snapshot
		activities := h.Recent(50)
		data, _ := json.Marshal(map[string]any{"activities": activities})
		if _, err := fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", data); err != nil {
			return
		}
		flusher.Flush()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case act, ok := <-subCh:
				if !ok {
					return
				}
				actData, err := json.Marshal(act)
				if err == nil {
					if _, err := fmt.Fprintf(w, "event: activity\ndata: %s\n\n", actData); err != nil {
						return
					}
					flusher.Flush()
				}
			case <-ticker.C:
				// Periodic heartbeat for reverse-proxy keepalive
				if _, err := fmt.Fprintf(w, "event: heartbeat\ndata: {\"ts\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339)); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}
