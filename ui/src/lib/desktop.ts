import { withSubPath } from './subpath'

export interface DesktopWindowOptions {
  title?: string
  width?: number
  height?: number
  minWidth?: number
  minHeight?: number
}

export interface NativeFileFilter {
  displayName: string
  pattern: string
}

export interface NativeFileOptions {
  title?: string
  message?: string
  buttonText?: string
  directory?: string
  readContent?: boolean
  filters?: NativeFileFilter[]
}

export interface NativeFileSelection {
  canceled: boolean
  path?: string
  name?: string
  content?: string
}

let desktopModePromise: Promise<boolean> | null = null

export function isDesktopMode(): Promise<boolean> {
  if (!desktopModePromise) {
    desktopModePromise = fetchDesktopMode()
  }
  return desktopModePromise
}

export async function openURL(
  url: string,
  options: DesktopWindowOptions = {}
): Promise<void> {
  try {
    if (await isDesktopMode()) {
      await postDesktop('/api/desktop/open-url', {
        url,
        ...options,
      })
      return
    }
  } catch (error) {
    console.error('Desktop open-url failed:', error)
  }

  window.open(url, '_blank', 'noopener,noreferrer')
}

export async function openNativeFile(
  options: NativeFileOptions = {}
): Promise<NativeFileSelection | null> {
  if (!(await isDesktopMode())) {
    return null
  }

  return postDesktop<NativeFileSelection>('/api/desktop/open-file', {
    readContent: true,
    ...options,
  })
}

export function installDesktopTargetBlankInterceptor(): () => void {
  let cleanup = () => {}
  let active = true

  void isDesktopMode().then((enabled) => {
    if (!active || !enabled) {
      return
    }

    const handleClick = (event: MouseEvent) => {
      if (
        event.defaultPrevented ||
        event.button !== 0 ||
        event.metaKey ||
        event.ctrlKey ||
        event.shiftKey ||
        event.altKey
      ) {
        return
      }

      const target = event.target
      if (!(target instanceof Element)) {
        return
      }

      const anchor = target.closest('a[target="_blank"]')
      if (!(anchor instanceof HTMLAnchorElement)) {
        return
      }

      const href = anchor.href || anchor.getAttribute('href')
      if (!href) {
        return
      }

      event.preventDefault()
      void openURL(href)
    }

    document.addEventListener('click', handleClick, true)
    cleanup = () => {
      document.removeEventListener('click', handleClick, true)
    }
  })

  return () => {
    active = false
    cleanup()
  }
}

async function fetchDesktopMode(): Promise<boolean> {
  try {
    const response = await fetch(withSubPath('/api/desktop/status'), {
      credentials: 'include',
    })
    if (!response.ok) {
      return false
    }

    const data = (await response.json().catch(() => ({}))) as {
      enabled?: boolean
    }
    return data.enabled === true
  } catch {
    return false
  }
}

async function postDesktop<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(withSubPath(path), {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })

  if (!response.ok) {
    const error = (await response.json().catch(() => ({}))) as {
      error?: string
    }
    throw new Error(error.error || `Desktop request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}
