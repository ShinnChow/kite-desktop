import { getAnalyticsPageUrl } from './analytics-route'

type AnalyticsValue = string | number | boolean

type UmamiPayload = {
  url?: string
  title?: string
  name?: string
  data?: Record<string, AnalyticsValue>
}

type UmamiPayloadTransformer = (payload: UmamiPayload) => UmamiPayload

type UmamiTracker = {
  track: (payload?: UmamiPayload | UmamiPayloadTransformer) => void
}

declare global {
  interface Window {
    __kite_analytics_enabled__?: boolean
    umami?: UmamiTracker
  }
}

function isWindowAvailable() {
  return typeof window !== 'undefined'
}

function getUmamiTracker() {
  if (!isWindowAvailable()) {
    return null
  }
  return typeof window.umami?.track === 'function' ? window.umami : null
}

export function isAnalyticsEnabled() {
  return isWindowAvailable() && window.__kite_analytics_enabled__ === true
}

export function trackPage(pageKey: string) {
  if (!isAnalyticsEnabled()) {
    return
  }

  const tracker = getUmamiTracker()
  if (!tracker) {
    return
  }

  const pageUrl = getAnalyticsPageUrl(pageKey)
  tracker.track((payload) => ({
    ...payload,
    url: pageUrl,
    title: `Kite Desktop - ${pageKey}`,
  }))
}

export function trackEvent(
  name: string,
  data?: Record<string, AnalyticsValue>
) {
  if (!isAnalyticsEnabled()) {
    return
  }

  const tracker = getUmamiTracker()
  if (!tracker) {
    return
  }

  tracker.track({
    name,
    ...(data ? { data } : {}),
  })
}
