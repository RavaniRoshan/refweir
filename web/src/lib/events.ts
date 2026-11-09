import { fetchEventSource } from '@microsoft/fetch-event-source';
import type { AgentActivityItem } from '@/lib/api';

export type { AgentActivityItem };

export interface AgentActivityStreamOptions {
  url?: string;
  onSnapshot?: (activities: AgentActivityItem[]) => void;
  onActivity?: (activity: AgentActivityItem) => void;
  onHeartbeat?: (timestamp: string) => void;
  onError?: (err: unknown) => void;
  signal?: AbortSignal;
}

/**
 * Subscribes to real-time agent activities via Server-Sent Events (SSE).
 * Handles initial snapshots, live activities with deduplication, and heartbeats.
 */
export function subscribeAgentActivity(opts: AgentActivityStreamOptions = {}): () => void {
  const url = opts.url ?? '/api/v1/agent/activities/stream';
  const controller = new AbortController();
  const combinedSignal = opts.signal
    ? AbortSignal.any([controller.signal, opts.signal])
    : controller.signal;

  const seenIds = new Set<string>();

  fetchEventSource(url, {
    method: 'GET',
    signal: combinedSignal,
    openWhenHidden: true,
    headers: {
      Accept: 'text/event-stream',
    },
    onmessage(ev) {
      if (!ev.data) return;

      try {
        if (ev.event === 'snapshot') {
          const payload = JSON.parse(ev.data) as { activities?: AgentActivityItem[] };
          const list = payload.activities || [];
          for (const item of list) {
            if (item.id) seenIds.add(item.id);
          }
          opts.onSnapshot?.(list);
          return;
        }

        if (ev.event === 'activity') {
          const item = JSON.parse(ev.data) as AgentActivityItem;
          if (item && item.id) {
            if (seenIds.has(item.id)) {
              return; // deduplicate
            }
            seenIds.add(item.id);
          }
          opts.onActivity?.(item);
          return;
        }

        if (ev.event === 'heartbeat') {
          const payload = JSON.parse(ev.data) as { ts?: string };
          if (payload.ts) {
            opts.onHeartbeat?.(payload.ts);
          }
          return;
        }

        // Default / legacy fallback
        const parsed = JSON.parse(ev.data);
        if (parsed.activities) {
          opts.onSnapshot?.(parsed.activities);
        } else if (parsed.id) {
          if (!seenIds.has(parsed.id)) {
            seenIds.add(parsed.id);
            opts.onActivity?.(parsed);
          }
        }
      } catch {
        // ignore malformed payloads
      }
    },
    onerror(err) {
      opts.onError?.(err);
    },
  });

  return () => controller.abort();
}
