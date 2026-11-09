import { fetchEventSource } from '@microsoft/fetch-event-source';

export interface AgentActivityEvent {
  type: string;
  timestamp: string;
  actor: string;
  actorType: string;
  action: string;
  resource: string;
  severity: string;
  details?: string;
  metadata?: Record<string, unknown>;
}

export interface SSEOptions {
  url: string;
  onEvent: (event: AgentActivityEvent) => void;
  onError?: (err: unknown) => void;
  signal?: AbortSignal;
}

export function subscribeAgentActivity(opts: SSEOptions): () => void {
  const controller = new AbortController();
  const combinedSignal = opts.signal
    ? AbortSignal.any([controller.signal, opts.signal])
    : controller.signal;

  fetchEventSource(opts.url, {
    method: 'GET',
    signal: combinedSignal,
    openWhenHidden: true,
    onmessage(ev) {
      if (!ev.data) return;
      try {
        const parsed = JSON.parse(ev.data) as AgentActivityEvent;
        opts.onEvent(parsed);
      } catch {
        // ignore malformed events
      }
    },
    onerror(err) {
      opts.onError?.(err);
      // fetch-event-source auto-retries; throw to stop
    },
  });

  return () => controller.abort();
}
EOF