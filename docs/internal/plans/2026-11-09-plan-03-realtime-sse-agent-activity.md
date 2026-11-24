# Real-Time SSE Agent Activity Stream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement zero-latency real-time agent activity streaming using Server-Sent Events (SSE) with Go backend pub/sub broadcast and `@microsoft/fetch-event-source` frontend client with deduplication, heartbeat, and polling fallback.

**Architecture:** `agentactivity.Hub` is upgraded to a thread-safe pub/sub broadcaster (`Subscribe() (<-chan Activity, func())`). `StreamHandler` emits an initial snapshot (`event: snapshot`), streams new activities immediately as they occur (`event: activity`), and sends heartbeats every 15s (`event: heartbeat`). On the frontend, `web/src/lib/events.ts` manages connection lifecycle, token auth, deduplication, and fallback to polling. `AgentConsoleView.tsx` and `GlobalHomeView.tsx` consume the live stream with immediate visual updates.

**Tech Stack:** Go 1.24, pure-Go HTTP standard library, React 19, TypeScript 6, `@microsoft/fetch-event-source`, Vitest, Testing Library.

**Spec:** `plan.md` (Phase 4: §4.3 SSE for agent activity).

## Global Constraints

- Pure-Go backend invariants: zero cgo, zero host OpenSSH/Git dependencies, pure-Go SQLite WAL.
- Zero mock data in live mode (`isDemoModeActive() === false`).
- Streaming headers: `Cache-Control: no-cache`, `X-Accel-Buffering: no`, `Connection: keep-alive`.
- Bounded memory buffers: slow SSE subscribers must never block `Record()` or leak memory.
- Backward compatibility: REST endpoint `GET /api/v1/agent/activities` continues to return JSON snapshot.
- All 47+ existing Vitest test files and all Go packages must remain 100% green.

## Review Focus

1. Pub/sub fanout non-blocking guarantee: a blocked or stalled SSE client connection cannot delay `Record()` calls by other agents.
2. Unsubscribe cleanup: when client closes HTTP request context, the subscriber channel is closed and removed from `Hub.subscribers`.
3. Event schema alignment: Go `Activity` JSON matches TypeScript `AgentActivityItem` exactly (`id`, `timestamp`, `agentName`, `agentType`, `action`, `toolName`, `resource`, `status`, `durationMs`, `details`).
4. Reconnect & deduplication: after reconnection or network drop, duplicate activities with identical `id` are discarded by the client.
5. Fallback path: in offline demo mode or when SSE fails, the UI falls back cleanly to periodic polling without throwing unhandled exceptions.

---

### Task 1: Hub Pub/Sub Broadcaster & Immediate Streaming in Go

**Files:**
- Modify: `internal/agentactivity/hub.go`
- Modify: `internal/agentactivity/stream.go`
- Test: `internal/agentactivity/stream_test.go`
- Test: `internal/agentactivity/hub_test.go`

**Interfaces:**
- Consumes: `Activity` struct, `idgen.New("act")`.
- Produces: `h.Subscribe(bufferSize int) (ch <-chan Activity, cancel func())`.
- Produces: `StreamHandler(h *Hub) http.HandlerFunc` supporting snapshot, live activity events, and heartbeats.

- [ ] **Step 1: Write failing tests for Hub pub/sub and StreamHandler in `internal/agentactivity/stream_test.go`**

Create `internal/agentactivity/stream_test.go` testing:
1. `Subscribe()` receives activities recorded via `Record()` immediately.
2. `cancel()` unsubscribes and cleans up channel.
3. `StreamHandler` sends snapshot on connection, followed immediately by streamed activity event when `Record()` is called.
4. Slow subscriber buffer overflow drops oldest or skips without blocking `Record()`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestStreamHandler ./internal/agentactivity`
Expected: FAIL because `Subscribe()` is not yet implemented on `Hub`.

- [ ] **Step 3: Implement Pub/Sub in `internal/agentactivity/hub.go` and `stream.go`**

- In `Hub`, add `subscribers map[chan Activity]struct{}` guarded by `sync.RWMutex`.
- In `Record(act Activity)`:
  - Append to ring buffer.
  - Broadcast to all subscribers in `subscribers` map with non-blocking send (`select { case ch <- act: default: }`).
- In `Subscribe(bufferSize int)`:
  - Create buffered channel.
  - Add to `subscribers`.
  - Return receive-only channel and cancel closure.
- In `StreamHandler`:
  - Write SSE headers (`text/event-stream`, `no-cache`, `X-Accel-Buffering: no`).
  - Send initial `event: snapshot\ndata: <JSON array>\n\n`.
  - Subscribe to hub.
  - Listen in `select` loop:
    - `case <-r.Context().Done()`: return and cancel subscription.
    - `case act := <-subCh`: send `event: activity\ndata: <JSON object>\n\n`, flush.
    - `case <-ticker.C`: send `event: heartbeat\ndata: {"ts":"..."}\n\n`, flush.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/agentactivity`
Expected: PASS (all tests green).

- [ ] **Step 5: Commit**

```bash
git add internal/agentactivity
git commit -m "feat(agentactivity): implement zero-latency pub/sub streaming and heartbeats"
```

---

### Task 2: Robust Client Stream & Deduplication in `web/src/lib/events.ts`

**Files:**
- Modify: `web/src/lib/events.ts`
- Test: `web/src/lib/__tests__/events.test.ts`

**Interfaces:**
- Consumes: `AgentActivityItem` from `@/lib/api`.
- Produces: `subscribeAgentActivity(opts: SSEOptions): () => void`.

- [ ] **Step 1: Write failing tests for `events.ts` in `web/src/lib/__tests__/events.test.ts`**

