package app

import (
	"context"

	"refweir.dev/internal/events"
	"refweir.dev/internal/store"
)

// WebhookDispatcher aliases events.WebhookDispatcher for backwards compatibility within internal/app.
type WebhookDispatcher = events.WebhookDispatcher

// NewWebhookDispatcher creates a new dispatcher with bounded concurrency and queue buffer.
func NewWebhookDispatcher(db *store.DB, workers int, queueSize int) *WebhookDispatcher {
	return events.NewWebhookDispatcher(db, workers, queueSize)
}

// DispatchRepoWebhook delivers an outbound event payload to a configured repository webhook.
func DispatchRepoWebhook(ctx context.Context, db *store.DB, hook *store.RepoWebhook, event string, payload any) (*store.RepoWebhookDelivery, error) {
	return events.DispatchRepoWebhook(ctx, db, hook, event, payload)
}

// DispatchRepoWebhookWithGUID delivers an outbound event payload, preserving delivery GUID across retries for subscriber idempotency.
func DispatchRepoWebhookWithGUID(ctx context.Context, db *store.DB, hook *store.RepoWebhook, event string, payload any, guid string) (*store.RepoWebhookDelivery, error) {
	return events.DispatchRepoWebhookWithGUID(ctx, db, hook, event, payload, guid)
}

// DispatchRepoEvent dispatches a repository event asynchronously to all active webhooks subscribed to that event.
func DispatchRepoEvent(ctx context.Context, db *store.DB, repoID string, event string, payload any) {
	events.DispatchRepoEvent(ctx, db, repoID, event, payload)
}
