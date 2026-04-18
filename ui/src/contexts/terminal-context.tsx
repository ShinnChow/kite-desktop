import { createContext, ReactNode, useContext, useState } from 'react'

import { trackEvent } from '@/lib/analytics'
import { getCurrentAnalyticsPageKey } from '@/lib/analytics-route'

interface TerminalContextType {
  isOpen: boolean
  isMinimized: boolean
  openTerminal: (entry?: string) => void
  closeTerminal: () => void
  minimizeTerminal: () => void
  toggleTerminal: (entry?: string) => void
}

const TerminalContext = createContext<TerminalContextType | undefined>(
  undefined
)

export function TerminalProvider({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)

  // Open (or un-minimize) the terminal
  const openTerminal = (entry: string = 'button') => {
    if (!isOpen) {
      trackEvent('kubectl_terminal_open', {
        runtime: 'desktop',
        entry,
        page: getCurrentAnalyticsPageKey(),
      })
    }
    setIsOpen(true)
    setIsMinimized(false)
  }

  // Fully close and destroy the session
  const closeTerminal = () => {
    setIsOpen(false)
    setIsMinimized(false)
  }

  // Hide the panel but keep the session alive
  const minimizeTerminal = () => {
    setIsMinimized(true)
  }

  // Toggle open/minimized
  const toggleTerminal = (entry: string = 'button') => {
    if (!isOpen) {
      openTerminal(entry)
    } else if (isMinimized) {
      setIsMinimized(false)
    } else {
      minimizeTerminal()
    }
  }

  return (
    <TerminalContext.Provider
      value={{
        isOpen,
        isMinimized,
        openTerminal,
        closeTerminal,
        minimizeTerminal,
        toggleTerminal,
      }}
    >
      {children}
    </TerminalContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useTerminal() {
  const context = useContext(TerminalContext)
  if (context === undefined) {
    throw new Error('useTerminal must be used within a TerminalProvider')
  }
  return context
}