Create `web/src/lib/__tests__/events.test.ts` verifying:
1. Parses `event: snapshot` and emits initial activities.
2. Parses `event: activity` and emits live activity item.
3. Dedupes activities with duplicate IDs.
4. Invokes `onError` and supports abort signal for disconnection.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run src/lib/__tests__/events.test.ts`
Expected: FAIL because `events.ts` types and event handling do not match snapshot/activity protocols.

- [ ] **Step 3: Implement client streaming in `web/src/lib/events.ts`**

- Align TypeScript interfaces with Go backend (`AgentActivityItem`).
- Support handling `event: 'snapshot'`, `event: 'activity'`, and `event: 'heartbeat'`.
- Maintain a local Set of seen IDs (`seenIds: Set<string>`) to deduplicate across reconnections.
- Handle fallback polling when SSE connection fails or times out.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx vitest run src/lib/__tests__/events.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/lib/events.ts web/src/lib/__tests__/events.test.ts
git commit -m "feat(web): update SSE event subscriber with snapshot, live activity, and deduplication"
```

---

### Task 3: Wire Real-Time Stream into `AgentConsoleView.tsx` & `GlobalHomeView.tsx`

**Files:**
- Modify: `web/src/components/views/AgentConsoleView.tsx`
- Modify: `web/src/components/views/GlobalHomeView.tsx`
- Test: `web/src/components/views/__tests__/AgentConsoleView.test.tsx`
- Test: `web/src/components/views/__tests__/GlobalHomeView.test.tsx`

**Interfaces:**
- Consumes: `subscribeAgentActivity` from `@/lib/events`.
- Produces: Instant activity prepend in console and home activity feeds without waiting 4s.

- [ ] **Step 1: Write failing test in `AgentConsoleView.test.tsx` verifying SSE live subscription**

Update `web/src/components/views/__tests__/AgentConsoleView.test.tsx` to verify:
1. `subscribeAgentActivity` is called on mount.
2. Incoming live activity event prepends immediately to the activities list.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run src/components/views/__tests__/AgentConsoleView.test.tsx`
Expected: FAIL because `AgentConsoleView` only uses `getAgentActivities()` polling.

- [ ] **Step 3: Wire `subscribeAgentActivity` into `AgentConsoleView` and `GlobalHomeView`**

- In `AgentConsoleView.tsx`:
  - Call `subscribeAgentActivity` on mount.
  - On snapshot: set activities.
  - On live activity: prepend item to `activities` (capping at 100 entries).
  - Clean up subscription on unmount.
  - Fall back to polling if demo mode is active or SSE errors.
- In `GlobalHomeView.tsx`:
  - Subscribe to real-time agent activities for the cluster telemetry feed.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx vitest run src/components/views/__tests__/AgentConsoleView.test.tsx src/components/views/__tests__/GlobalHomeView.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/views/AgentConsoleView.tsx web/src/components/views/GlobalHomeView.tsx web/src/components/views/__tests__/AgentConsoleView.test.tsx web/src/components/views/__tests__/GlobalHomeView.test.tsx
git commit -m "feat(web): wire live SSE agent activity streaming to AgentConsoleView and GlobalHomeView"
```

---

### Task 4: Documentation & Reverse Proxy Configuration Guide

**Files:**
- Modify: `site/content/docs/architecture.mdx` or create `docs/proxy/sse.md`
- Test: `site/scripts/check-links.mjs`, `npm run verify`

- [ ] **Step 1: Document SSE proxy requirements**

Add configuration examples for Nginx (`proxy_buffering off`, `proxy_read_timeout 3600s`) and Cloudflare (`SSE buffer bypass`, `X-Accel-Buffering: no`).

- [ ] **Step 2: Verify site links and claims**

Run: `cd site && npm run verify`
Expected: PASS with 0 broken links.

- [ ] **Step 3: Commit**

```bash
git add site/
git commit -m "docs: document SSE proxy buffering and stream configuration"
```

---

### Task 5: End-to-End Verification, Capability Counts & Integration Pass

**Files:**
- Verify: Full repo

- [ ] **Step 1: Run full frontend Vitest test suite**

Run: `cd web && npm test`
Expected: 100% pass across all test files (≥48 test files, ≥185 tests).

- [ ] **Step 2: Run web production build**

Run: `cd web && npm run build`
Expected: PASS with 0 errors.

- [ ] **Step 3: Update capability counts**

Run: `go run ./internal/tools/gencounts && git diff --exit-code docs/generated/capability_counts.md`
If diff: `git add docs/generated/capability_counts.md && git commit -m "docs: update capability counts for sse test files"`

- [ ] **Step 4: Run Go backend test suite**

Run: `go test -v ./...`
Expected: PASS across all packages.

- [ ] **Step 5: Merge feature branch**

```bash
git checkout main
git merge --ff-only feat/sse-agent-activity
git branch -d feat/sse-agent-activity
```

---

## Self-Review

1. **Spec coverage:** Implements all requirements of `plan.md` §4.3 (GET stream, 15-30s heartbeats, `@microsoft/fetch-event-source`, reconnect, deduplication, proxy documentation).
2. **Step scan:** Every step contains exact files, test commands, expected outputs, and commit messages.
3. **Type consistency:** Field names between Go `Activity` and TypeScript `AgentActivityItem` are unified.
4. **Review Focus:**
   - Non-blocking channel broadcast with `select { case ch <- act: default: }`.
   - Context cancellation unregisters listener to prevent memory leaks.
   - Snapshot + live activity event separation.
   - Deduplication via `Set<string>`.
5. **Proportion:** Clear, modular, and testable.
