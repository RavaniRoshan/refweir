'use client';

import { useState } from 'react';
import { LiveClock } from './LiveClock';
import { Logomark } from './icons';
import { asset } from '@/lib/asset';

const COLUMNS: Array<[title: string, links: Array<[label: string, href: string]>]> = [
  [
    'product',
    [
      ['Features & Bento', '#product'],
      ['Agent Containment Loop', '#loop'],
      ['Containment Demo Video', '#demo'],
      ['Quickstart & Install', '#quickstart'],
      ['Releases', 'https://github.com/RavaniRoshan/refweir/releases'],
    ],
  ],
  [
    'documentation',
    [
      ['Getting Started Guide', asset('/docs/getting-started/')],
      ['Architecture Blueprint', asset('/docs/architecture/')],
      ['Model Context Protocol', asset('/docs/reference/mcp-tools/')],
      ['Verified Claims Ledger', 'https://github.com/RavaniRoshan/refweir/blob/main/CLAIMS.md'],
      ['Threat Model', 'https://github.com/RavaniRoshan/refweir/blob/main/THREAT_MODEL.md'],
    ],
  ],
  [
    'security & core',
    [
      ['Capability Leases (pkg)', 'https://github.com/RavaniRoshan/refweir/blob/main/pkg/agentlease/lease.go'],
      ['Native MCP Server (:8080)', 'https://github.com/RavaniRoshan/refweir/tree/main/internal/mcp'],
      ['Pure-Go SSH Daemon (:2222)', 'https://github.com/RavaniRoshan/refweir/tree/main/internal/sshserver'],
      ['Security Disclosure', 'https://github.com/RavaniRoshan/refweir/blob/main/SECURITY.md'],
      ['Apache-2.0 License', 'https://github.com/RavaniRoshan/refweir/blob/main/LICENSE'],
    ],
  ],
  [
    'community',
    [
      ['GitHub Repository', 'https://github.com/RavaniRoshan/refweir'],
      ['Issue Tracker', 'https://github.com/RavaniRoshan/refweir/issues'],
      ['Pull Requests', 'https://github.com/RavaniRoshan/refweir/pulls'],
      ['Contributing Guide', 'https://github.com/RavaniRoshan/refweir/blob/main/CONTRIBUTING.md'],
      ['Code of Conduct', 'https://github.com/RavaniRoshan/refweir/blob/main/CODE_OF_CONDUCT.md'],
    ],
  ],
];

export function StatusFooter() {
  const [copied, setCopied] = useState(false);
  const bootstrapCmd = 'refweird bootstrap --email admin@example.com --namespace acme --data-dir ./data';

  const copyBootstrap = async () => {
    try {
      await navigator.clipboard.writeText(bootstrapCmd);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
    }
  };

  return (
    <footer className="relative overflow-hidden border-t border-line/60 bg-canvas text-text-1">
      {/* Top Subtle Gradient Separator */}
      <div className="h-px w-full bg-gradient-to-r from-transparent via-patina/40 to-transparent" />

      {/* Main Footer Body */}
      <div className="container-site relative z-10 py-16 lg:py-20">
        <div className="grid gap-12 lg:grid-cols-12 lg:gap-8">
          {/* Brand & Mission Column */}
          <div className="lg:col-span-4">
            <a href={asset('/')} className="group inline-flex items-center gap-3" aria-label="refweir.dev — home">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg border border-line bg-slab transition-transform duration-200 group-hover:scale-105 group-hover:border-patina/60">
                <Logomark size={20} />
              </div>
              <span className="text-xl font-bold tracking-tight text-text-1">
                refweir<span className="text-patina">.dev</span>
              </span>
            </a>

            <p className="mt-4 max-w-sm text-sm leading-relaxed text-text-2">
              The neutral, self-hosted Git forge built for autonomous AI coding agents alongside human engineering teams. Single binary, zero CGO, pure-Go SSH.
            </p>

            {/* Live Operational Status Card */}
            <div className="mt-6 flex max-w-xs items-center justify-between rounded-lg border border-line bg-slab/80 p-3 shadow-xs">
              <div className="flex items-center gap-2.5">
                <span className="relative flex h-2.5 w-2.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-emerald-500" />
                </span>
                <span className="mono text-[11px] font-medium tracking-wide text-text-2">
                  All Systems Operational
                </span>
              </div>
              <span className="mono rounded border border-emerald-500/20 bg-emerald-500/10 px-1.5 py-0.5 text-[9px] font-semibold text-emerald-400 uppercase">
                v0.6.0-beta
              </span>
            </div>

            {/* Quick Copy Command Pill */}
            <div className="mt-4 flex max-w-xs items-center justify-between rounded-lg border border-line/80 bg-raised px-3 py-2">
              <code className="mono truncate text-[11px] text-text-2">
                {bootstrapCmd}
              </code>
              <button
                type="button"
                onClick={copyBootstrap}
                className="mono ml-2 shrink-0 rounded bg-slab px-2 py-1 text-[10px] font-semibold text-text-1 uppercase transition-colors hover:bg-patina hover:text-white"
                title="Copy bootstrap command"
              >
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
          </div>

          {/* Navigation Columns */}
          <div className="grid grid-cols-2 gap-8 sm:grid-cols-4 lg:col-span-8">
            {COLUMNS.map(([title, links]) => (
              <nav key={title} aria-label={title} className="flex flex-col">
                <p className="mono text-[11px] font-semibold tracking-wider text-text-3 uppercase">
                  {title}
                </p>
                <ul className="mt-4 flex flex-col space-y-2.5">
                  {links.map(([label, href]) => (
                    <li key={label}>
                      <a
                        href={href}
                        className="group inline-flex items-center gap-1 text-[13px] text-text-2 transition-colors duration-150 hover:text-patina"
                        rel={href.startsWith('http') ? 'noreferrer' : undefined}
                      >
                        <span className="transition-transform duration-150 group-hover:translate-x-0.5">
                          {label}
                        </span>
                        <span
                          aria-hidden="true"
                          className="text-[10px] opacity-0 transition-opacity duration-150 group-hover:opacity-100 text-patina"
                        >
                          ↗
                        </span>
                      </a>
                    </li>
                  ))}
                </ul>
              </nav>
            ))}
          </div>
        </div>

        {/* Massive AI-Style Background Watermark */}
        <div
          aria-hidden="true"
          className="pointer-events-none select-none overflow-hidden pt-12 pb-2 text-center"
        >
          <span className="inline-block text-[clamp(4.5rem,14vw,11.5rem)] font-black tracking-tighter text-text-1 opacity-[0.04] leading-none">
            REFWEIR
          </span>
        </div>
      </div>

      {/* Bottom Status Bar */}
      <div className="border-t border-line/70 bg-slab/90 py-3 text-text-3 backdrop-blur-xs">
        <div className="container-site flex flex-wrap items-center justify-between gap-4 text-xs">
          <div className="flex flex-wrap items-center gap-3">
            <span className="mono font-medium text-text-2">
              v0.6.0-beta
            </span>
            <span className="text-line-strong">•</span>
            <span>pure Go</span>
            <span className="text-line-strong">•</span>
            <span>zero CGO</span>
            <span className="text-line-strong">•</span>
            <span>apache-2.0</span>
            <span className="text-line-strong hidden sm:inline">•</span>
            <span className="hidden sm:inline mono text-[11px]">
              :8080 web / rest / mcp &nbsp;|&nbsp; :2222 ssh
            </span>
          </div>

          <div className="flex items-center gap-4">
            <LiveClock />
            <span className="text-text-3">© 2026 refweir.dev contributors</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
