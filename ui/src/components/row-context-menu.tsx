import { ReactNode } from 'react'

import {
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
} from '@/components/ui/context-menu'

type Predicate<T> = boolean | ((item: T) => boolean)

interface RowContextMenuBaseItem<T> {
  key: string
  icon?: ReactNode
  shortcut?: string
  hidden?: Predicate<T>
  disabled?: Predicate<T>
}

export interface RowContextMenuActionItem<T> extends RowContextMenuBaseItem<T> {
  type?: 'item'
  label: ReactNode
  variant?: 'default' | 'destructive'
  onSelect: (item: T) => void | Promise<void>
}

export interface RowContextMenuSubmenuItem<
  T,
> extends RowContextMenuBaseItem<T> {
  type: 'submenu'
  label: ReactNode
  children: RowContextMenuItem<T>[]
}

export interface RowContextMenuSeparatorItem {
  type: 'separator'
  key: string
}

export type RowContextMenuItem<T> =
  | RowContextMenuActionItem<T>
  | RowContextMenuSeparatorItem
  | RowContextMenuSubmenuItem<T>

function resolvePredicate<T>(value: Predicate<T> | undefined, item: T) {
  if (typeof value === 'function') {
    return value(item)
  }
  return Boolean(value)
}

function getVisibleItems<T>(items: RowContextMenuItem<T>[], item: T) {
  return items.filter((menuItem) => {
    if (menuItem.type === 'separator') {
      return true
    }
    return !resolvePredicate(menuItem.hidden, item)
  })
}

function renderMenuItem<T>(menuItem: RowContextMenuItem<T>, item: T) {
  if (menuItem.type === 'separator') {
    return <ContextMenuSeparator key={menuItem.key} />
  }

  if (menuItem.type === 'submenu') {
    const visibleChildren = getVisibleItems(menuItem.children, item)
    if (visibleChildren.length === 0) {
      return null
    }

    return (
      <ContextMenuSub key={menuItem.key}>
        <ContextMenuSubTrigger
          disabled={resolvePredicate(menuItem.disabled, item)}
        >
          {menuItem.icon}
          {menuItem.label}
        </ContextMenuSubTrigger>
        <ContextMenuSubContent>
          {visibleChildren.map((child) => renderMenuItem(child, item))}
        </ContextMenuSubContent>
      </ContextMenuSub>
    )
  }

  return (
    <ContextMenuItem
      key={menuItem.key}
      variant={menuItem.variant}
      disabled={resolvePredicate(menuItem.disabled, item)}
      onSelect={() => {
        void menuItem.onSelect(item)
      }}
    >
      {menuItem.icon}
      {menuItem.label}
      {menuItem.shortcut ? (
        <ContextMenuShortcut>{menuItem.shortcut}</ContextMenuShortcut>
      ) : null}
    </ContextMenuItem>
  )
}

interface RowContextMenuContentProps<T> {
  item: T
  items: RowContextMenuItem<T>[]
}

export function RowContextMenuContentRenderer<T>({
  item,
  items,
}: RowContextMenuContentProps<T>) {
  const visibleItems = getVisibleItems(items, item)
  const normalizedItems = visibleItems.filter((menuItem, index) => {
    if (menuItem.type !== 'separator') {
      return true
    }

    const prev = visibleItems[index - 1]
    const next = visibleItems[index + 1]
    return Boolean(
      prev && next && prev.type !== 'separator' && next.type !== 'separator'
    )
  })

  if (normalizedItems.length === 0) {
    return null
  }

  return (
    <ContextMenuContent>
      {normalizedItems.map((menuItem) => renderMenuItem(menuItem, item))}
    </ContextMenuContent>
  )
}
