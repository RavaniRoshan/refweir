'use client';

import { useState } from 'react';
import { asset } from '@/lib/asset';
import { IconCheck, IconShield, IconCopy, IconArrowUpRight } from './icons';

interface Scenario {
  id: string;
  name: string;
  ref: string;
  status: 'passed' | 'rejected' | 'expired';
  httpStatus: string;
  desc: string;
  steps: Array<{
    text: string;
    tone: 'cmd' | 'info' | 'success' | 'danger' | 'dim';
  }>;
  auditEntry: {
    event: string;
    ref: string;
    actor: string;
    decision: 'ALLOW' | 'DENY';
    reason: string;
  };
}

const SCENARIOS: Scenario[] = [
  {
    id: 'valid',
    name: 'Scoped Agent Push',
    ref: 'refs/heads/agent/feature-login',
    status: 'passed',
    httpStatus: '200 OK',
    desc: 'Agent presents HMAC lease scoped to refs/heads/agent/*. Pushes valid commits within lease TTL.',
    steps: [
      { text: '$ git push origin agent/feature-login', tone: 'cmd' },
      { text: 'remote: Verifying lease HMAC-SHA256...', tone: 'info' },
      { text: 'remote: Lease valid (ttl: 12m remaining, actor: claude-code-1)', tone: 'info' },
      { text: 'remote: Pre-receive hook: checking target ref refs/heads/agent/feature-login...', tone: 'info' },
      { text: 'remote: Ref-prefix matched pattern: refs/heads/agent/*', tone: 'success' },
      { text: 'remote: Secret scan: clean (0 hardcoded secrets detected)', tone: 'success' },
      { text: 'To http://127.0.0.1:8080/acme/forge.git\n * [new branch] agent/feature-login -> agent/feature-login', tone: 'success' },
      { text: 'remote: Audit log: written event git_push_allow (lease: lse_994a2b)', tone: 'dim' },
    ],
    auditEntry: {
      event: 'git.receive-pack.push',
      ref: 'refs/heads/agent/feature-login',
      actor: 'agent:claude-code-1',
      decision: 'ALLOW',
      reason: 'scope matched refs/heads/agent/* and lease active',
    },
  },
  {
    id: 'overreach',
    name: 'Over-reach Violation',
    ref: 'refs/heads/main',
    status: 'rejected',
    httpStatus: '403 Forbidden (GH007)',
    desc: 'Compromised or runaway agent attempts to bypass PR review and force-push directly to protected main.',
    steps: [
      { text: '$ git push origin HEAD:main --force', tone: 'cmd' },
      { text: 'remote: Verifying lease HMAC-SHA256...', tone: 'info' },
      { text: 'remote: Lease valid (ttl: 14m remaining, actor: cline-agent)', tone: 'info' },
      { text: 'remote: Pre-receive hook: checking target ref refs/heads/main...', tone: 'info' },
      { text: 'remote: ! [remote rejected] GH007: Agent lease cannot write to protected branch refs/heads/main', tone: 'danger' },
      { text: 'remote: Agent capability leases are strictly constrained to refs/heads/agent/*', tone: 'danger' },
      { text: 'error: failed to push some refs to http://127.0.0.1:8080/acme/forge.git', tone: 'danger' },
      { text: 'remote: Audit log: written security violation DENY (actor: cline-agent, code: GH007)', tone: 'dim' },
    ],
    auditEntry: {
      event: 'security.boundary_violation',
      ref: 'refs/heads/main',
      actor: 'agent:cline-agent',
      decision: 'DENY',
      reason: 'GH007: lease restricted to refs/heads/agent/*',
    },
  },
  {
    id: 'expired',
    name: 'Expired Lease',
    ref: 'refs/heads/agent/patch-v2',
    status: 'expired',
    httpStatus: '401 Unauthorized',
    desc: 'Agent presents lease token whose 15-minute time-to-live has expired. Daemon terminates connection immediately.',
    steps: [
      { text: '$ git push origin agent/patch-v2', tone: 'cmd' },
      { text: 'remote: Verifying lease HMAC-SHA256...', tone: 'info' },
      { text: 'fatal: Authentication failed: lease lse_812c has expired (expired 3m ago)', tone: 'danger' },
      { text: 'fatal: Request a fresh lease via POST /api/v1/leases or MCP lease_request', tone: 'danger' },
      { text: 'remote: Audit log: written auth failure DENY (expired_lease)', tone: 'dim' },
    ],
    auditEntry: {
      event: 'auth.lease_verification_failure',
      ref: 'refs/heads/agent/patch-v2',
      actor: 'unknown (expired token)',
      decision: 'DENY',
      reason: 'HMAC signature expired; zero standing access',
    },
  },
];

