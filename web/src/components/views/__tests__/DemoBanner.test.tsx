import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { DemoBanner } from '../DemoBanner'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    runDemoAgent: vi.fn(),
    mergeDemoPullRequest: vi.fn(),
  }
})

const mockedRun = vi.mocked(api.runDemoAgent)
const mockedMerge = vi.mocked(api.mergeDemoPullRequest)

describe('DemoBanner', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the demo notice, self-host CTA, and run button', () => {
    render(<DemoBanner repo="demo/agent-lab" />)
    expect(screen.getByText(/live demo/i)).toBeTruthy()
    expect(screen.getByText(/demo\/agent-lab/)).toBeTruthy()
    expect(screen.getByRole('link', { name: /self-host refweir/i })).toHaveAttribute(
      'href',
      'https://github.com/RavaniRoshan/refweir#quickstart'
    )
    expect(screen.getByRole('button', { name: /try an agent run/i })).toBeTruthy()
  })

  it('runs a real agent and links to the resulting pull request', async () => {
    mockedRun.mockResolvedValue({
      ok: true,
      agent: 'demo-agent',
      leaseId: 'lease_123',
      expiresAt: '2026-11-09T12:10:00Z',
      branch: 'agent/demo-20261109-120000',
      title: 'docs: add an agent-authored run note',
      prNumber: 7,
      prUrl: '/?tab=prs',
      files: ['notes/agent-runs/20261109-120000.md'],
      auditTail: [],
    })

    render(<DemoBanner repo="demo/agent-lab" />)
    fireEvent.click(screen.getByRole('button', { name: /try an agent run/i }))

    await waitFor(() => expect(mockedRun).toHaveBeenCalledTimes(1))
    await waitFor(() => expect(screen.getByText(/PR #7 opened/i)).toBeTruthy())
    expect(screen.getByRole('link', { name: /review the pr/i })).toHaveAttribute(
      'href',
      '/?tab=prs'
    )
  })

  it('merges the resulting agent PR through the demo endpoint', async () => {
    mockedRun.mockResolvedValue({
      ok: true,
      agent: 'demo-agent',
      leaseId: 'lease_123',
      expiresAt: '2026-11-09T12:10:00Z',
      branch: 'agent/demo-20261109-120000',
      title: 'docs: add an agent-authored run note',
      prNumber: 7,
      prUrl: '/?tab=prs',
      files: ['notes/agent-runs/20261109-120000.md'],
      auditTail: [],
    })
    mockedMerge.mockResolvedValue({
      ok: true,
      prNumber: 7,
      mergeCommitSha: 'abcdef1234567890',
      baseRef: 'main',
    })

    render(<DemoBanner repo="demo/agent-lab" />)
    fireEvent.click(screen.getByRole('button', { name: /try an agent run/i }))
    await waitFor(() => expect(screen.getByText(/PR #7 opened/i)).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /merge the demo pull request/i }))
    await waitFor(() => expect(mockedMerge).toHaveBeenCalledWith(7))
    await waitFor(() => expect(screen.getByText(/Merged PR #7 into/i)).toBeTruthy())
    expect(screen.getByText(/abcdef1/)).toBeTruthy()
  })

  it('surfaces the server error without fabricating a result', async () => {
    mockedRun.mockRejectedValue(new Error('The demo allows one agent run every 20 seconds.'))

    render(<DemoBanner />)
    fireEvent.click(screen.getByRole('button', { name: /try an agent run/i }))

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent(
        'The demo allows one agent run every 20 seconds.'
      )
    )
    expect(screen.queryByText(/PR #/)).toBeNull()
  })
})
