// Package events provides domain event publishing, structured audit logging,
// and resilient asynchronous webhook delivery with bounded worker pools.
package events

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"refweir.dev/internal/audit"
	"refweir.dev/internal/security"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

// DomainEvent represents a high-level lifecycle event in Refweir.
type DomainEvent struct {
	RepoID    string         `json:"repo_id"`
	Event     string         `json:"event"`
	Payload   any            `json:"payload"`
	Actor     string         `json:"actor,omitempty"`
	ActorType string         `json:"actor_type,omitempty"`
	Action    string         `json:"action,omitempty"`
	Resource  string         `json:"resource,omitempty"`
	Severity  audit.Severity `json:"severity,omitempty"`
	Details   string         `json:"details,omitempty"`
}

// DispatchRepoWebhook delivers an outbound event payload to a configured repository webhook.
func DispatchRepoWebhook(ctx context.Context, db *store.DB, hook *store.RepoWebhook, event string, payload any) (*store.RepoWebhookDelivery, error) {
	return DispatchRepoWebhookWithGUID(ctx, db, hook, event, payload, "")
}

// DispatchRepoWebhookWithGUID delivers an outbound event payload, preserving delivery GUID across retries for subscriber idempotency.
func DispatchRepoWebhookWithGUID(ctx context.Context, db *store.DB, hook *store.RepoWebhook, event string, payload any, guid string) (*store.RepoWebhookDelivery, error) {
	if hook == nil {
		return nil, fmt.Errorf("webhook is nil")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if guid == "" {
		guid = idgen.New("guid")
	}
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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return security.ValidateWebhookURL(req.URL.String())
		},
		Transport: security.NewSafeTransport(hook.InsecureSSL),
	}

	start := time.Now()
	if err := security.ValidateWebhookURL(hook.URL); err != nil {
		delivery.ResponseStatusCode = 0
		delivery.ResponseBody = fmt.Sprintf(`{"error": %q}`, err.Error())
		delivery.DurationMs = int(time.Since(start).Milliseconds())
		delivery.Delivered = false
		_ = db.RecordRepoWebhookDelivery(ctx, delivery)
		return delivery, err
	}
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

// WebhookDispatcher manages a bounded worker pool for resilient, asynchronous webhook deliveries
// with backpressure handling and retry capabilities.
type WebhookDispatcher struct {
	db           *store.DB
	workers      int
	queue        chan webhookTask
	stopCh       chan struct{}
	wg           sync.WaitGroup
	droppedTotal atomic.Uint64
	started      atomic.Bool
	stopOnce     sync.Once
}

type webhookTask struct {
	hook    *store.RepoWebhook
	event   string
	payload any
}

// NewWebhookDispatcher creates a new dispatcher with bounded concurrency and queue buffer.
func NewWebhookDispatcher(db *store.DB, workers int, queueSize int) *WebhookDispatcher {
	if workers <= 0 {
		workers = 0 // permits test-mode with 0 workers
	}
	if queueSize <= 0 {
		queueSize = 1024
	}
	return &WebhookDispatcher{
		db:      db,
		workers: workers,
		queue:   make(chan webhookTask, queueSize),
		stopCh:  make(chan struct{}),
	}
}

// Start spawns the background worker goroutines.
func (d *WebhookDispatcher) Start() {
	if !d.started.CompareAndSwap(false, true) {
		return
	}
	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go d.workerLoop()
	}
}

// Stop gracefully terminates all worker goroutines after draining remaining jobs.
func (d *WebhookDispatcher) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopCh)
		d.wg.Wait()
		d.started.Store(false)
	})
}

// DroppedTotal returns the count of deliveries dropped due to queue saturation backpressure.
func (d *WebhookDispatcher) DroppedTotal() uint64 {
	return d.droppedTotal.Load()
}

