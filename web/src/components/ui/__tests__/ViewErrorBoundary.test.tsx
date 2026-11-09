import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { ViewErrorBoundary } from '../ViewErrorBoundary'

const ThrowingComponent = () => {
  throw new Error('Chunk load failed: /assets/PullRequestsView-oldhash.js')
}

const SafeComponent = () => <div>Normal View Content</div>

describe('ViewErrorBoundary Component', () => {
  it('renders children when no error occurs', () => {
    render(
      <ViewErrorBoundary>
        <SafeComponent />
      </ViewErrorBoundary>
    )

    expect(screen.getByText('Normal View Content')).toBeInTheDocument()
  })

  it('catches render errors and displays Primer error card with reload button', () => {
    // Suppress console.error in this test since React logs boundary errors
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})

    render(
      <ViewErrorBoundary>
        <ThrowingComponent />
      </ViewErrorBoundary>
    )

    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.getByText('Failed to load view')).toBeInTheDocument()
    expect(screen.getByText(/A new version of Refweir may have been deployed/i)).toBeInTheDocument()
    const reloadBtn = screen.getByRole('button', { name: /reload page/i })
    const tryAgainBtn = screen.getByRole('button', { name: /try again/i })
    expect(reloadBtn).toBeInTheDocument()
    expect(tryAgainBtn).toBeInTheDocument()

    // Test clicking try again resets error state
    fireEvent.click(tryAgainBtn)

    spy.mockRestore()
  })
})
