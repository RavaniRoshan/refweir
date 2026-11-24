package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

// DispatchRepoWebhook delivers an outbound event payload to a configured repository webhook.
func DispatchRepoWebhook(ctx context.Context, db *store.DB, hook *store.RepoWebhook, event string, payload any) (*store.RepoWebhookDelivery, error) {
	if hook == nil {
		return nil, fmt.Errorf("webhook is nil")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	guid := idgen.New("guid")
	now := time.Now().UTC()

	reqHeaders := http.Header{
		"Content-Type":       []string{"application/json"},
		"User-Agent":         []string{"refweir-Hookshot/1.0"},
		"X-GitHub-Event":     []string{event},
		"X-GitHub-Delivery":  []string{guid},
		"X-Refweir-Event":    []string{event},
		"X-Refweir-Delivery": []string{guid},
	}

	if hook.Secret != "" {
		mac := hmac.New(sha256.New, []byte(hook.Secret))
		mac.Write(payloadBytes)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		reqHeaders.Set("X-Hub-Signature-256", sig)
	}

	reqHeadersJSON, _ := json.Marshal(reqHeaders)

	delivery := &store.RepoWebhookDelivery{
		WebhookID:      hook.ID,
		Event:          event,
		GUID:           guid,
		RequestHeaders: string(reqHeadersJSON),
		RequestPayload: string(payloadBytes),
		CreatedAt:      now.Format(time.RFC3339),
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: hook.InsecureSSL, // #nosec G402 -- controlled by repo webhook setting
			},
		},
	}

	start := time.Now()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		delivery.ResponseStatusCode = 0
		delivery.ResponseBody = fmt.Sprintf(`{"error": %q}`, err.Error())
		delivery.DurationMs = int(time.Since(start).Milliseconds())
		delivery.Delivered = false
		_ = db.RecordRepoWebhookDelivery(ctx, delivery)
		return delivery, err
	}

	for k, vv := range reqHeaders {
		for _, v := range vv {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := client.Do(httpReq)
	delivery.DurationMs = int(time.Since(start).Milliseconds())

	if err != nil {
		delivery.ResponseStatusCode = 0
		delivery.ResponseBody = fmt.Sprintf(`{"error": %q}`, err.Error())
		delivery.Delivered = false
	} else {
		defer resp.Body.Close()
		delivery.ResponseStatusCode = resp.StatusCode
		respHeadersJSON, _ := json.Marshal(resp.Header)
		delivery.ResponseHeaders = string(respHeadersJSON)

		// Read up to 4KB of response body
		lr := io.LimitReader(resp.Body, 4096)
		respBodyBytes, _ := io.ReadAll(lr)
		delivery.ResponseBody = string(respBodyBytes)
		delivery.Delivered = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	_ = db.RecordRepoWebhookDelivery(ctx, delivery)
	return delivery, nil
}

// DispatchRepoEvent dispatches a repository event asynchronously to all active webhooks subscribed to that event.
func DispatchRepoEvent(ctx context.Context, db *store.DB, repoID string, event string, payload any) {
	hooks, err := db.ListRepoWebhooks(ctx, repoID)
	if err != nil || len(hooks) == 0 {
		return
	}

	for _, hook := range hooks {
		if !hook.Active {
			continue
		}
		// Check event subscription
		var subscribedEvents []string
		_ = json.Unmarshal([]byte(hook.Events), &subscribedEvents)
		subscribed := false
		for _, e := range subscribedEvents {
			if e == event || e == "*" {
				subscribed = true
				break
			}
		}
		if !subscribed {
			continue
		}

		go func(h *store.RepoWebhook) {
			dispatchCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			del, err := DispatchRepoWebhook(dispatchCtx, db, h, event, payload)
			// If initial attempt failed or returned a 5xx server error, retry once after short backoff
			if (err != nil || (del != nil && !del.Delivered && del.ResponseStatusCode >= 500)) && dispatchCtx.Err() == nil {
				select {
				case <-time.After(500 * time.Millisecond):
					_, _ = DispatchRepoWebhook(dispatchCtx, db, h, event, payload)
				case <-dispatchCtx.Done():
				}
			}
		}(hook)
	}
}
