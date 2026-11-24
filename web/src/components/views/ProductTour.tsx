import React, { useState } from 'react'
import {
  KeyRound,
  Bot,
  Terminal,
  CheckCircle2,
  Circle,
  ArrowRight,
  ArrowLeft,
  X,
  Copy,
  Check,
  BookOpen,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'

export interface ProductTourProps {
  isOpen: boolean
  onClose: () => void
  onOpenManual?: () => void
}

interface ChecklistStep {
  id: string
  title: string
  shortLabel: string
  subtitle: string
  icon: React.ReactNode
  badge: string
  content: React.ReactNode
}

export const ProductTour: React.FC<ProductTourProps> = ({
  isOpen,
  onClose,
  onOpenManual,
}) => {
  const [activeStepIndex, setActiveStepIndex] = useState(0)
  const [completedSteps, setCompletedSteps] = useState<Record<number, boolean>>({})
  const [copiedKey, setCopiedKey] = useState<string | null>(null)

  const handleCopy = (text: string, key: string) => {
    navigator.clipboard.writeText(text)
    setCopiedKey(key)
    setTimeout(() => setCopiedKey(null), 2000)
  }

  const handleSkip = () => {
    localStorage.setItem('refweir_onboarding_tour_completed', 'true')
    onClose()
  }

  const handleFinish = () => {
    localStorage.setItem('refweir_onboarding_tour_completed', 'true')
    onClose()
  }

  const toggleStepDone = (index: number) => {
    setCompletedSteps((prev) => ({
      ...prev,
      [index]: !prev[index],
    }))
  }

  const steps: ChecklistStep[] = [
    {
      id: 'create-token',
      title: 'Create Personal Access Token',
      shortLabel: '1. Create Token',
      subtitle: 'Generate an API token with repo and agent capability permissions',
      icon: <KeyRound className="w-5 h-5 text-primary" />,
      badge: 'Step 1 of 3',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            Autonomous coding agents and CLI runners authenticate to Refweir using scoped personal access tokens or cryptographic capability leases.
          </p>

          <div className="p-3 rounded-md bg-secondary/40 border border-border space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-semibold text-foreground">CLI Command</span>
              <Badge variant="outline" className="font-mono text-[10px]">refweir-cli</Badge>
            </div>
            <div className="flex items-center justify-between bg-input p-2 rounded border border-border font-mono text-xs text-foreground">
              <span className="truncate">refweir auth token create --name "agent-token" --scopes "repo,agent"</span>
              <button
                onClick={() => handleCopy('refweir auth token create --name "agent-token" --scopes "repo,agent"', 'token-cli')}
                aria-label="Copy CLI token command"
                className="p-1 text-muted-foreground hover:text-foreground cursor-pointer shrink-0 ml-2"
              >
                {copiedKey === 'token-cli' ? <Check className="w-3.5 h-3.5 text-success" /> : <Copy className="w-3.5 h-3.5" />}
              </button>
            </div>
          </div>

          <div className="p-3 rounded-md bg-secondary/20 border border-border space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-semibold text-foreground">cURL API Alternative</span>
              <span className="text-[10px] text-muted-foreground font-mono">POST /api/v1/auth/tokens</span>
            </div>
            <div className="flex items-center justify-between bg-input p-2 rounded border border-border font-mono text-[11px] text-foreground">
              <span className="truncate">curl -X POST http://localhost:8080/api/v1/auth/tokens -H "Authorization: Bearer $REFWEIR_TOKEN" -d '&#123;"name":"agent"&#125;'</span>
              <button
                onClick={() => handleCopy('curl -X POST http://localhost:8080/api/v1/auth/tokens -H "Authorization: Bearer $REFWEIR_TOKEN" -d \'{"name":"agent"}\'', 'token-curl')}
                aria-label="Copy cURL token command"
                className="p-1 text-muted-foreground hover:text-foreground cursor-pointer shrink-0 ml-2"
              >
                {copiedKey === 'token-curl' ? <Check className="w-3.5 h-3.5 text-success" /> : <Copy className="w-3.5 h-3.5" />}
              </button>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'copy-mcp-config',
      title: 'Copy MCP Config JSON',
      shortLabel: '2. Copy MCP Config',
      subtitle: 'Add Refweir native MCP server to your Claude Code, Cursor, or Cline configuration',
      icon: <Bot className="w-5 h-5 text-primary" />,
      badge: 'Step 2 of 3',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            Refweir mounts a native JSON-RPC 2.0 Model Context Protocol server exposing repository primitives, AST symbols, and capability leases directly to your AI agents.
          </p>

          <div className="p-3 rounded-md bg-secondary/40 border border-border space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-semibold text-foreground">mcpServers Config (JSON)</span>
              <Badge variant="outline" className="font-mono text-[10px]">~/.cursor/mcp.json</Badge>
            </div>
            <pre className="p-2.5 rounded bg-input border border-border font-mono text-[11px] text-foreground overflow-x-auto">
{`{
  "mcpServers": {
    "refweir": {
      "url": "http://127.0.0.1:8080/mcp",
      "headers": {
        "Authorization": "Bearer <YOUR_TOKEN>"
      }
    }
  }
}`}
            </pre>
            <div className="flex justify-end">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => handleCopy(`{\n  "mcpServers": {\n    "refweir": {\n      "url": "http://127.0.0.1:8080/mcp",\n      "headers": {\n        "Authorization": "Bearer <YOUR_TOKEN>"\n      }\n    }\n  }\n}`, 'mcp-json')}
                className="text-xs gap-1.5"
              >
                {copiedKey === 'mcp-json' ? <Check className="w-3.5 h-3.5 text-success" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copiedKey === 'mcp-json' ? 'Copied MCP Config' : 'Copy MCP Config JSON'}</span>
              </Button>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'run-agent-loop',
      title: 'Run Agent Loop on Seeded Demo Repo',
      shortLabel: '3. Run Agent Loop',
      subtitle: 'Seed the data room with demo repositories and start the autonomous agent loop',
      icon: <Terminal className="w-5 h-5 text-primary" />,
      badge: 'Step 3 of 3',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            Seed the demo repositories with passing checks, leases, and an audit trail, then trigger the autonomous loop to execute branch leases and pull requests.
          </p>

          <div className="p-3 rounded-md bg-secondary/40 border border-border space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-semibold text-foreground">Execute Loop on Seeded Demo</span>
              <Badge variant="outline" className="font-mono text-[10px]">bash</Badge>
            </div>
            <div className="flex items-center justify-between bg-input p-2 rounded border border-border font-mono text-xs text-foreground">
              <span className="truncate">./scripts/seed.sh && refweir agent loop --repo demo/agent-lab</span>
              <button
                onClick={() => handleCopy('./scripts/seed.sh && refweir agent loop --repo demo/agent-lab', 'seed-cmd')}
                aria-label="Copy seed and agent loop command"
                className="p-1 text-muted-foreground hover:text-foreground cursor-pointer shrink-0 ml-2"
              >
                {copiedKey === 'seed-cmd' ? <Check className="w-3.5 h-3.5 text-success" /> : <Copy className="w-3.5 h-3.5" />}
              </button>
            </div>
          </div>

          <div className="p-3 rounded-md bg-secondary/20 border border-border text-xs space-y-1.5">
            <div className="font-semibold text-foreground flex items-center gap-1.5">
              <CheckCircle2 className="w-3.5 h-3.5 text-success" />
              <span>What Happens Next</span>
            </div>
            <p className="text-[11px] text-muted-foreground leading-relaxed">
              The agent requests an ephemeral capability lease, pushes to <code className="px-1 py-0.5 rounded bg-secondary font-mono text-[10px]">refs/heads/agent/*</code>, opens a pull request, and triggers automated policy and review checks.
            </p>
          </div>
        </div>
      ),
    },
  ]

  if (!isOpen) return null

  const currentStep = steps[activeStepIndex]

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Refweir Quickstart Checklist"
      className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6"
    >
      <div
        className="fixed inset-0 bg-background/80 backdrop-blur-xs"
        onClick={handleSkip}
      />

      <div className="relative w-full max-w-2xl max-h-[90vh] rounded-lg border border-border bg-card text-card-foreground shadow-2xl z-10 flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Modal Header */}
        <div className="px-6 py-4 border-b border-border bg-secondary/30 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-md bg-primary/10 border border-primary/20 flex items-center justify-center text-primary">
              {currentStep.icon}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-semibold text-foreground tracking-tight">
                  Refweir Quickstart Checklist
                </h2>
                <Badge variant="outline" className="text-[10px] font-mono">
                  {currentStep.badge}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                Complete these 3 steps to get your self-hosted forge agent-ready
              </p>
            </div>
          </div>

          <button
            onClick={handleSkip}
            aria-label="Close checklist"
            className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary/60 transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* 3-Step Interactive Tab Bar */}
        <div className="grid grid-cols-3 border-b border-border bg-secondary/10 shrink-0 text-xs">
          {steps.map((s, idx) => {
            const isDone = completedSteps[idx]
            const isActive = activeStepIndex === idx
            return (
              <button
                key={s.id}
                onClick={() => setActiveStepIndex(idx)}
                className={cn(
                  'flex items-center justify-center gap-1.5 py-2.5 px-2 border-b-2 font-medium transition-colors text-center cursor-pointer',
                  isActive
                    ? 'border-primary text-foreground font-semibold bg-secondary/20'
                    : 'border-transparent text-muted-foreground hover:text-foreground hover:bg-secondary/10'
                )}
              >
                <span
                  onClick={(e) => {
                    e.stopPropagation()
                    toggleStepDone(idx)
                  }}
                  className="cursor-pointer"
                  title="Click to toggle step complete"
                >
                  {isDone ? (
                    <CheckCircle2 className="w-3.5 h-3.5 text-success" />
                  ) : (
                    <Circle className="w-3.5 h-3.5 text-muted-foreground" />
                  )}
                </span>
                <span className="truncate">{s.shortLabel}</span>
              </button>
            )
          })}
        </div>

        {/* Step Content */}
        <div className="p-6 overflow-y-auto space-y-4 flex-1 text-xs">
          <div>
            <h3 className="text-sm font-semibold text-foreground">{currentStep.title}</h3>
            <p className="text-muted-foreground text-xs mt-0.5">{currentStep.subtitle}</p>
          </div>

          <div key={currentStep.id} className="transition-opacity duration-150">
            {currentStep.content}
          </div>
        </div>

        {/* Modal Footer */}
        <div className="px-6 py-3.5 border-t border-border bg-secondary/20 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleSkip}
              className="text-xs"
            >
              Skip
            </Button>
            {onOpenManual && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  handleFinish()
                  onOpenManual()
                }}
                className="text-xs text-muted-foreground gap-1.5"
              >
                <BookOpen className="w-3.5 h-3.5" />
                <span>Open Docs</span>
              </Button>
            )}
          </div>

          <div className="flex items-center gap-2">
            {activeStepIndex > 0 && (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setActiveStepIndex((prev) => prev - 1)}
                className="text-xs gap-1"
              >
                <ArrowLeft className="w-3.5 h-3.5" />
                <span>Back</span>
              </Button>
            )}

            {activeStepIndex < steps.length - 1 ? (
              <Button
                size="sm"
                onClick={() => {
                  setCompletedSteps((prev) => ({ ...prev, [activeStepIndex]: true }))
                  setActiveStepIndex((prev) => prev + 1)
                }}
                className="text-xs gap-1"
              >
                <span>Next</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={handleFinish}
                className="text-xs bg-primary text-primary-foreground gap-1 font-semibold"
              >
                <span>Get Started</span>
                <Check className="w-3.5 h-3.5" />
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
