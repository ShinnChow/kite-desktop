import { beforeEach, describe, expect, it, vi } from 'vitest'

import { isAnalyticsEnabled, trackEvent, trackPage } from './analytics'

describe('analytics', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    window.__kite_analytics_enabled__ = false
    delete window.umami
  })

  it('reports whether analytics is enabled from the injected runtime flag', () => {
    expect(isAnalyticsEnabled()).toBe(false)
    window.__kite_analytics_enabled__ = true
    expect(isAnalyticsEnabled()).toBe(true)
  })

  it('tracks a sanitized page view when analytics is enabled', () => {
    const track = vi.fn()
    window.__kite_analytics_enabled__ = true
    window.umami = { track }

    trackPage('pods/detail')

    expect(track).toHaveBeenCalledTimes(1)
    const transformer = track.mock.calls[0][0] as (payload: {
      url?: string
      title?: string
    }) => { url?: string; title?: string }
    expect(transformer({ url: '/pods/default/nginx' })).toEqual({
      url: '/pods/detail',
      title: 'Kite Desktop - pods/detail',
    })
  })

  it('tracks custom events only when analytics is enabled', () => {
    const track = vi.fn()
    window.umami = { track }

    trackEvent('cluster_switch', { runtime: 'desktop' })
    expect(track).not.toHaveBeenCalled()

    window.__kite_analytics_enabled__ = true
    trackEvent('cluster_switch', { runtime: 'desktop' })

    expect(track).toHaveBeenCalledWith({
      name: 'cluster_switch',
      data: { runtime: 'desktop' },
    })
  })
})
