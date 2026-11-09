import '@testing-library/jest-dom/vitest'
import { vi, afterEach } from 'vitest'
import { configure } from '@testing-library/react'
import { queryClient } from '@/lib/queryClient'

configure({ asyncUtilTimeout: 4000 })

afterEach(() => {
  queryClient.clear()
})

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock navigator.clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockImplementation(() => Promise.resolve()),
  },
})
