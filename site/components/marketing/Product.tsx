'use client';

import { useState } from 'react';
import { ProductPreview } from './ProductPreview';
import { MergeGate } from './MergeGate';

const LEASE_TOKEN = 'lease_hmac_9a2f7c01e8b44d12';

const STREAM_ITEMS = [
  { agent: 'claude-code', action: 'lease_request', target: 'refs/heads/agent/fix-auth', time: '12s ago', status: 'granted' },
  { agent: 'cursor-agent', action: 'git_receive_pack', target: 'refs/heads/agent/nav-polish', time: '45s ago', status: 'verified' },
  { agent: 'agent-reviewer', action: 'check_run:ast_scan', target: 'PR #142', time: '1m ago', status: 'passed' },
];

export function Product() {
  const [activeProto, setActiveProto] = useState<'http' | 'ssh'>('http');
  const [copiedToken, setCopiedToken] = useState(false);

  const copyToken = async () => {
    try {
      await navigator.clipboard.writeText(LEASE_TOKEN);
      setCopiedToken(true);
      setTimeout(() => setCopiedToken(false), 2000);
    } catch {
      // fallback
    }
  };

  return (
    <section id="product" className="relative border-b border-line bg-canvas py-20 lg:py-28 scroll-mt-20">
      <div className="container-site">
        {/* Section Header */}
        <div className="max-w-2xl">
          <p className="eyebrow text-patina font-semibold uppercase tracking-widest text-xs">
            02 — Forge Architecture
          </p>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-text-1 sm:text-5xl">
            Policy at every step, <span className="text-patina">not just at the edge.</span>
          </h2>
          <p className="mt-4 text-base leading-relaxed text-text-2">
            Every layer bounds what an autonomous coding agent can execute. Security is enforced by the pure-Go server kernel — not conventions or advisory files.
          </p>
        </div>

        {/* Live Product UI Showcase */}
        <div className="mt-12">
          <ProductPreview />
        </div>

        {/* Bento Grid Feature Cards */}
        <div className="mt-12 grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-12">
          
          {/* Bento Card 1: Capability Leases (Span 7) */}
          <div className="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-line bg-slab p-6 transition-all duration-200 hover:border-patina/50 lg:col-span-7">
            <div>
              <div className="flex items-center justify-between">
                <span className="mono rounded-md border border-patina/30 bg-patina/10 px-2.5 py-1 text-xs font-semibold text-patina uppercase">
                  pkg/agentlease
                </span>
                <span className="mono text-xs text-text-3">HMAC-SHA256 Signed</span>
              </div>

              <h3 className="mt-4 text-xl font-bold text-text-1">
                Ephemeral Capability Leases
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-text-2">
                Agents never get static passwords or long-lived tokens. They receive cryptographically signed capability leases strictly scoped to <code className="mono rounded bg-raised px-1.5 py-0.5 text-xs text-patina">refs/heads/agent/*</code> with automatic 1-hour expiration.
              </p>

              {/* Interactive Token Simulation Pill */}
              <div className="mt-5 rounded-lg border border-line bg-raised p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                    <code className="mono text-xs font-semibold text-text-1">{LEASE_TOKEN}</code>
                  </div>
                  <button
                    type="button"
                    onClick={copyToken}
                    className="mono rounded bg-slab px-2.5 py-1 text-[11px] font-medium text-text-2 transition-colors hover:bg-patina hover:text-white"
                  >
                    {copiedToken ? 'Copied' : 'Copy'}
                  </button>
                </div>
                <div className="mt-3 flex items-center justify-between text-xs text-text-3">
                  <span>Scope: <span className="text-text-1 font-mono">refs/heads/agent/*</span></span>
                  <span className="text-emerald-500 font-mono">TTL: 58m remaining</span>
                </div>
              </div>
            </div>

            <div className="mt-6 flex items-center justify-between border-t border-line/80 pt-4 text-xs text-text-3">
              <span>Standing write access to main: <strong className="text-red-400">BLOCKED</strong></span>
              <a
                href="https://github.com/RavaniRoshan/refweir/blob/main/pkg/agentlease/lease.go"
                className="text-patina font-medium hover:underline"
                target="_blank"
                rel="noreferrer"
              >
                Inspect pkg/agentlease ↗
              </a>
            </div>
          </div>

          {/* Bento Card 2: Dual Native Transports (Span 5) */}
          <div className="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-line bg-slab p-6 transition-all duration-200 hover:border-patina/50 lg:col-span-5">
            <div>
              <div className="flex items-center justify-between">
                <span className="mono rounded-md border border-line bg-raised px-2.5 py-1 text-xs font-semibold text-text-2 uppercase">
                  Zero CGO
                </span>
                <span className="mono text-xs text-emerald-400">0 Host Daemons</span>
              </div>

              <h3 className="mt-4 text-xl font-bold text-text-1">
                Dual Git Transports
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-text-2">
                Pure-Go Smart-HTTP on <code className="mono text-xs font-semibold text-text-1">:8080</code> and an embedded standalone SSH daemon on <code className="mono text-xs font-semibold text-text-1">:2222</code>. No host OpenSSH required.
              </p>

              {/* Protocol Switcher */}
              <div className="mt-5">
                <div className="flex rounded-lg border border-line bg-raised p-1">
                  <button
                    type="button"
                    onClick={() => setActiveProto('http')}
                    className={`flex-1 rounded-md py-1.5 text-xs font-semibold transition-all ${
                      activeProto === 'http'
                        ? 'bg-slab text-text-1 shadow-xs'
                        : 'text-text-3 hover:text-text-1'
                    }`}
                  >
                    Smart-HTTP (:8080)
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveProto('ssh')}
                    className={`flex-1 rounded-md py-1.5 text-xs font-semibold transition-all ${
                      activeProto === 'ssh'
                        ? 'bg-slab text-text-1 shadow-xs'
                        : 'text-text-3 hover:text-text-1'
                    }`}
                  >
                    Pure-Go SSH (:2222)
                  </button>
                </div>

                <div className="mt-3 rounded-lg border border-line/80 bg-raised p-3">
                  <code className="mono block truncate text-xs text-text-2">
                    {activeProto === 'http'
                      ? 'git clone http://localhost:8080/acme/repo.git'
                      : 'git clone ssh://git@localhost:2222/acme/repo.git'}
                  </code>
                </div>
              </div>
            </div>

            <div className="mt-6 border-t border-line/80 pt-4 text-xs text-text-3">
              Deploy keys strictly isolated by repository UUID.
            </div>
          </div>

          {/* Bento Card 3: Autonomous PR Reviewer & Check Runs (Span 5) */}
          <div className="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-line bg-slab p-6 transition-all duration-200 hover:border-patina/50 lg:col-span-5">
            <div>
              <div className="flex items-center justify-between">
                <span className="mono rounded-md border border-line bg-raised px-2.5 py-1 text-xs font-semibold text-text-2 uppercase">
                  internal/agentreviewer
                </span>
                <span className="mono text-xs text-text-3">AST Scanner</span>
              </div>

              <h3 className="mt-4 text-xl font-bold text-text-1">
                Autonomous Review Pipeline
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-text-2">
                Every pushed diff is automatically parsed for leaked secret patterns, unit test coverage, and code symbol impacts before check runs report conclusions.
              </p>

              {/* Simulated Check Run Badges */}
              <div className="mt-5 space-y-2.5">
                <div className="flex items-center justify-between rounded-lg border border-line/80 bg-raised px-3 py-2 text-xs">
                  <span className="text-text-2">Secret Pattern Scanner</span>
                  <span className="mono font-semibold text-emerald-400">Passed (Clean ✓)</span>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-line/80 bg-raised px-3 py-2 text-xs">
                  <span className="text-text-2">AST Semantic Symbol Index</span>
                  <span className="mono font-semibold text-emerald-400">Indexed (100% ✓)</span>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-line/80 bg-raised px-3 py-2 text-xs">
                  <span className="text-text-2">SIGSPEC &amp; GPG Verification</span>
                  <span className="mono font-semibold text-emerald-400">Verified ✓</span>
                </div>
              </div>
            </div>

            <div className="mt-6 border-t border-line/80 pt-4 text-xs text-text-3">
              Rich markdown diagnostics reported directly to PR threads.
            </div>
          </div>

          {/* Bento Card 4: Interactive Merge Gate (Span 7) */}
          <div className="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-line bg-slab p-6 transition-all duration-200 hover:border-patina/50 lg:col-span-7">
            <MergeGate />
          </div>

          {/* Bento Card 5: Real-Time SSE Activity Stream (Span 6) */}
          <div className="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-line bg-slab p-6 transition-all duration-200 hover:border-patina/50 lg:col-span-6">
            <div>
              <div className="flex items-center justify-between">
                <span className="mono rounded-md border border-line bg-raised px-2.5 py-1 text-xs font-semibold text-text-2 uppercase">
                  Zero Latency SSE
                </span>
                <span className="mono flex items-center gap-1.5 text-xs text-emerald-400">
                  <span className="h-2 w-2 rounded-full bg-emerald-500 animate-ping" />
                  Live Stream
                </span>
              </div>

              <h3 className="mt-4 text-xl font-bold text-text-1">
                Real-Time Agent Telemetry
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-text-2">
                Server-Sent Events broadcast autonomous agent tool calls, PR reviews, and merge conclusions immediately to the frontend ring buffer.
              </p>

              <div className="mt-5 space-y-2">
                {STREAM_ITEMS.map((item, idx) => (
                  <div key={idx} className="flex items-center justify-between rounded-lg border border-line/70 bg-raised px-3.5 py-2.5 text-xs">
                    <div className="flex items-center gap-2.5">
                      <span className="mono font-semibold text-text-1">{item.agent}</span>
                      <span className="mono text-text-3">→</span>
                      <code className="mono text-patina">{item.action}</code>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="mono text-[11px] text-text-3">{item.time}</span>
                      <span className="mono rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-400 uppercase">
                        {item.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="mt-6 border-t border-line/80 pt-4 text-xs text-text-3">
              Bounded in-memory ring buffer with automatic client deduplication.
            </div>
          </div>

          {/* Bento Card 6: Single Hermetic Binary & SQLite WAL (Span 6) */}
          <div className="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-line bg-slab p-6 transition-all duration-200 hover:border-patina/50 lg:col-span-6">
            <div>
              <div className="flex items-center justify-between">
                <span className="mono rounded-md border border-line bg-raised px-2.5 py-1 text-xs font-semibold text-text-2 uppercase">
                  SQLite WAL Engine
                </span>
                <span className="mono text-xs text-emerald-400">~9,880 tx/sec</span>
              </div>

              <h3 className="mt-4 text-xl font-bold text-text-1">
                Hermetic Single-Binary Daemon
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-text-2">
                Smart-HTTP, embedded SSH server, native MCP server, and embedded React 19 UI compile into a single static binary. Zero Redis, zero PostgreSQL, zero external queues.
              </p>

              <div className="mt-5 grid grid-cols-2 gap-3 text-center">
                <div className="rounded-lg border border-line bg-raised p-3.5">
                  <div className="text-2xl font-bold text-text-1">101 µs</div>
                  <div className="mono mt-1 text-[11px] text-text-3 uppercase">Write Transaction</div>
                </div>
                <div className="rounded-lg border border-line bg-raised p-3.5">
                  <div className="text-2xl font-bold text-text-1">14.9 µs</div>
                  <div className="mono mt-1 text-[11px] text-text-3 uppercase">Index Read Lookup</div>
                </div>
              </div>
            </div>

            <div className="mt-6 border-t border-line/80 pt-4 text-xs text-text-3">
              CGO_ENABLED=0 across linux/amd64, linux/arm64, darwin/arm64, windows/amd64.
            </div>
          </div>

        </div>
      </div>
    </section>
  );
}
