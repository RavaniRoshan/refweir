import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ClusterHealth } from '../admin/ClusterHealth'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api')
  return {
    ...actual,
    getClusterMetrics: vi.fn(),
    triggerClusterCheckpoint: vi.fn(),
    triggerClusterSnapshot: vi.fn(),
    getAdminPreflight: vi.fn(),
  }
})

describe('ClusterHealth Admin View', () => {
  const mockMetrics: api.ClusterMetrics = {
    serverStatus: 'healthy',
    uptime: '420s',
    version: 'refweird v0.4.5-enterprise',
    activeGitSessions: 2,
    sqliteWalSizeBytes: 1048576,
    databasePath: 'data/meta.db',
    goroutines: 12,
    allocatedMemoryMb: 14.5,
    cpuUsagePercent: 0.5,
    litestreamSyncStatus: 'active',
    lastBackupTimestamp: '2026-10-15T11:00:00Z',
    mcpMountedToolsCount: 35,
  }

  const mockPreflight: api.PreflightReport = {
    status: 'healthy',
    evaluated_at: '2026-11-24T12:00:00Z',
    checks: [
      {
        id: 'sqlite_wal_mode',
        name: 'SQLite WAL Journal Mode',
        status: 'pass',
        message: 'WAL journal mode active for concurrent reader/writer operations',
      },
      {
        id: 'sqlite_integrity',
        name: 'Database Structural Integrity',
        status: 'pass',
        message: 'PRAGMA quick_check passed: 0 B-tree corruptions detected',
      },
    ],
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getClusterMetrics).mockResolvedValue(mockMetrics)
    vi.mocked(api.getAdminPreflight).mockResolvedValue(mockPreflight)
  })

  it('renders loading state then displays cluster telemetry metrics', async () => {
    render(<ClusterHealth />)

    expect(screen.getByText(/Loading cluster health telemetry/i)).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    expect(screen.getByText(/refweird v0.4.5-enterprise/i)).toBeInTheDocument()
    expect(screen.getByText(/System Operational/i)).toBeInTheDocument()
    expect(screen.getByText('1.0 MB')).toBeInTheDocument() // formatBytes(1048576)
    expect(screen.getByText('Continuous')).toBeInTheDocument()
    expect(screen.getByText('14.5')).toBeInTheDocument()
    expect(screen.getByText('MB RAM')).toBeInTheDocument()
  })

  it('refreshes telemetry when clicking Refresh button', async () => {
    render(<ClusterHealth />)

    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    const refreshBtn = screen.getByRole('button', { name: /Refresh Telemetry/i })
    fireEvent.click(refreshBtn)

    expect(api.getClusterMetrics).toHaveBeenCalledTimes(2)
  })

  it('copies command snippet to clipboard', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    })

    render(<ClusterHealth />)

    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    const copyBtns = screen.getAllByRole('button', { name: /Copy/i })
    if (copyBtns.length > 0) {
      fireEvent.click(copyBtns[0])
      expect(writeTextMock).toHaveBeenCalled()
    }
  })

  it('renders historical telemetry chart and switches chart modes', async () => {
    render(<ClusterHealth />)

    await waitFor(() => {
      expect(screen.getByText('Cluster Telemetry & Historical Trends')).toBeInTheDocument()
    })

    expect(screen.getByTestId('telemetry-chart')).toBeInTheDocument()

    // Switch to Engine Overhead
    const resourcesBtn = screen.getByRole('button', { name: /Engine Overhead/i })
    fireEvent.click(resourcesBtn)
    expect(screen.getByTestId('telemetry-chart')).toBeInTheDocument()

    // Switch back to Throughput
    const throughputBtn = screen.getByRole('button', { name: /Throughput/i })
    fireEvent.click(throughputBtn)
    expect(screen.getByTestId('telemetry-chart')).toBeInTheDocument()
  })

  it('triggers on-demand WAL checkpoint and displays confirmation badge', async () => {
    vi.mocked(api.triggerClusterCheckpoint).mockResolvedValueOnce({
      status: 'checkpointed',
      mode: 'PASSIVE',
    })

    render(<ClusterHealth />)
    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    const checkpointBtn = screen.getByRole('button', { name: /Run WAL Checkpoint/i })
    fireEvent.click(checkpointBtn)

    await waitFor(() => {
      expect(screen.getByText(/WAL Checkpointed/i)).toBeInTheDocument()
    })
    expect(api.triggerClusterCheckpoint).toHaveBeenCalledTimes(1)
  })

  it('triggers atomic snapshot and displays filename and size', async () => {
    vi.mocked(api.triggerClusterSnapshot).mockResolvedValueOnce({
      filename: 'snapshot_000001_1795504266.db',
      walSequence: 42,
      checksum: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
      sizeBytes: 823296,
      createdAt: '2026-11-24T07:11:06Z',
    })

    render(<ClusterHealth />)
    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    const snapshotBtn = screen.getByRole('button', { name: /Capture Atomic Snapshot/i })
    fireEvent.click(snapshotBtn)

    await waitFor(() => {
      expect(screen.getByText(/snapshot_000001_1795504266.db/i)).toBeInTheDocument()
    })
    expect(api.triggerClusterSnapshot).toHaveBeenCalledTimes(1)
  })

  it('renders self-host production pre-flight diagnostics and re-evaluates on click', async () => {
    render(<ClusterHealth />)
    await waitFor(() => {
      expect(screen.getByText('Self-Host Production Pre-flight Diagnostics')).toBeInTheDocument()
    })

    expect(screen.getByText('SQLite WAL Journal Mode')).toBeInTheDocument()
    expect(screen.getByText('Database Structural Integrity')).toBeInTheDocument()
    expect(screen.getByText('ALL CHECKS PASSED')).toBeInTheDocument()

    const diagBtn = screen.getByRole('button', { name: /Run Diagnostics/i })
    fireEvent.click(diagBtn)

    expect(api.getAdminPreflight).toHaveBeenCalledTimes(2)
  })
})