export function ContainmentDemo() {
  const [activeId, setActiveId] = useState('valid');
  const scenario = SCENARIOS.find((s) => s.id === activeId) || SCENARIOS[0];

  return (
    <section id="containment" className="relative border-b border-line scroll-mt-20">
      <div className="container-site grid gap-10 py-16 lg:grid-cols-12 lg:py-24">
        {/* Left Column: Context & Interactive Scenario Switcher */}
        <div className="lg:col-span-5 flex flex-col justify-between">
          <div>
            <p className="eyebrow">04 — containment engine</p>
            <h2 className="mt-4 text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
              <span className="text-patina">Contain</span> every agent.<br />
              Prove every merge.
            </h2>
            <p className="mt-5 text-[15px] leading-relaxed text-text-2">
              An AI agent should never hold long-lived credentials to <code className="font-mono text-[13px] text-text-1">main</code>.
              Refweir grants signed, short-lived HMAC capability leases scoped exclusively to{' '}
              <code className="font-mono text-[13px] text-patina">refs/heads/agent/*</code>.
              The pre-receive gate rejects over-reach at the Git boundary — not in client UI.
            </p>

            {/* Scenario Selector Pills */}
            <div className="mt-8">
              <span className="mono text-[10px] uppercase tracking-[0.16em] text-text-3">
                Select Simulation Scenario:
              </span>
              <div className="mt-2.5 space-y-2">
                {SCENARIOS.map((s) => {
                  const isSelected = s.id === activeId;
                  return (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => setActiveId(s.id)}
                      className={`w-full text-left p-3.5 rounded-md border transition-all text-xs flex items-center justify-between ${
                        isSelected
                          ? 'border-patina bg-patina/10 text-text-1 shadow-sm'
                          : 'border-line bg-slab hover:border-line-strong text-text-2 hover:text-text-1'
                      }`}
                    >
                      <div className="flex items-center gap-2.5">
                        <span
                          className={`w-2 h-2 rounded-full ${
                            s.status === 'passed'
                              ? 'bg-emerald-400'
                              : s.status === 'rejected'
                              ? 'bg-rose-500'
                              : 'bg-amber-400'
                          }`}
                        />
                        <span className="font-medium text-[13px]">{s.name}</span>
                      </div>
                      <span className="mono text-[10.5px] text-text-3">{s.httpStatus}</span>
                    </button>
                  );
                })}
              </div>
            </div>

            <p className="mt-4 text-xs text-text-3 leading-relaxed">
              {scenario.desc}
            </p>
          </div>

          <div className="mt-8 flex flex-wrap items-center gap-3">
            <a
              href="https://github.com/RavaniRoshan/refweir/blob/main/pkg/agentlease/lease.go"
              target="_blank"
              rel="noreferrer"
              className="btn btn-signal"
            >
              Inspect Leases (Go)
              <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
            </a>
            <a
              href={asset('/docs/tutorials/first-agent/')}
              className="btn"
            >
              Agent Tutorial
              <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
            </a>
          </div>
        </div>

        {/* Right Column: Motion Terminal & Live Policy Gate */}
        <div className="lg:col-span-7 flex flex-col">
          <div className="window flex-1 flex flex-col">
            <div className="window-bar flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="window-dots" aria-hidden="true">
                  <span />
                  <span />
                  <span />
                </span>
                <span className="window-title">gate-enforcement — {scenario.ref}</span>
              </div>
              <div className="flex items-center gap-2">
                <span
                  className={`mono text-[9px] uppercase px-2 py-0.5 rounded font-semibold tracking-wider ${
                    scenario.status === 'passed'
                      ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                      : scenario.status === 'rejected'
                      ? 'bg-rose-500/20 text-rose-400 border border-rose-500/30'
                      : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                  }`}
                >
                  {scenario.httpStatus}
                </span>
              </div>
            </div>

            {/* Terminal Body */}
            <div key={scenario.id} className="p-5 flex-1 font-mono text-[12px] leading-relaxed overflow-x-auto space-y-1.5 rise-in">
              {scenario.steps.map((st, i) => (
                <div
                  key={i}
                  className={`whitespace-pre-wrap ${
                    st.tone === 'cmd'
                      ? 'text-patina font-semibold'
                      : st.tone === 'success'
                      ? 'text-emerald-400'
                      : st.tone === 'danger'
                      ? 'text-rose-400 font-semibold'
                      : st.tone === 'dim'
                      ? 'text-text-3'
                      : 'text-text-2'
                  }`}
                >
                  {st.text}
                </div>
              ))}
              <div className="pt-2">
                <span className="caret" aria-hidden="true" />
              </div>
            </div>

            {/* Real-Time Durable Audit Ledger Footprint */}
            <div className="border-t border-line bg-raised/50 p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="mono text-[10px] uppercase tracking-wider text-text-3 flex items-center gap-1.5">
                  <IconShield width={13} height={13} className="text-patina" />
                  Durable JSONL Audit Ledger Entry:
                </span>
                <span
                  className={`mono text-[10px] font-bold px-1.5 py-0.5 rounded ${
                    scenario.auditEntry.decision === 'ALLOW'
                      ? 'bg-emerald-500/20 text-emerald-300'
                      : 'bg-rose-500/20 text-rose-300'
                  }`}
                >
                  DECISION: {scenario.auditEntry.decision}
                </span>
              </div>
              <pre className="mono text-[10.5px] bg-canvas p-2.5 rounded border border-line text-text-2 overflow-x-auto">
{JSON.stringify(scenario.auditEntry, null, 2)}
              </pre>
            </div>
          </div>
          <p className="mono mt-3 text-[10.5px] text-text-3 flex items-center justify-between">
            <span>interactive policy gate simulation · zero stale media</span>
            <span>internal/security/hook.go</span>
          </p>
        </div>
      </div>
    </section>
  );
}
