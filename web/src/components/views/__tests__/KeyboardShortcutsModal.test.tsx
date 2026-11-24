import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { KeyboardShortcutsModal } from '../KeyboardShortcutsModal'

describe('KeyboardShortcutsModal Component', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('renders modal with shortcut categories and keybadges when open', () => {
    render(<KeyboardShortcutsModal open={true} onOpenChange={() => {}} />)

    expect(screen.getByText('Keyboard Shortcuts')).toBeInTheDocument()
    expect(screen.getByText('Global & Navigation')).toBeInTheDocument()
    expect(screen.getByText('Forge & Repository Views')).toBeInTheDocument()
    expect(screen.getByText('Code Review & Diff Viewer')).toBeInTheDocument()
    expect(screen.getByText('Agent Capability & Leases')).toBeInTheDocument()

    // Specific key labels
    expect(screen.getByText('Open Command Palette')).toBeInTheDocument()
    expect(screen.getByText('Issue ephemeral capability lease')).toBeInTheDocument()
    expect(screen.getByText('WCAG 2.1.4 (Character Key Shortcuts) settings')).toBeInTheDocument()
  })

  it('does not render dialog when open is false', () => {
    const { container } = render(<KeyboardShortcutsModal open={false} onOpenChange={() => {}} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('calls onOpenChange(false) when clicking close button', () => {
    const handleClose = vi.fn()
    render(<KeyboardShortcutsModal open={true} onOpenChange={handleClose} />)

    const closeBtn = screen.getByRole('button', { name: /close shortcuts dialog/i })
    fireEvent.click(closeBtn)

    expect(handleClose).toHaveBeenCalledWith(false)
  })

  it('calls onOpenChange(false) on Escape key press', () => {
    const handleClose = vi.fn()
    render(<KeyboardShortcutsModal open={true} onOpenChange={handleClose} />)

    fireEvent.keyDown(window, { key: 'Escape' })
    expect(handleClose).toHaveBeenCalledWith(false)
  })

  it('toggles single-character shortcuts setting and persists to localStorage', () => {
    render(<KeyboardShortcutsModal open={true} onOpenChange={() => {}} />)

    const checkbox = screen.getByRole('checkbox', {
      name: /enable single-character keybindings/i,
    }) as HTMLInputElement

    expect(checkbox.checked).toBe(true)

    // Toggle off
    fireEvent.click(checkbox)
    expect(checkbox.checked).toBe(false)
    expect(localStorage.getItem('refweir.character_shortcuts_enabled')).toBe('false')

    // Toggle on
    fireEvent.click(checkbox)
    expect(checkbox.checked).toBe(true)
    expect(localStorage.getItem('refweir.character_shortcuts_enabled')).toBe('true')
  })
})
