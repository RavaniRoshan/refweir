import React, { useState } from 'react'
import { ArrowUpRight, CheckCircle2, FlaskConical, Loader2, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import {
  mergeDemoPullRequest,
  runDemoAgent,
  type DemoAgentRunResult,
  type DemoMergeResult,
} from '@/lib/api'

export interface DemoBannerProps {
  /** owner/name of the seeded demo repository, for context in the copy. */
  repo?: string
}

/**
 * Persistent banner for the hosted demo. It only renders when the server reports
 * REFWEIR_DEMO_MODE=1 via /api/v1/system/info; normal self-hosted instances never
 * show it. The agent run is a real lease-backed push, not canned data.
 */
export const DemoBanner: React.FC<DemoBannerProps> = ({ repo }) => {
  const [running, setRunning] = useState(false)
  const [merging, setMerging] = useState(false)
  const [result, setResult] = useState<DemoAgentRunResult | null>(null)
  const [merged, setMerged] = useState<DemoMergeResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleRun = async () => {
    setRunning(true)
    setError(null)
    setResult(null)
    setMerged(null)
    try {
      setResult(await runDemoAgent())
    } catch (err: any) {
      setError(err?.message || 'The demo agent run failed.')
    } finally {
      setRunning(false)
    }
  }

  const handleMerge = async () => {
    if (!result) return
    setMerging(true)
    setError(null)
    try {
      setMerged(await mergeDemoPullRequest(result.prNumber))
    } catch (err: any) {
      setError(err?.message || 'The demo merge failed.')
    } finally {
      setMerging(false)
    }
  }

  return (
    <div className="w-full border-b border-hairline bg-slab">
      <div className="mx-auto flex max-w-[1440px] flex-col gap-2 px-5 py-2.5 lg:flex-row lg:items-center lg:justify-between lg:px-8">
        <div className="flex items-start gap-2 text-xs text-text-2">
          <FlaskConical className="mt-0.5 size-3.5 shrink-0 text-signal" aria-hidden="true" />
          <p>
            <span className="font-semibold text-text-1">Live demo.</span>{' '}
            Seeded repository{repo ? ` (${repo})` : ''}, reset on a schedule. An agent run
            pushes a real <code className="rounded bg-raised px-1 py-0.5 text-[11px]">agent/*</code>{' '}
            branch under a scoped lease and opens a real pull request.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {result ? (
            merged ? (
              <div className="flex items-center gap-2 text-xs text-text-2" role="status">
                <CheckCircle2 className="size-3.5 text-signal" aria-hidden="true" />
                <span>
                  Merged PR #{merged.prNumber} into{' '}
                  <code className="rounded bg-raised px-1 py-0.5 text-[11px]">
                    {merged.baseRef}
                  </code>{' '}
                  as{' '}
                  <code className="rounded bg-raised px-1 py-0.5 text-[11px]">
                    {merged.mergeCommitSha.slice(0, 7)}
                  </code>
                </span>
                <a
                  href={result.prUrl}
                  className="inline-flex items-center gap-1 font-medium text-signal hover:underline"
                >
                  View the PR
                  <ArrowUpRight className="size-3" aria-hidden="true" />
                </a>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-xs text-text-2" role="status">
                <ShieldCheck className="size-3.5 text-signal" aria-hidden="true" />
                <span>
                  PR #{result.prNumber} opened on{' '}
                  <code className="rounded bg-raised px-1 py-0.5 text-[11px]">{result.branch}</code>
                </span>
                <a
                  href={result.prUrl}
                  className="inline-flex items-center gap-1 font-medium text-signal hover:underline"
                >
                  Review the PR
                  <ArrowUpRight className="size-3" aria-hidden="true" />
                </a>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={handleMerge}
                  disabled={merging}
                  aria-label="Merge the demo pull request"
                >
                  {merging ? (
                    <Loader2 className="size-3.5 animate-spin" aria-hidden="true" />
                  ) : (
                    <CheckCircle2 className="size-3.5" aria-hidden="true" />
                  )}
                  {merging ? 'Merging…' : 'Merge the PR'}
                </Button>
              </div>
            )
          ) : null}
          {error ? (
            <span className="text-xs text-destructive" role="alert">
              {error}
            </span>
          ) : null}

          <Button
            size="sm"
            variant="primary"
            onClick={handleRun}
            disabled={running}
            aria-label="Try an agent run on the live demo"
          >
            {running ? (
              <Loader2 className="size-3.5 animate-spin" aria-hidden="true" />
            ) : (
              <FlaskConical className="size-3.5" aria-hidden="true" />
            )}
            {running ? 'Running agent…' : 'Try an agent run'}
          </Button>
          <a
            href="https://github.com/RavaniRoshan/refweir#quickstart"
            className="inline-flex items-center gap-1 text-xs font-medium text-text-2 hover:text-text-1"
          >
            Self-host Refweir
            <ArrowUpRight className="size-3" aria-hidden="true" />
          </a>
        </div>
      </div>
    </div>
  )
}
