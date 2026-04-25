# Frontend Context Menu Pattern

## Goal

This document defines the standard implementation pattern for row-level right-click menus in the shared React UI.

It applies when:

- adding context menu support to resource lists
- adding row-level right-click actions to settings tables
- extending existing table actions into reusable context menu entries

## Design Principle

Right-click capability should be implemented as a reusable UI pattern, not as ad-hoc page logic.

When a list or table needs context menu actions:

1. Prefer the existing shared context menu primitives and wrappers.
2. Prefer wiring the capability at the shared table layer first.
3. Let each page provide only its row-specific menu items.
4. Do not duplicate right-click event handling in each page component unless the table is not using a shared table abstraction.

## Current Stack Choice

The current shared implementation is based on:

- `@radix-ui/react-context-menu`
- `ui/src/components/ui/context-menu.tsx`
- `ui/src/components/row-context-menu.tsx`

This should remain the default choice for future context menu work unless the user explicitly approves another component library or interaction model.

## Shared Entry Points

Current shared entry points are:

- `ui/src/components/resource-table.tsx`
  - shared resource list state, filtering, selection, toolbar, pagination
- `ui/src/components/resource-table-view.tsx`
  - shared row rendering entry for most resource lists
- `ui/src/components/action-table.tsx`
  - shared settings-style table with per-row actions

When adding context menus to resource lists, prefer `ResourceTable` / `ResourceTableView` first.

When adding context menus to settings tables, prefer `ActionTable` first.

Only fall back to `SimpleTable` or page-local table rendering when the table is not covered by the shared abstractions.

## Standard Data Shape

Shared row context menus should be driven by menu item data, not hard-coded render branches in page components.

Current shared shape:

```ts
type RowContextMenuItem<T> =
  | {
      key: string
      type?: 'item'
      label: React.ReactNode
      icon?: React.ReactNode
      shortcut?: string
      variant?: 'default' | 'destructive'
      hidden?: boolean | ((item: T) => boolean)
      disabled?: boolean | ((item: T) => boolean)
      onSelect: (item: T) => void | Promise<void>
    }
  | {
      key: string
      type: 'submenu'
      label: React.ReactNode
      icon?: React.ReactNode
      hidden?: boolean | ((item: T) => boolean)
      disabled?: boolean | ((item: T) => boolean)
      children: RowContextMenuItem<T>[]
    }
  | {
      key: string
      type: 'separator'
    }
```

`ResourceTable` exposes:

```ts
getRowContextMenuItems?: (item: T) => RowContextMenuItem<T>[]
```

Pages should pass menu definitions through this API instead of binding raw `onContextMenu` handlers.

## Default Usage Pattern

Recommended order:

1. Reuse existing `ResourceTable` or `ActionTable`
2. Provide `getRowContextMenuItems`
3. Keep page-level logic focused on business actions only

Typical menu items should prefer existing project capabilities:

- open details via React Router navigation
- copy values via `ui/src/lib/desktop.ts` and `copyTextToClipboard`
- destructive actions should reuse existing confirmation flows when available

## First-Class Reusable Actions

When a resource list has a detail page, common first actions are:

- view details
- copy resource name
- copy namespace when namespace-scoped

Resource-specific actions can then extend on top, such as:

- Pod: copy Pod IP
- Service: copy ClusterIP
- Node: copy IP

If multiple pages share the same menu items, extract a shared helper instead of repeating the same menu definitions.

## Interaction Rules

- Right-click menus are row-level actions by default.
- Batch actions should remain in the table toolbar unless there is an explicit product requirement to mix them into the row menu.
- Dangerous actions should use `destructive` styling.
- Hidden and disabled state should be expressed through menu item metadata, not manual render branching around the entire menu.

## What To Avoid

- Do not hand-roll floating menu positioning logic.
- Do not introduce another menu library when the current Radix-based stack is sufficient.
- Do not bind one-off right-click handlers directly in each page when the table already uses `ResourceTable` or `ActionTable`.
- Do not bypass `ui/src/lib/desktop.ts` for clipboard or desktop-sensitive actions when a shared helper already exists.

## Current Reference Implementations

Current reference examples:

- `ui/src/pages/pod-list-page.tsx`
- `ui/src/pages/service-list-page.tsx`
- `ui/src/pages/node-list-page.tsx`

These pages should be treated as the baseline pattern for future row context menu work.
