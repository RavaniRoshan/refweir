import { describe, it, expect, vi } from 'vitest';
import { subscribeAgentActivity, normalizeAgentActivity } from '../events';
import * as fetchEventSourceModule from '@microsoft/fetch-event-source';

vi.mock('@microsoft/fetch-event-source', () => ({
  fetchEventSource: vi.fn(),
}));

describe('normalizeAgentActivity', () => {
  it('normalizes raw Go Activity JSON with durationMs and details', () => {
    const rawGo = {
      id: 'act-go-1',
      agentName: 'refweir-reviewer',
      agentType: 'ast-scanner',
      action: 'ast_grep',
      resource: 'internal/app/server.go',
      status: 'success',
      durationMs: 42,
      details: 'Discovered 0 security vulnerabilities',
      timestamp: '2026-11-09T14:30:00Z',
    };

    const normalized = normalizeAgentActivity(rawGo);
    expect(normalized.id).toBe('act-go-1');
    expect(normalized.agentName).toBe('refweir-reviewer');
    expect(normalized.toolName).toBe('ast_grep');
    expect(normalized.status).toBe('success');
    expect(normalized.latencyMs).toBe(42);
    expect(normalized.responseSummary).toBe('Discovered 0 security vulnerabilities');
    expect(normalized.requestPayload).toEqual({
      resource: 'internal/app/server.go',
      action: 'ast_grep',
    });
    expect(typeof normalized.timestamp).toBe('string');
  });

  it('maps Go status failed to error and in_progress to executing', () => {
    expect(normalizeAgentActivity({ id: '1', status: 'failed' }).status).toBe('error');
    expect(normalizeAgentActivity({ id: '2', status: 'in_progress' }).status).toBe('executing');
    expect(normalizeAgentActivity({ id: '3', status: 'success' }).status).toBe('success');
  });

  it('preserves already normalized AgentActivityItem fields', () => {
    const existing = {
      id: 'act-front-1',
      agentName: 'claude',
      agentType: 'refweir-mcp' as const,
      toolName: 'read_file',
      latencyMs: 120,
      status: 'success' as const,
      timestamp: '12:00:00 PM',
      requestPayload: { path: 'README.md' },
      responseSummary: 'File read ok',
    };
    const normalized = normalizeAgentActivity(existing);
    expect(normalized).toEqual(existing);
  });
});

describe('subscribeAgentActivity', () => {
  it('dispatches snapshot events to onSnapshot callback and normalizes items', () => {
    let messageHandler: ((ev: any) => void) | undefined;
    vi.mocked(fetchEventSourceModule.fetchEventSource).mockImplementation((_url, opts: any) => {
      messageHandler = opts.onmessage;
      return Promise.resolve();
    });

    const onSnapshot = vi.fn();
    const onActivity = vi.fn();

    subscribeAgentActivity({
      url: '/api/v1/agent/activities/stream',
      onSnapshot,
      onActivity,
    });

    expect(fetchEventSourceModule.fetchEventSource).toHaveBeenCalled();
    expect(messageHandler).toBeDefined();

    messageHandler?.({
      event: 'snapshot',
      data: JSON.stringify({
        activities: [
          { id: 'act-1', agentName: 'claude', action: 'clone', durationMs: 25 },
          { id: 'act-2', agentName: 'cursor', action: 'read', durationMs: 10 },
        ],
      }),
    });

    expect(onSnapshot).toHaveBeenCalledTimes(1);
    const items = onSnapshot.mock.calls[0][0];
    expect(items).toHaveLength(2);
    expect(items[0].id).toBe('act-1');
    expect(items[0].toolName).toBe('clone');
    expect(items[0].latencyMs).toBe(25);
    expect(items[1].id).toBe('act-2');
    expect(items[1].toolName).toBe('read');
    expect(items[1].latencyMs).toBe(10);
  });

  it('dispatches live activity events to onActivity and dedupes repeated IDs', () => {
    let messageHandler: ((ev: any) => void) | undefined;
    vi.mocked(fetchEventSourceModule.fetchEventSource).mockImplementation((_url, opts: any) => {
      messageHandler = opts.onmessage;
      return Promise.resolve();
    });

    const onActivity = vi.fn();

    subscribeAgentActivity({
      onActivity,
    });

    // First event with id act-1
    messageHandler?.({
      event: 'activity',
      data: JSON.stringify({ id: 'act-1', agentName: 'codex', action: 'test', durationMs: 50 }),
    });

    expect(onActivity).toHaveBeenCalledTimes(1);
    expect(onActivity.mock.calls[0][0].id).toBe('act-1');
    expect(onActivity.mock.calls[0][0].latencyMs).toBe(50);

    // Duplicate event with same id act-1 (e.g. from network reconnect replay)
    messageHandler?.({
      event: 'activity',
      data: JSON.stringify({ id: 'act-1', agentName: 'codex', action: 'test', durationMs: 50 }),
    });

    // Should NOT call onActivity again for duplicate
    expect(onActivity).toHaveBeenCalledTimes(1);

    // New event with id act-2
    messageHandler?.({
      event: 'activity',
      data: JSON.stringify({ id: 'act-2', agentName: 'codex', action: 'push' }),
    });

    expect(onActivity).toHaveBeenCalledTimes(2);
  });

  it('dispatches heartbeat event to onHeartbeat', () => {
    let messageHandler: ((ev: any) => void) | undefined;
    vi.mocked(fetchEventSourceModule.fetchEventSource).mockImplementation((_url, opts: any) => {
      messageHandler = opts.onmessage;
      return Promise.resolve();
    });

    const onHeartbeat = vi.fn();

    subscribeAgentActivity({
      onHeartbeat,
    });

    const ts = new Date().toISOString();
    messageHandler?.({
      event: 'heartbeat',
      data: JSON.stringify({ ts }),
    });

    expect(onHeartbeat).toHaveBeenCalledWith(ts);
  });

  it('aborts controller on unsubscribe cleanup and ignores error after abort', () => {
    let signalReceived: AbortSignal | undefined;
    let errorHandler: ((err: any) => void) | undefined;
    vi.mocked(fetchEventSourceModule.fetchEventSource).mockImplementation((_url, opts: any) => {
      signalReceived = opts.signal;
      errorHandler = opts.onerror;
      return Promise.resolve();
    });

    const onError = vi.fn();
    const unsubscribe = subscribeAgentActivity({ onError });
    expect(signalReceived?.aborted).toBe(false);

    unsubscribe();
    expect(signalReceived?.aborted).toBe(true);

    // Simulate fetchEventSource raising error after abort
    errorHandler?.(new Error('AbortError'));
    expect(onError).not.toHaveBeenCalled();
  });
});
