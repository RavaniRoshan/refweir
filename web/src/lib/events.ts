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

const MAX_SEEN_IDS = 1000;
const PRUNE_COUNT = 200;

/**
 * Normalizes an activity object from Go backend JSON, raw event payload,
 * or legacy fixture into a typed AgentActivityItem for frontend UI rendering.
 */
export function normalizeAgentActivity(raw: any): AgentActivityItem {
  if (!raw || typeof raw !== 'object') {
    return {
      id: `act-${Date.now()}`,
      agentName: 'refweir-agent',
      agentType: 'refweir-mcp',
      toolName: 'operation',
      status: 'executing',
      latencyMs: 14,
      timestamp: 'Just now',
      requestPayload: {},
      responseSummary: 'Completed autonomously',
    };
  }

  const rawStatus = raw.status || raw.Status;
  let status: 'success' | 'executing' | 'error' = 'executing';
  if (rawStatus === 'success') {
    status = 'success';
  } else if (rawStatus === 'failed' || rawStatus === 'error') {
    status = 'error';
  } else if (rawStatus === 'executing' || rawStatus === 'in_progress') {
    status = 'executing';
  }

  const rawType = raw.agentType || raw.AgentType;
  const agentType: 'claude-code' | 'cursor' | 'codex' | 'refweir-mcp' =
    rawType === 'claude-code' || rawType === 'cursor' || rawType === 'codex' || rawType === 'refweir-mcp'
      ? rawType
      : 'refweir-mcp';

  const rawDuration = raw.latencyMs ?? raw.durationMs ?? raw.DurationMs;
  const latencyMs = typeof rawDuration === 'number' && !isNaN(rawDuration) ? rawDuration : 14;

  const rawAction = raw.action || raw.Action;
  const rawResource = raw.resource || raw.Resource;
  const requestPayload =
    raw.requestPayload && typeof raw.requestPayload === 'object'
      ? raw.requestPayload
      : rawResource || rawAction
        ? { resource: rawResource, action: rawAction }
        : {};

  const responseSummary =
    raw.details || raw.Details || raw.responseSummary || 'Completed autonomously';

  let formattedTime = 'Just now';
  const rawTs = raw.timestamp || raw.Timestamp;
  if (rawTs) {
    if (typeof rawTs === 'string' && (rawTs.includes('T') || rawTs.includes('-'))) {
      const d = new Date(rawTs);
      if (!isNaN(d.getTime())) {
        formattedTime = d.toLocaleTimeString();
      } else {
        formattedTime = rawTs;
      }
    } else {
      formattedTime = String(rawTs);
    }
  }

  return {
    id: raw.id || raw.ID || `act-${Date.now()}`,
    agentName: raw.agentName || raw.AgentName || 'refweir-agent',
    agentType,
    toolName: raw.toolName || raw.ToolName || rawAction || 'operation',
    status,
    latencyMs,
    timestamp: formattedTime,
    requestPayload,
    responseSummary,
  };
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

  function recordSeen(id: string) {
    if (seenIds.size >= MAX_SEEN_IDS) {
      let count = 0;
      for (const oldId of seenIds) {
        seenIds.delete(oldId);
        count++;
        if (count >= PRUNE_COUNT) break;
      }
    }
    seenIds.add(id);
  }

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
          const payload = JSON.parse(ev.data) as { activities?: any[] };
          const rawList = payload.activities || [];
          const list = rawList.map(normalizeAgentActivity);
          for (const item of list) {
            if (item.id) recordSeen(item.id);
          }
          opts.onSnapshot?.(list);
          return;
        }

        if (ev.event === 'activity') {
          const rawItem = JSON.parse(ev.data);
          const item = normalizeAgentActivity(rawItem);
          if (item && item.id) {
            if (seenIds.has(item.id)) {
              return; // deduplicate
            }
            recordSeen(item.id);
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
          const rawList = parsed.activities;
          const list = rawList.map(normalizeAgentActivity);
          for (const item of list) {
            if (item.id) recordSeen(item.id);
          }
          opts.onSnapshot?.(list);
        } else if (parsed.id || parsed.agentName) {
          const item = normalizeAgentActivity(parsed);
          if (!seenIds.has(item.id)) {
            recordSeen(item.id);
            opts.onActivity?.(item);
          }
        }
      } catch {
        // ignore malformed payloads
      }
    },
    onerror(err) {
      if (combinedSignal.aborted) {
        return;
      }
      opts.onError?.(err);
    },
  });

  return () => controller.abort();
}
