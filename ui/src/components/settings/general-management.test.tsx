import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { useGeneralSetting, updateGeneralSetting } = vi.hoisted(() => ({
  useGeneralSetting: vi.fn(),
  updateGeneralSetting: vi.fn(),
}))

const { successToast, errorToast } = vi.hoisted(() => ({
  successToast: vi.fn(),
  errorToast: vi.fn(),
}))

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>()
  return {
    ...actual,
    useTranslation: () => ({
      t: (
        key: string,
        fallbackOrOptions?: string | { defaultValue?: string; version?: string }
      ) => {
        if (typeof fallbackOrOptions === 'string') {
          return fallbackOrOptions
        }
        if (fallbackOrOptions?.defaultValue) {
          return fallbackOrOptions.defaultValue.replace(
            '{{version}}',
            fallbackOrOptions.version ?? ''
          )
        }
        return key
      },
    }),
  }
})

vi.mock('@/contexts/runtime-context', () => ({
  useRuntime: vi.fn(() => ({
    isDesktop: true,
  })),
}))

vi.mock('@/lib/api', () => ({
  useGeneralSetting,
  updateGeneralSetting,
}))

vi.mock('sonner', () => ({
  toast: {
    success: successToast,
    error: errorToast,
  },
}))

import * as GeneralManagementModule from './general-management'

function renderComponent() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <GeneralManagementModule.GeneralManagement />
    </QueryClientProvider>
  )
}

describe('GeneralManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    useGeneralSetting.mockReturnValue({
      data: {
        aiAgentEnabled: false,
        aiProvider: 'openai',
        aiModel: 'gpt-4o-mini',
        aiApiKey: '',
        aiApiKeyConfigured: false,
        aiBaseUrl: '',
        aiMaxTokens: 4096,
        aiChatHistorySessionLimit: 200,
        aiChatOpenMode: 'sidecar',
        kubectlEnabled: true,
        kubectlImage: 'docker.cnb.cool/znb/images/kubectl:latest',
        nodeTerminalImage: 'docker.cnb.cool/znb/images/busybox:latest',
        enableAnalytics: true,
        enableVersionCheck: true,
        updateSource: 'auto',
      },
      isLoading: false,
    })
  })

  it('reloads the app after analytics is disabled so the injected script is removed', async () => {
    const user = userEvent.setup()
    const reloadSpy = vi
      .spyOn(GeneralManagementModule.browserRuntime, 'reloadWindow')
      .mockImplementation(() => undefined)

    updateGeneralSetting.mockResolvedValue({
      aiAgentEnabled: false,
      aiProvider: 'openai',
      aiModel: 'gpt-4o-mini',
      aiApiKey: '',
      aiApiKeyConfigured: false,
      aiBaseUrl: '',
      aiMaxTokens: 4096,
      aiChatHistorySessionLimit: 200,
      aiChatOpenMode: 'sidecar',
      kubectlEnabled: true,
      kubectlImage: 'docker.cnb.cool/znb/images/kubectl:latest',
      nodeTerminalImage: 'docker.cnb.cool/znb/images/busybox:latest',
      enableAnalytics: false,
      enableVersionCheck: true,
      updateSource: 'auto',
    })

    renderComponent()

    await user.click(screen.getByRole('switch', { name: 'Enable analytics' }))
    await user.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(updateGeneralSetting).toHaveBeenCalledWith(
        expect.objectContaining({ enableAnalytics: false })
      )
    })
    await waitFor(() => {
      expect(reloadSpy).toHaveBeenCalledTimes(1)
    })
  })

  it('does not reload when analytics setting is unchanged', async () => {
    const user = userEvent.setup()
    const reloadSpy = vi
      .spyOn(GeneralManagementModule.browserRuntime, 'reloadWindow')
      .mockImplementation(() => undefined)

    updateGeneralSetting.mockResolvedValue({
      aiAgentEnabled: false,
      aiProvider: 'openai',
      aiModel: 'gpt-4o-mini',
      aiApiKey: '',
      aiApiKeyConfigured: false,
      aiBaseUrl: '',
      aiMaxTokens: 4096,
      aiChatHistorySessionLimit: 200,
      aiChatOpenMode: 'sidecar',
      kubectlEnabled: true,
      kubectlImage: 'docker.cnb.cool/znb/images/kubectl:latest',
      nodeTerminalImage: 'docker.cnb.cool/znb/images/busybox:latest',
      enableAnalytics: true,
      enableVersionCheck: true,
      updateSource: 'auto',
    })

    renderComponent()

    await user.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(updateGeneralSetting).toHaveBeenCalledWith(
        expect.objectContaining({ enableAnalytics: true })
      )
    })
    expect(reloadSpy).not.toHaveBeenCalled()
    expect(successToast).toHaveBeenCalled()
  })
})

describe('shouldReloadForAnalyticsChange', () => {
  it('only reloads when the analytics toggle actually changes', () => {
    expect(
      GeneralManagementModule.shouldReloadForAnalyticsChange(true, false)
    ).toBe(true)
    expect(
      GeneralManagementModule.shouldReloadForAnalyticsChange(false, true)
    ).toBe(true)
    expect(
      GeneralManagementModule.shouldReloadForAnalyticsChange(true, true)
    ).toBe(false)
    expect(
      GeneralManagementModule.shouldReloadForAnalyticsChange(undefined, false)
    ).toBe(false)
  })
})
