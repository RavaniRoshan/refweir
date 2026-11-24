import React, { useEffect, useState, useRef } from 'react'
import { X, Keyboard, ShieldAlert } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Kbd, KbdGroup } from '@/components/primitives/kbd'

export interface KeyboardShortcutsModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const SHORTCUT_GROUPS = [
  {
    category: 'Global & Navigation',
    shortcuts: [
      { keys: ['⌘', 'K'], label: 'Open Command Palette' },
      { keys: ['/'], label: 'Focus global search' },
      { keys: ['?'], label: 'Open Keyboard Shortcuts cheat sheet' },
      { keys: ['esc'], label: 'Close dialog, drawer, or modal' },
    ],
  },
  {
    category: 'Forge & Repository Views',
    shortcuts: [
      { keys: ['g', 'h'], label: 'Go to Enterprise Feed / Home' },
      { keys: ['g', 'c'], label: 'Go to Code Explorer' },
      { keys: ['g', 'i'], label: 'Go to Issues' },
      { keys: ['g', 'p'], label: 'Go to Pull Requests' },
      { keys: ['g', 'a'], label: 'Go to Agent Console' },
      { keys: ['g', 's'], label: 'Go to Repository Settings' },
    ],
  },
  {
    category: 'Code Review & Diff Viewer',
    shortcuts: [
      { keys: ['j'], label: 'Move cursor down to next line' },
      { keys: ['k'], label: 'Move cursor up to previous line' },
      { keys: ['w'], label: 'Toggle ignore whitespace changes (-w)' },
      { keys: ['s'], label: 'Toggle Split (side-by-side) vs Unified diff' },
      { keys: ['c'], label: 'Add inline review comment on active line' },
    ],
  },
  {
    category: 'Agent Capability & Leases',
    shortcuts: [
      { keys: ['c', 'l'], label: 'Issue ephemeral capability lease' },
      { keys: ['r', 'l'], label: 'Inspect active agent leases' },
    ],
  },
]

export const KeyboardShortcutsModal: React.FC<KeyboardShortcutsModalProps> = ({
  open,
  onOpenChange,
}) => {
  const [characterShortcutsEnabled, setCharacterShortcutsEnabled] = useState(() => {
    return localStorage.getItem('refweir.character_shortcuts_enabled') !== 'false'
  })
  const dialogRef = useRef<HTMLDivElement>(null)

  const handleToggleShortcuts = (enabled: boolean) => {
    setCharacterShortcutsEnabled(enabled)
    localStorage.setItem('refweir.character_shortcuts_enabled', enabled ? 'true' : 'false')
  }

  // Handle ESC key to dismiss
  useEffect(() => {
    if (!open) return
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        onOpenChange(false)
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [open, onOpenChange])

  // Trap focus inside modal
  useEffect(() => {
    if (!open) return
    const el = dialogRef.current
    if (!el) return

    const focusableElements = el.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )
    const firstElement = focusableElements[0]
    firstElement?.focus()
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-xs"
      role="dialog"
      aria-modal="true"
      aria-labelledby="shortcuts-modal-title"
      ref={dialogRef}
    >
      <div className="w-full max-w-2xl bg-card border border-border rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
        {/* Header */}
        <div className="px-5 py-4 border-b border-border flex items-center justify-between bg-secondary/30">
          <div className="flex items-center gap-2.5">
            <Keyboard className="w-5 h-5 text-primary" />
            <div>
              <h2 id="shortcuts-modal-title" className="text-sm font-semibold text-foreground">
                Keyboard Shortcuts
              </h2>
              <p className="text-[11px] text-muted-foreground">
                Premier navigation and code review keybindings for Refweir.
              </p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="xs"
            onClick={() => onOpenChange(false)}
            aria-label="Close shortcuts dialog"
            className="text-muted-foreground hover:text-foreground p-1 h-7 w-7"
          >
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Content body */}
        <div className="p-5 overflow-y-auto space-y-5 text-xs divide-y divide-border/60">
          {SHORTCUT_GROUPS.map((group, idx) => (
            <div key={group.category} className={idx > 0 ? 'pt-4 space-y-2.5' : 'space-y-2.5'}>
              <h3 className="font-semibold text-foreground text-xs uppercase tracking-wider text-[10px] text-muted-foreground">
                {group.category}
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {group.shortcuts.map((sc, sIdx) => (
                  <div
                    key={sIdx}
                    className="flex items-center justify-between p-2 rounded-md bg-secondary/20 border border-border/70"
                  >
                    <span className="text-foreground/90 font-sans">{sc.label}</span>
                    <KbdGroup>
                      {sc.keys.map((k, kIdx) => (
                        <Kbd key={kIdx} className="font-mono text-[11px] uppercase">
                          {k}
                        </Kbd>
                      ))}
                    </KbdGroup>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* WCAG 2.1.4 Compliance Footer */}
        <div className="px-5 py-3 border-t border-border bg-secondary/20 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs">
          <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
            <ShieldAlert className="w-3.5 h-3.5 text-primary shrink-0" />
            <span>WCAG 2.1.4 (Character Key Shortcuts) settings</span>
          </div>
          <label className="flex items-center gap-2 cursor-pointer select-none text-[11px] text-foreground font-medium">
            <input
              type="checkbox"
              checked={characterShortcutsEnabled}
              onChange={(e) => handleToggleShortcuts(e.target.checked)}
              className="rounded border-border text-primary focus:ring-primary"
            />
            <span>Enable single-character keybindings (?, j, k, w, s)</span>
          </label>
        </div>
      </div>
    </div>
  )
}
