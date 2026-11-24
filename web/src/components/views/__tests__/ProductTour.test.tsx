import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { ProductTour } from '../ProductTour'

describe('ProductTour Component - 3-Step First-Run Checklist', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('renders step 1 (Create Personal Access Token) by default with fresh state', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    expect(screen.getByRole('dialog', { name: /refweir quickstart checklist/i })).toBeInTheDocument()
    expect(screen.getByText('Create Personal Access Token')).toBeInTheDocument()
    expect(screen.getByText(/1\. Create Token/i)).toBeInTheDocument()
    expect(screen.getByText(/2\. Copy MCP Config/i)).toBeInTheDocument()
    expect(screen.getByText(/3\. Run Agent Loop/i)).toBeInTheDocument()
  })

  it('navigates through the 3 checklist steps using Next and Back buttons', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    // Step 1 -> Step 2
    const nextBtn = screen.getByRole('button', { name: /next/i })
    fireEvent.click(nextBtn)

    expect(screen.getAllByText('Copy MCP Config JSON').length).toBeGreaterThan(0)
    expect(screen.getByText('mcpServers Config (JSON)')).toBeInTheDocument()

    // Step 2 -> Step 3
    fireEvent.click(screen.getByRole('button', { name: /next/i }))
    expect(screen.getByText('Run Agent Loop on Seeded Demo Repo')).toBeInTheDocument()
    expect(screen.getByText('Execute Loop on Seeded Demo')).toBeInTheDocument()

    // Step 3 -> Back to Step 2
    const backBtn = screen.getByRole('button', { name: /back/i })
    fireEvent.click(backBtn)
    expect(screen.getAllByText('Copy MCP Config JSON').length).toBeGreaterThan(0)
  })

  it('navigates directly by clicking step tabs and toggles completed status', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    // Click step 3 tab directly
    const step3Tab = screen.getByText(/3\. Run Agent Loop/i)
    fireEvent.click(step3Tab)
    expect(screen.getByText('Run Agent Loop on Seeded Demo Repo')).toBeInTheDocument()

    // Toggle step done status
    const toggleDone = screen.getAllByTitle('Click to toggle step complete')[0]
    fireEvent.click(toggleDone)
  })

  it('persists tour completion to localStorage when clicking Get Started on last step', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    // Advance to Step 3
    fireEvent.click(screen.getByRole('button', { name: /next/i }))
    fireEvent.click(screen.getByRole('button', { name: /next/i }))

    const finishBtn = screen.getByRole('button', { name: /get started/i })
    fireEvent.click(finishBtn)

    expect(localStorage.getItem('refweir_onboarding_tour_completed')).toBe('true')
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('persists skip to localStorage and closes when clicking Skip', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    const skipBtn = screen.getByRole('button', { name: /skip/i })
    fireEvent.click(skipBtn)

    expect(localStorage.getItem('refweir_onboarding_tour_completed')).toBe('true')
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('copies MCP configuration JSON to clipboard', () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    })

    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    // Navigate to step 2
    fireEvent.click(screen.getByRole('button', { name: /next/i }))

    const copyBtn = screen.getByRole('button', { name: /copy mcp config json/i })
    fireEvent.click(copyBtn)

    expect(writeTextMock).toHaveBeenCalledWith(expect.stringContaining('"refweir": {'))
  })

  it('contains zero occurrences of Sovereign wording', () => {
    const onClose = vi.fn()
    const { container } = render(<ProductTour isOpen={true} onClose={onClose} />)

    expect(container.textContent).not.toMatch(/sovereign/i)
  })
})