// Dispatch enqueues a webhook delivery task into the bounded queue.
// Returns false if the queue is full and the delivery was dropped under backpressure.
func (d *WebhookDispatcher) Dispatch(ctx context.Context, hook *store.RepoWebhook, event string, payload any) bool {
	if hook == nil {
		return false
	}
	select {
	case d.queue <- webhookTask{hook: hook, event: event, payload: payload}:
		return true
	default:
		d.droppedTotal.Add(1)
		return false
	}
}

func (d *WebhookDispatcher) workerLoop() {
	defer d.wg.Done()
	for {
		select {
		case <-d.stopCh:
			return
		case task := <-d.queue:
			d.deliverTask(task)
		}
	}
}

func (d *WebhookDispatcher) deliverTask(task webhookTask) {
	dispatchCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	del, err := DispatchRepoWebhook(dispatchCtx, d.db, task.hook, task.event, task.payload)
	// If initial attempt failed or returned a 5xx server error, retry once after short backoff preserving GUID for idempotency
	if (err != nil || (del != nil && !del.Delivered && del.ResponseStatusCode >= 500)) && dispatchCtx.Err() == nil {
		deliveryGUID := ""
		if del != nil && del.GUID != "" {
			deliveryGUID = del.GUID
		}
		select {
		case <-time.After(500 * time.Millisecond):
			_, _ = DispatchRepoWebhookWithGUID(dispatchCtx, d.db, task.hook, task.event, task.payload, deliveryGUID)
		case <-dispatchCtx.Done():
		case <-d.stopCh:
			return
		}
	}
}

var (
	defaultDispatcherMu sync.RWMutex
	defaultDispatcher   *WebhookDispatcher
)

func getDefaultDispatcher(db *store.DB) *WebhookDispatcher {
	defaultDispatcherMu.RLock()
	d := defaultDispatcher
	defaultDispatcherMu.RUnlock()
	if d != nil && d.db == db {
		return d
	}

	defaultDispatcherMu.Lock()
	defer defaultDispatcherMu.Unlock()
	if defaultDispatcher == nil || defaultDispatcher.db != db {
		if defaultDispatcher != nil {
			defaultDispatcher.Stop()
		}
		defaultDispatcher = NewWebhookDispatcher(db, 16, 1024)
		defaultDispatcher.Start()
	}
	return defaultDispatcher
}

// DispatchRepoEvent dispatches a repository event asynchronously to all active webhooks subscribed to that event.
func DispatchRepoEvent(ctx context.Context, db *store.DB, repoID string, event string, payload any) {
	hooks, err := db.ListRepoWebhooks(ctx, repoID)
	if err != nil || len(hooks) == 0 {
		return
	}

	dispatcher := getDefaultDispatcher(db)

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

		_ = dispatcher.Dispatch(ctx, hook, event, payload)
	}
}

// Bus unites synchronous audit logging and asynchronous webhook event dissemination.
type Bus struct {
	db         *store.DB
	dispatcher *WebhookDispatcher
}

// NewBus creates an event bus for the given database and dispatcher.
func NewBus(db *store.DB, dispatcher *WebhookDispatcher) *Bus {
	if dispatcher == nil {
		dispatcher = getDefaultDispatcher(db)
	}
	return &Bus{
		db:         db,
		dispatcher: dispatcher,
	}
}

// Publish writes an audit log if specified and fans out to all active repo webhooks asynchronously.
func (b *Bus) Publish(ctx context.Context, evt DomainEvent) {
	if evt.Action != "" {
		sev := evt.Severity
		if sev == "" {
			sev = audit.SeverityInfo
		}
		actType := evt.ActorType
		if actType == "" {
			actType = "user"
		}
		_ = audit.Log(audit.Event{
			Actor:     evt.Actor,
			ActorType: actType,
			Action:    evt.Action,
			Resource:  evt.Resource,
			Severity:  sev,
			Details:   evt.Details,
		})
	}

	if evt.RepoID != "" && evt.Event != "" {
		DispatchRepoEvent(ctx, b.db, evt.RepoID, evt.Event, evt.Payload)
	}
}
