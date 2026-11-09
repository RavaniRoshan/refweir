'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { IconCheck, IconPullRequest, IconShield } from './icons';

type Phase = 'idle' | 'merging' | 'merged';

const CHECKS = [
  { id: 'rulesets', label: 'Rulesets & branch guard', detail: 'refs/heads/agent/* isolation · main protected' },
  { id: 'secrets', label: 'Pre-receive secret scanner', detail: 'zero leaked API keys · 0 private tokens' },
  { id: 'review', label: 'Autonomous reviewer approval', detail: 'reviewer agent + human review sign-off' },
  { id: 'linear', label: 'Fast-forward linear history', detail: 'rebased clean diff · zero merge commits' },
] as const;

export function MergeGate() {
  const [checked, setChecked] = useState<Record<string, boolean>>({
    rulesets: true,
    secrets: true,
    review: true,
    linear: false,
  });
  const [phase, setPhase] = useState<Phase>('idle');

  const passed = CHECKS.filter((check) => checked[check.id]).length;
  const all = passed === CHECKS.length;

  useEffect(() => {
    if (phase !== 'merging') return;
    const id = window.setTimeout(() => setPhase('merged'), 950);
    return () => window.clearTimeout(id);
  }, [phase]);

  if (phase === 'merged') {
    return (
      <div className="frame frame-strong rise-in flex h-full min-h-[300px] flex-col justify-between bg-canvas p-6 text-text-1 sm:p-7 border-patina shadow-lg" data-testid="merge-gate-merged">
        <div>
          <div className="flex items-center justify-between gap-4">
            <span className="mono text-[11px] uppercase tracking-[0.18em] text-text-3">pr-#214 · agent/capture</span>
            <span className="mono rounded-full border border-patina/40 bg-patina/10 px-2.5 py-1 text-[10px] uppercase font-bold tracking-[0.18em] text-patina">
              merged · 200 ok
            </span>
          </div>
          <div className="mt-6 flex items-center gap-4">
            <span className="relative flex h-11 w-11 items-center justify-center rounded-full border border-emerald-500/60 bg-emerald-500/20 text-emerald-400 shadow-sm">
              <IconCheck width={22} height={22} strokeWidth={2.5} />
            </span>
            <div>
              <p className="text-xl font-medium leading-tight tracking-tight text-text-1">Merge Gate Passed</p>
              <p className="mono text-[11px] text-text-3 mt-0.5">Commit 9b2f1c4 landed on refs/heads/main</p>
            </div>
          </div>
        </div>
        <div className="mt-6 border-t border-line pt-4 bg-slab/50 -mx-6 -mb-6 p-6 rounded-b-lg">
          <div className="flex items-center justify-between text-xs mb-2">
            <span className="mono text-[10px] uppercase tracking-wider text-text-3 flex items-center gap-1.5">
              <IconShield width={13} height={13} className="text-patina" />
              Durable Audit Ledger Entry:
            </span>
            <span className="mono text-[10px] text-emerald-400 font-bold">ALLOW (a-1042)</span>
          </div>
          <pre className="mono text-[11px] text-text-2 bg-canvas p-2.5 rounded border border-line overflow-x-auto leading-relaxed">
{`{
  "event": "pr.merged",
  "pr": 214,
  "head": "agent/capture",
  "base": "main",
  "lease_revoked": true,
  "status": "200_OK"
}`}
          </pre>
          <button
            type="button"
            className="mono mt-4 text-[10.5px] uppercase tracking-[0.16em] text-patina hover:underline cursor-pointer"
            onClick={() => {
              setChecked({ rulesets: true, secrets: true, review: true, linear: false });
              setPhase('idle');
            }}
          >
            ← Reset policy gate simulation
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="frame frame-strong flex h-full flex-col p-5 sm:p-6" data-testid="merge-gate">
      <div className="flex items-center justify-between gap-4 border-b border-line pb-4">
        <span className="mono flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-text-2 font-medium">
          <IconPullRequest width={14} height={14} className="text-patina" /> pr-#214 · agent/capture
        </span>
        <span
          className={`mono text-[10px] uppercase px-2 py-0.5 rounded font-semibold tracking-wider ${
            all
              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
              : 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
          }`}
        >
          {all ? 'READY (200 OK)' : `BLOCKED (${passed}/${CHECKS.length})`}
        </span>
      </div>

      <fieldset className="mt-4 flex flex-col gap-2.5">
        <legend className="sr-only">Required merge checks</legend>
        {CHECKS.map((check) => {
          const isChecked = Boolean(checked[check.id]);
          return (
            <div
              key={check.id}
              className={`gate-item transition-all cursor-pointer rounded-md border p-3 flex items-center gap-3 ${
                isChecked
                  ? 'border-line-strong bg-slab/80'
                  : 'border-rose-500/20 bg-rose-500/5'
              }`}
              data-checked={isChecked}
              onClick={() => setChecked((state) => ({ ...state, [check.id]: !state[check.id] }))}
            >
              <span className="contents" onClick={(event) => event.stopPropagation()}>
                <Checkbox
                  unstyled
                  aria-label={check.label}
                  className="gate-box cursor-pointer"
                  isSelected={isChecked}
                  onChange={(selected) =>
                    setChecked((state) => ({ ...state, [check.id]: selected }))
                  }
                />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block text-[14px] font-medium leading-tight text-text-1">{check.label}</span>
                <span className="mono mt-1 block text-[10.5px] leading-tight text-text-3">{check.detail}</span>
              </span>
              <span className={`mono text-[10px] font-semibold uppercase tracking-[0.18em] ${isChecked ? 'text-emerald-400' : 'text-rose-400'}`}>
                {isChecked ? 'pass' : 'fail'}
              </span>
            </div>
          );
        })}
      </fieldset>

      <div className="mt-5 border-t border-line pt-4">
        <Button
          unstyled
          type="button"
          className="btn btn-signal w-full justify-center disabled:cursor-not-allowed disabled:border-line disabled:bg-raised disabled:text-text-3 font-semibold"
          isDisabled={!all || phase === 'merging'}
          onPress={() => setPhase('merging')}
        >
          {phase === 'merging' ? (
            <>
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2" aria-hidden="true" />
              Evaluating policy & merging…
            </>
          ) : (
            'Merge pull request'
          )}
        </Button>
        <p className="mono mt-3 text-center text-[10.5px] leading-relaxed text-text-3" role="status">
          {all ? 'All checks satisfied — merge permitted by forge policy' : 'Protected ref violation — direct merge denied (GH007)'}
        </p>
      </div>
    </div>
  );
}
