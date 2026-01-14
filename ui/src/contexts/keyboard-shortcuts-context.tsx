/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"

export type ShortcutDefinition = {
  id: string
  label: string
  category?: string
  defaultKeys: string[]
}

type KeyboardShortcutsContextType = {
  shortcuts: Record<string, ShortcutDefinition>
  overrides: Record<string, string[]>
  register: (def: ShortcutDefinition) => void
  unregister: (id: string) => void
  updateOverride: (id: string, keys: string[]) => void
  resetOverride: (id: string) => void
  getKeys: (id: string) => string[]
}

const KeyboardShortcutsContext = React.createContext<KeyboardShortcutsContextType | null>(null)

const STORAGE_KEY = "mcp_any_shortcut_overrides"

export function KeyboardShortcutsProvider({ children }: { children: React.ReactNode }) {
  const [shortcuts, setShortcuts] = React.useState<Record<string, ShortcutDefinition>>({})
  const [overrides, setOverrides] = React.useState<Record<string, string[]>>({})
  const [isLoaded, setIsLoaded] = React.useState(false)

  // Load overrides from local storage
  React.useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        setOverrides(JSON.parse(stored))
      }
    } catch (e) {
      console.error("Failed to load shortcut overrides", e)
    }
    setIsLoaded(true)
  }, [])

  // Save overrides to local storage
  React.useEffect(() => {
    if (!isLoaded) return
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(overrides))
    } catch (e) {
      console.error("Failed to save shortcut overrides", e)
    }
  }, [overrides, isLoaded])

  const register = React.useCallback((def: ShortcutDefinition) => {
    setShortcuts((prev) => {
      if (JSON.stringify(prev[def.id]) === JSON.stringify(def)) {
        return prev
      }
      return { ...prev, [def.id]: def }
    })
  }, [])

  const unregister = React.useCallback((id: string) => {
    setShortcuts((prev) => {
      const next = { ...prev }
      delete next[id]
      return next
    })
  }, [])

  const updateOverride = React.useCallback((id: string, keys: string[]) => {
    setOverrides((prev) => ({ ...prev, [id]: keys }))
  }, [])

  const resetOverride = React.useCallback((id: string) => {
    setOverrides((prev) => {
      const next = { ...prev }
      delete next[id]
      return next
    })
  }, [])

  const getKeys = React.useCallback((id: string) => {
    return overrides[id] || shortcuts[id]?.defaultKeys || []
  }, [overrides, shortcuts])

  const value = React.useMemo(() => ({
    shortcuts,
    overrides,
    register,
    unregister,
    updateOverride,
    resetOverride,
    getKeys
  }), [shortcuts, overrides, register, unregister, updateOverride, resetOverride, getKeys])

  return (
    <KeyboardShortcutsContext.Provider value={value}>
      {children}
    </KeyboardShortcutsContext.Provider>
  )
}

export function useKeyboardShortcuts() {
  const context = React.useContext(KeyboardShortcutsContext)
  if (!context) {
    throw new Error("useKeyboardShortcuts must be used within a KeyboardShortcutsProvider")
  }
  return context
}

// Utility to match event against key definition
function matchesKey(event: KeyboardEvent, keyDef: string): boolean {
  const parts = keyDef.toLowerCase().split("+")
  const key = parts.pop()
  const modifiers = parts

  if (!key) return false

  const meta = modifiers.includes("meta") || modifiers.includes("cmd") || modifiers.includes("command")
  const ctrl = modifiers.includes("ctrl") || modifiers.includes("control")
  const alt = modifiers.includes("alt") || modifiers.includes("option")
  const shift = modifiers.includes("shift")

  if (event.metaKey !== meta) return false
  if (event.ctrlKey !== ctrl) return false
  if (event.altKey !== alt) return false
  if (event.shiftKey !== shift) return false

  // Handle special keys
  if (key === "space") return event.code === "Space"

  return event.key.toLowerCase() === key
}

export function useShortcut(
  id: string,
  defaultKeys: string[],
  action: (e: KeyboardEvent) => void,
  options: { label: string; category?: string; enabled?: boolean } = { label: "", enabled: true }
) {
  const { register, unregister, getKeys } = useKeyboardShortcuts()

  // Register on mount
  React.useEffect(() => {
    register({
      id,
      label: options.label,
      category: options.category,
      defaultKeys
    })
    // Don't unregister on unmount immediately if we want to keep it in the list?
    // But for now, let's unregister to keep the list clean with only active shortcuts.
    return () => {
      unregister(id)
    }
  }, [id, JSON.stringify(defaultKeys), options.label, options.category, register, unregister])

  // Listen for keys
  React.useEffect(() => {
    if (options.enabled === false) return

    const handleKeyDown = (event: KeyboardEvent) => {
      const activeKeys = getKeys(id)
      // Check if any of the active key combos match
      if (activeKeys.some(k => matchesKey(event, k))) {
        event.preventDefault()
        event.stopPropagation()
        action(event)
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [id, getKeys, action, options.enabled])
}
