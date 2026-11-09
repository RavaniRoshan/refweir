package agentactivity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// StreamHandler returns an SSE handler that pushes agent activities to the
// client as they are recorded, with heartbeats every 15 seconds.
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

		// Send initial snapshot
		activities := h.Recent(50)
		data, _ := json.Marshal(map[string]any{"activities": activities})
		fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", data)
		flusher.Flush()

		lastCount := h.Count()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				// Heartbeat
				fmt.Fprintf(w, "event: heartbeat\ndata: {\"ts\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
				flusher.Flush()

				// Check for new entries
				current := h.Recent(50)
				if newCount := h.Count(); newCount != lastCount {
					data, _ := json.Marshal(map[string]any{"activities": current})
					fmt.Fprintf(w, "event: update\ndata: %s\n\n", data)
					flusher.Flush()
					lastCount = newCount
				}
			}
		}
	}
}
