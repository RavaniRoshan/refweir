import { describe, it, expect, vi } from 'vitest';
import { subscribeAgentActivity } from '../events';
import * as fetchEventSourceModule from '@microsoft/fetch-event-source';

vi.mock('@microsoft/fetch-event-source', () => ({
  fetchEventSource: vi.fn(),
}));

describe('subscribeAgentActivity', () => {
  it('dispatches snapshot events to onSnapshot callback', () => {
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
          { id: 'act-1', agentName: 'claude', action: 'clone' },
          { id: 'act-2', agentName: 'cursor', action: 'read' },
        ],
      }),
    });

    expect(onSnapshot).toHaveBeenCalledWith([
      { id: 'act-1', agentName: 'claude', action: 'clone' },
      { id: 'act-2', agentName: 'cursor', action: 'read' },
    ]);
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
      data: JSON.stringify({ id: 'act-1', agentName: 'codex', action: 'test' }),
    });

    expect(onActivity).toHaveBeenCalledTimes(1);
    expect(onActivity).toHaveBeenCalledWith({ id: 'act-1', agentName: 'codex', action: 'test' });

    // Duplicate event with same id act-1 (e.g. from network reconnect replay)
    messageHandler?.({
      event: 'activity',
      data: JSON.stringify({ id: 'act-1', agentName: 'codex', action: 'test' }),
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

  it('aborts controller on unsubscribe cleanup', () => {
    let signalReceived: AbortSignal | undefined;
    vi.mocked(fetchEventSourceModule.fetchEventSource).mockImplementation((_url, opts: any) => {
      signalReceived = opts.signal;
      return Promise.resolve();
    });

    const unsubscribe = subscribeAgentActivity({});
    expect(signalReceived?.aborted).toBe(false);

    unsubscribe();
    expect(signalReceived?.aborted).toBe(true);
  });
});
