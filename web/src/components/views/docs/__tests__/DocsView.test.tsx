import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { DocsView } from '../DocsView'

describe('DocsView Component (GitHub Docs Architecture)', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    window.history.replaceState({}, '', '/')
    // Mock navigator.clipboard
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockImplementation(() => Promise.resolve()),
      },
    })
    // Mock window.scrollTo
    window.scrollTo = vi.fn()
  })

  it('renders documentation header, version badge, search bar, and breadcrumbs', () => {
    render(<DocsView />)

    expect(screen.getByText('Refweir Documentation')).toBeInTheDocument()
    expect(screen.getByText(/v0.6.0-beta/i)).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/Search docs or articles/i)).toBeInTheDocument()
    expect(screen.getByText('Refweir Docs')).toBeInTheDocument()
  })

  it('renders default quickstart article with code snippets and copy button', async () => {
    render(<DocsView initialArticleId="quickstart" />)

    expect(screen.getByRole('heading', { level: 1, name: /Quickstart: 5-Minute Bootstrap/i })).toBeInTheDocument()
    expect(screen.getByText(/Get a production-ready Refweir instance/i)).toBeInTheDocument()
    expect(screen.getByText(/refweird bootstrap --email/i)).toBeInTheDocument()

    const copyBtn = screen.getByRole('button', { name: /copy code/i })
    fireEvent.click(copyBtn)
    expect(navigator.clipboard.writeText).toHaveBeenCalled()
    await waitFor(() => {
      expect(screen.getByText(/Copied!/i)).toBeInTheDocument()
    })
  })

  it('switches between code tabs (Step 1, Step 2, Step 3)', () => {
    render(<DocsView initialArticleId="quickstart" />)

    const step2Tab = screen.getByRole('button', { name: /Step 2: Git Clone/i })
    fireEvent.click(step2Tab)
    expect(screen.getByText(/git clone ssh:\/\/git@localhost:2222/i)).toBeInTheDocument()

    const step3Tab = screen.getByRole('button', { name: /Step 3: Claude Code/i })
    fireEvent.click(step3Tab)
    expect(screen.getByText(/"mcpServers":/i)).toBeInTheDocument()
  })

  it('navigates to different articles using the sidebar', async () => {
    render(<DocsView />)

    const leasesBtn = screen.getByRole('button', { name: /Capability Leases & Time-Bound Scopes/i })
    fireEvent.click(leasesBtn)

    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1, name: /Capability Leases & Time-Bound Scopes/i })).toBeInTheDocument()
    })
    expect(screen.getByText(/Cryptographic HMAC-SHA256 leases/i)).toBeInTheDocument()
  })

  it('supports instant search filtering across articles and highlights matches', async () => {
    render(<DocsView />)

    const searchInput = screen.getByPlaceholderText(/Search docs or articles/i)
    fireEvent.change(searchInput, { target: { value: 'Ruleset' } })

    await waitFor(() => {
      expect(screen.getByText(/Search Results/i)).toBeInTheDocument()
      expect(screen.getAllByText(/Branch Rulesets & Push Protection/i).length).toBeGreaterThan(0)
    })
  })

  it('submits article feedback with helpful feedback widget', async () => {
    render(<DocsView />)

    expect(screen.getByText('Did this article help you?')).toBeInTheDocument()
    const yesBtn = screen.getByRole('button', { name: /Yes/i })
    fireEvent.click(yesBtn)

    await waitFor(() => {
      expect(screen.getByText(/Thank you for your feedback!/i)).toBeInTheDocument()
    })
  })

  it('navigates via Previous and Next article buttons', async () => {
    render(<DocsView initialArticleId="overview" />)

    const nextBtn = screen.getByRole('button', { name: /Next/i })
    expect(nextBtn).toBeInTheDocument()
    fireEvent.click(nextBtn)

    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1, name: /Quickstart: 5-Minute Bootstrap/i })).toBeInTheDocument()
    })
  })
})
