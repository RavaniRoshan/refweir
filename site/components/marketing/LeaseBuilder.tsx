'use client';

import { useState, useMemo } from 'react';

const AGENTS = ['claude-code', 'cursor-agent', 'codestral'] as const;
const TTLS = [
  { label: '5m', seconds: 300 },
  { label: '15m', seconds: 900 },
  { label: '30m', seconds: 1800 },
  { label: '60m', seconds: 3600 },
] as const;

export function LeaseBuilder() {
  const [selectedAgent, setSelectedAgent] = useState<typeof AGENTS[number]>('claude-code');
  const [selectedTTL, setSelectedTTL] = useState<number>(900);
  const [targetScope, setTargetScope] = useState<'isolated' | 'main'>('isolated');
  const [copied, setCopied] = useState(false);

  // Deterministic mock HMAC signature based on current parameters
  const { token, signature, isAllowed } = useMemo(() => {
    const isAllowed = targetScope === 'isolated';
    const agentHex = selectedAgent.slice(0, 4);
    const ttlHex = selectedTTL.toString(16);
    const scopeHex = isAllowed ? '7c01' : '0000';
    const token = `lease_hmac_${agentHex}${ttlHex}${scopeHex}e8b44d12`;
    const signature = `hmac-sha256:${agentHex}${ttlHex}f93a8d11c05e927a`;
    return { token, signature, isAllowed };
  }, [selectedAgent, selectedTTL, targetScope]);

  const copyToken = async () => {
    try {
      await navigator.clipboard.writeText(token);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
    }
  };

  return (
    <div className="mt-4 flex flex-col space-y-4">
      {/* Interactive Controls Bar */}
      <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-3">
        {/* Agent Selector */}
        <div>
          <label className="mono block text-[10px] font-semibold text-text-3 uppercase tracking-wider">
            Agent Identity
          </label>
          <div className="mt-1 flex rounded-md border border-line bg-raised p-0.5">
            {AGENTS.map((agent) => (
              <button
                key={agent}
                type="button"
                onClick={() => setSelectedAgent(agent)}
                className={`flex-1 rounded px-1.5 py-1 mono text-[11px] font-medium transition-colors ${
                  selectedAgent === agent
                    ? 'bg-slab text-text-1 shadow-xs'
                    : 'text-text-3 hover:text-text-2'
                }`}
              >
                {agent.split('-')[0]}
              </button>
            ))}
          </div>
        </div>

        {/* TTL Selector */}
        <div>
          <label className="mono block text-[10px] font-semibold text-text-3 uppercase tracking-wider">
            Lease Duration (TTL)
          </label>
          <div className="mt-1 flex rounded-md border border-line bg-raised p-0.5">
            {TTLS.map((ttl) => (
              <button
                key={ttl.label}
                type="button"
                onClick={() => setSelectedTTL(ttl.seconds)}
                className={`flex-1 rounded px-1.5 py-1 mono text-[11px] font-medium transition-colors ${
                  selectedTTL === ttl.seconds
                    ? 'bg-slab text-patina font-semibold shadow-xs'
                    : 'text-text-3 hover:text-text-2'
                }`}
              >
                {ttl.label}
              </button>
            ))}
          </div>
        </div>

        {/* Ref Scope Target */}
        <div>
          <label className="mono block text-[10px] font-semibold text-text-3 uppercase tracking-wider">
            Target Ref Scope
          </label>
          <div className="mt-1 flex rounded-md border border-line bg-raised p-0.5">
            <button
              type="button"
              onClick={() => setTargetScope('isolated')}
              className={`flex-1 rounded px-1.5 py-1 mono text-[11px] font-medium transition-colors ${
                targetScope === 'isolated'
                  ? 'bg-slab text-emerald-400 font-semibold shadow-xs'
                  : 'text-text-3 hover:text-text-2'
              }`}
            >
              agent/*
            </button>
            <button
              type="button"
              onClick={() => setTargetScope('main')}
              className={`flex-1 rounded px-1.5 py-1 mono text-[11px] font-medium transition-colors ${
                targetScope === 'main'
                  ? 'bg-slab text-rose-400 font-semibold shadow-xs'
                  : 'text-text-3 hover:text-text-2'
              }`}
            >
              main ⚠
            </button>
          </div>
        </div>
      </div>

      {/* Simulated Live Token Box */}
      <div className="rounded-lg border border-line bg-raised p-3.5 transition-all duration-200">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span
              className={`h-2 w-2 rounded-full ${
                isAllowed ? 'bg-emerald-500 animate-pulse' : 'bg-rose-500'
              }`}
            />
            <code className="mono text-xs font-semibold text-text-1">{token}</code>
          </div>
          <button
            type="button"
            onClick={copyToken}
            className="mono rounded bg-slab px-2.5 py-1 text-[11px] font-medium text-text-2 transition-colors hover:bg-patina hover:text-white"
          >
            {copied ? 'Copied' : 'Copy'}
          </button>
        </div>

        {/* Scope and Policy Status Line */}
        <div className="mt-2.5 flex flex-wrap items-center justify-between gap-2 border-t border-line/60 pt-2 text-[11px]">
          <div className="mono text-text-3">
            Scope:{' '}
            <code className="text-text-1">
              {isAllowed ? 'refs/heads/agent/*' : 'refs/heads/main'}
            </code>
          </div>
          <div className="mono font-semibold">
            {isAllowed ? (
              <span className="text-emerald-400">
                TTL: {selectedTTL / 60}m · ISOLATION ACTIVE ✓
              </span>
            ) : (
              <span className="text-rose-400">
                BLOCKED: POLICY GH006 REJECTED ✕
              </span>
            )}
          </div>
        </div>

        {/* Micro Policy Feedback Callout */}
        <p className="mt-2 text-[11px] leading-relaxed text-text-3">
          {isAllowed
            ? '✓ Verified: Autonomous push granted strictly under capability lease to refs/heads/agent/*. Standing write to production main is denied.'
            : '✕ Violation GH006: Autonomous capability leases cannot target protected default branches directly. A pull request is required.'}
        </p>
      </div>
    </div>
  );
}
