import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const {
  getDesktopUpdateState,
  checkDesktopUpdate,
  startDesktopUpdateDownload,
  applyDesktopUpdate,
} = vi.hoisted(() => ({
  getDesktopUpdateState: vi.fn(),
  checkDesktopUpdate: vi.fn(),
  startDesktopUpdateDownload: vi.fn(),
  applyDesktopUpdate: vi.fn(),
}))

const { trackEvent } = vi.hoisted(() => ({
  trackEvent: vi.fn(),
}))

vi.mock('@/contexts/runtime-context', () => ({
  useRuntime: vi.fn(() => ({
    isDesktop: true,
    isReady: true,
  })),
}))

vi.mock('@/lib/desktop', () => ({
  getDesktopUpdateState,
  checkDesktopUpdate,
  startDesktopUpdateDownload,
  applyDesktopUpdate,
  cancelDesktopUpdateDownload: vi.fn(),
  clearIgnoredDesktopUpdate: vi.fn(),
  ignoreDesktopUpdate: vi.fn(),
  retryDesktopUpdateDownload: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  checkVersionUpdate: vi.fn(),
}))

vi.mock('@/lib/analytics', () => ({
  trackEvent,
}))

import { useDesktopUpdate } from './use-desktop-update'

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
  }
}

describe('useDesktopUpdate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getDesktopUpdateState.mockResolvedValue({ ignoredVersion: '' })
    checkDesktopUpdate.mockResolvedValue({
      currentVersion: 'v1.0.0',
      latestVersion: 'v1.0.1',
      comparison: 'update_available',
      hasNewVersion: true,
      releaseUrl: '',
      releaseNotes: '',
      publishedAt: '',
      ignored: false,
      assetAvailable: true,
      checkedAt: '',
    })
    startDesktopUpdateDownload.mockResolvedValue({ ignoredVersion: '' })
    applyDesktopUpdate.mockResolvedValue(true)
  })

  it('tracks manual update actions with sanitized metadata', async () => {
    const { result } = renderHook(() => useDesktopUpdate(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => {
      expect(getDesktopUpdateState).toHaveBeenCalled()
    })

    act(() => {
      result.current.check(true)
      result.current.startDownload('v1.0.1')
      result.current.applyUpdate()
    })

    expect(trackEvent).toHaveBeenNthCalledWith(1, 'update_check_clicked', {
      runtime: 'desktop',
      page: 'overview',
    })
    expect(trackEvent).toHaveBeenNthCalledWith(2, 'update_download_started', {
      runtime: 'desktop',
      page: 'overview',
    })
    expect(trackEvent).toHaveBeenNthCalledWith(3, 'update_install_started', {
      runtime: 'desktop',
      page: 'overview',
    })
  })

  it('does not track silent background checks', async () => {
    const { result } = renderHook(() => useDesktopUpdate(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => {
      expect(getDesktopUpdateState).toHaveBeenCalled()
    })

    act(() => {
      result.current.check(false)
    })

    expect(trackEvent).not.toHaveBeenCalled()
  })
})
