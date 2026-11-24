import Link from 'next/link';
import { Logomark, IconGitHub, IconArrowUpRight } from '@/components/icons';
import { CopyInstall } from '@/components/copy-install';
import { asset } from '@/lib/asset';

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col bg-canvas text-text-1">
      {/* Top Navigation */}
      <header className="sticky top-0 z-40 border-b border-hairline bg-canvas/90 backdrop-blur-sm">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
          <Link href="/" className="flex items-center gap-2.5">
            <Logomark size={22} />
            <span className="text-[17px] font-semibold tracking-[-0.03em] text-text-1">refweir.dev</span>
            <span className="rounded-sm border border-hairline bg-raised px-1.5 py-0.5 text-[9.5px] font-mono uppercase tracking-[0.18em] text-text-3">
              v0.6.0-beta
            </span>
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link
              href="/docs/"
              className="text-text-2 hover:text-text-1 transition-colors"
            >
              Documentation
            </Link>
            <a
              href="https://github.com/RavaniRoshan/refweir/releases"
              target="_blank"
              rel="noreferrer"
              className="text-text-2 hover:text-text-1 transition-colors hidden sm:inline"
            >
              Releases
            </a>
            <a
              href="https://github.com/RavaniRoshan/refweir"
              target="_blank"
              rel="noreferrer"
              aria-label="GitHub Repository"
              className="flex items-center gap-1.5 text-text-2 hover:text-text-1 transition-colors"
            >
              <IconGitHub className="size-4" />
              <span className="hidden sm:inline">GitHub</span>
            </a>
          </nav>
        </div>
      </header>

      <main className="flex-1">
        {/* Hero Section */}
        <section className="mx-auto max-w-5xl px-4 pt-16 pb-12 sm:px-6 sm:pt-20 sm:pb-16 text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-hairline bg-raised px-3 py-1 text-xs font-mono text-text-2 mb-6">
            <span className="size-2 rounded-full bg-patina inline-block" />
            <span>v0.6.0-beta · Apache-2.0 · Zero CGO</span>
          </div>

          <h1 className="text-3xl font-bold tracking-tight text-text-1 sm:text-5xl max-w-3xl mx-auto">
            The self-hosted Git forge for coding agents.
          </h1>

          <p className="mt-4 text-base sm:text-lg text-text-2 max-w-2xl mx-auto">
            Scoped capability leases, native MCP, and policy-gated merges on your own hardware in a single pure-Go binary.
          </p>

          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/docs/"
              className="inline-flex items-center justify-center gap-1.5 rounded-md bg-patina px-4 py-2.5 text-sm font-medium text-patina-ink hover:bg-patina-hover transition-colors"
            >
              <span>Read Documentation</span>
              <IconArrowUpRight className="size-4" />
            </Link>
            <a
              href="https://github.com/RavaniRoshan/refweir"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center justify-center gap-1.5 rounded-md border border-hairline bg-slab px-4 py-2.5 text-sm font-medium text-text-1 hover:bg-raised transition-colors"
            >
              <IconGitHub className="size-4" />
              <span>GitHub</span>
            </a>
          </div>

          <div className="mt-8 flex justify-center">
            <CopyInstall />
          </div>
        </section>

        {/* Demo Video Section */}
        <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6">
          <div className="overflow-hidden rounded-md border border-hairline bg-slab shadow-sm">
            <div className="flex items-center justify-between border-b border-hairline bg-raised px-4 py-2.5">
              <span className="font-mono text-xs text-text-2">
                demo.mp4 — 70-second terminal walkthrough
              </span>
              <span className="font-mono text-[11px] text-text-3">1200×680</span>
            </div>
            <div className="p-2 sm:p-4 bg-canvas">
              <video
                controls
                playsInline
                preload="metadata"
                title="Refweir 70-second terminal demo: bootstrap, agent leases, ref isolation, PR review, and auto-merge"
                className="w-full rounded border border-hairline bg-raised"
                src={asset('/demo.mp4')}
              >
                Your browser does not support the video tag.
              </video>
            </div>
          </div>
        </section>

        {/* 3 Feature Blocks */}
        <section className="mx-auto max-w-5xl px-4 py-12 sm:px-6">
          <div className="grid gap-6 md:grid-cols-3">
            <div className="rounded-md border border-hairline bg-slab p-6">
              <div className="font-mono text-xs text-patina font-medium mb-2">01 / CONTAINMENT</div>
              <h2 className="text-lg font-semibold text-text-1">Scoped Capability Leases</h2>
              <p className="mt-2 text-sm text-text-2 leading-relaxed">
                HMAC-SHA256 capability tokens (<code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">pkg/agentlease</code>) with 1-hour expiration. Agents are strictly confined to <code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">refs/heads/agent/*</code>. Direct pushes to <code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">main</code> are cryptographically rejected at the pre-receive boundary.
              </p>
            </div>

            <div className="rounded-md border border-hairline bg-slab p-6">
              <div className="font-mono text-xs text-patina font-medium mb-2">02 / INTEROPERABILITY</div>
              <h2 className="text-lg font-semibold text-text-1">Native Model Context Protocol</h2>
              <p className="mt-2 text-sm text-text-2 leading-relaxed">
                Embedded JSON-RPC 2.0 MCP endpoints at <code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">/mcp</code> and a read-only sandbox at <code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">/mcp/readonly</code>. Autonomous coding agents read trees, inspect diffs, open pull requests, and review code via standard MCP tools.
              </p>
            </div>

            <div className="rounded-md border border-hairline bg-slab p-6">
              <div className="font-mono text-xs text-patina font-medium mb-2">03 / HERMETIC ARCHITECTURE</div>
              <h2 className="text-lg font-semibold text-text-1">Single Hermetic Binary</h2>
              <p className="mt-2 text-sm text-text-2 leading-relaxed">
                Pure-Go zero CGO executable (<code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">CGO_ENABLED=0</code>). Embedded React 19 product UI, bare Git object store, and SQLite WAL database. Serves Smart-HTTP on port <code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">:8080</code> and pure-Go SSH on port <code className="font-mono text-xs bg-raised px-1 py-0.5 rounded border border-hairline">:2222</code>.
              </p>
            </div>
          </div>
        </section>

        {/* Honest Limits Section */}
        <section className="mx-auto max-w-5xl px-4 py-12 sm:px-6 border-t border-hairline">
          <div className="max-w-3xl">
            <div className="font-mono text-xs text-brass font-medium mb-2">TRANSPARENCY</div>
            <h2 className="text-2xl font-bold tracking-tight text-text-1">Honest Architectural Limits</h2>
            <p className="mt-2 text-sm text-text-2">
              Refweir documents its engineering trade-offs openly:
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-md border border-hairline bg-slab p-4">
                <h3 className="text-sm font-semibold text-text-1">1. SQLite WAL Concurrency (Single-Writer)</h3>
                <p className="mt-1 text-sm text-text-2 leading-relaxed">
                  Refweir uses pure-Go SQLite in Write-Ahead Logging (WAL) mode. Reads scale concurrently across goroutines, while disk write transactions (<code className="font-mono text-xs bg-raised px-1 rounded border border-hairline">git push</code>, PR creation, comments) are coordinated through SQLite WAL locking. Tuned pragmas deliver ~9,880 write transactions/sec and ~67,100 read lookups/sec, optimized for teams of 1–50 concurrent agents and engineers.
                </p>
              </div>

              <div className="rounded-md border border-hairline bg-slab p-4">
                <h3 className="text-sm font-semibold text-text-1">2. go-git Performance Profile</h3>
                <p className="mt-1 text-sm text-text-2 leading-relaxed">
                  Bare repository operations use <code className="font-mono text-xs bg-raised px-1 rounded border border-hairline">go-git/v6</code> directly in Go memory. While hermetic and CGO-free, <code className="font-mono text-xs bg-raised px-1 rounded border border-hairline">go-git</code> exhibits higher memory consumption on massive repositories (&gt; 2 GB) compared to native C git.
                </p>
              </div>

              <div className="rounded-md border border-hairline bg-slab p-4">
                <h3 className="text-sm font-semibold text-text-1">3. Event-Driven Auto-Merge</h3>
                <p className="mt-1 text-sm text-text-2 leading-relaxed">
                  Refweir provides an event-driven auto-merge engine (<code className="font-mono text-xs bg-raised px-1 rounded border border-hairline">AutoMergeEngine</code>) that monitors PR requirements and executes squash or 3-way merge operations once required status check suites pass.
                </p>
              </div>

              <div className="rounded-md border border-hairline bg-slab p-4">
                <h3 className="text-sm font-semibold text-text-1">4. Telemetry (Off by Default)</h3>
                <p className="mt-1 text-sm text-text-2 leading-relaxed">
                  Telemetry is off by default and consent-based. When explicitly enabled, Refweir sends only an anonymous instance ID, event name, version, and OS family. It never transmits repository names, code, commit content, or user identifiers.
                </p>
              </div>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-hairline bg-raised mt-16">
        <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-4 px-4 py-8 sm:flex-row sm:px-6">
          <div className="flex items-center gap-2">
            <Logomark size={18} />
            <span className="text-xs text-text-2">
              Refweir · Apache-2.0 License · Pure-Go
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-6 text-xs text-text-2">
            <Link href="/docs/" className="hover:text-text-1 transition-colors">
              Documentation
            </Link>
            <Link href="/docs/architecture/" className="hover:text-text-1 transition-colors">
              Architecture
            </Link>
            <Link href="/docs/legal/" className="hover:text-text-1 transition-colors">
              Privacy & Terms
            </Link>
            <a
              href="https://github.com/RavaniRoshan/refweir"
              target="_blank"
              rel="noreferrer"
              className="hover:text-text-1 transition-colors"
            >
              GitHub
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
