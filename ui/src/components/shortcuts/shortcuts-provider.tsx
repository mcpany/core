/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import React, { createContext, useContext, useEffect, useState, useCallback, useMemo } from "react"

export type Shortcut = {
  id: string
  key: string
  description: string
  category: string
  // If action is provided, the provider will try to handle it.
  // However, often components handle it themselves (like `keydown` listeners).
  // In that case, this is just for display.
  action?: () => void
}

type ShortcutsContextType = {
  register: (shortcut: Omit<Shortcut, "id">) => string
  unregister: (id: string) => void
  openModal: () => void
  closeModal: () => void
  isModalOpen: boolean
  shortcuts: Shortcut[]
}

const ShortcutsContext = createContext<ShortcutsContextType | null>(null)

export function useShortcuts() {
  const context = useContext(ShortcutsContext)
  if (!context) {
    throw new Error("useShortcuts must be used within a ShortcutsProvider")
  }
  return context
}

export function ShortcutsProvider({ children }: { children: React.ReactNode }) {
  const [shortcuts, setShortcuts] = useState<Shortcut[]>([])
  const [isModalOpen, setIsModalOpen] = useState(false)

  const register = useCallback((shortcut: Omit<Shortcut, "id">) => {
    const id = Math.random().toString(36).substr(2, 9)
    setShortcuts((prev) => {
        // Avoid duplicates if possible, or just append
        return [...prev, { ...shortcut, id }]
    })
    return id
  }, [])

  const unregister = useCallback((id: string) => {
    setShortcuts((prev) => prev.filter((s) => s.id !== id))
  }, [])

  const openModal = useCallback(() => setIsModalOpen(true), [])
  const closeModal = useCallback(() => setIsModalOpen(false), [])

  // Register the global help shortcut (?)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "?" && !e.metaKey && !e.ctrlKey && !e.altKey && (document.activeElement?.tagName !== "INPUT" && document.activeElement?.tagName !== "TEXTAREA")) {
          // Only trigger if not typing in an input
          e.preventDefault()
          setIsModalOpen((prev) => !prev)
      }
      // Also Cmd+/ or Ctrl+/
      if (e.key === "/" && (e.metaKey || e.ctrlKey)) {
          e.preventDefault()
          setIsModalOpen((prev) => !prev)
      }
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [])

  const value = useMemo(
    () => ({
      register,
      unregister,
      openModal,
      closeModal,
      isModalOpen,
      shortcuts,
    }),
    [register, unregister, openModal, closeModal, isModalOpen, shortcuts]
  )

  return (
    <ShortcutsContext.Provider value={value}>
      {children}
    </ShortcutsContext.Provider>
  )
}

export function useShortcut(
    key: string,
    description: string,
    category: string,
    action?: (e: KeyboardEvent) => void,
    deps: React.DependencyList = []
) {
    const { register, unregister } = useShortcuts();

    useEffect(() => {
        const id = register({ key, description, category });

        const handler = (e: KeyboardEvent) => {
            if (!action) return;

            // Simple parser for "Cmd+K", "Ctrl+B", etc.
            const parts = key.toLowerCase().split("+");

            // Detect "Universal" Cmd/Ctrl
            // If "Cmd" is used, it typically implies "Meta" (Command) on Mac and "Control" on others.
            // If "Ctrl" is strictly used, it usually means Control.

            const isMac = typeof navigator !== 'undefined' && navigator.platform.toLowerCase().includes('mac');

            const usesCmd = parts.includes("cmd");
            const usesMeta = parts.includes("meta");
            const usesCtrl = parts.includes("ctrl");
            const usesShift = parts.includes("shift");
            const usesAlt = parts.includes("alt");

            const keyPart = parts.find(p => !["cmd", "ctrl", "meta", "shift", "alt"].includes(p));

            if (!keyPart) return;

            const matchesKey = e.key.toLowerCase() === keyPart.toLowerCase();

            // Check Modifiers
            const pressedMeta = e.metaKey;
            const pressedCtrl = e.ctrlKey;
            const pressedShift = e.shiftKey;
            const pressedAlt = e.altKey;

            let match = true;

            if (usesCmd) {
                // "Cmd" -> Meta on Mac, Ctrl on Windows/Linux
                if (isMac) {
                    if (!pressedMeta) match = false;
                } else {
                    if (!pressedCtrl) match = false;
                }
            }

            if (usesMeta && !pressedMeta) match = false;
            if (usesCtrl && !pressedCtrl) match = false; // "Ctrl" strictly requires Ctrl key
            if (usesShift && !pressedShift) match = false;
            if (usesAlt && !pressedAlt) match = false;

            // Also ensure no EXTRA modifiers are pressed?
            // Usually yes, but sometimes no. Let's start with loose matching for provided modifiers.

            if (match && matchesKey) {
                // If action is provided, call it.
                 e.preventDefault();
                 action(e);
            }
        };

        if (action) {
            window.addEventListener("keydown", handler);
        }

        return () => {
            unregister(id);
            if (action) {
                window.removeEventListener("keydown", handler);
            }
        };
    }, [key, description, category, register, unregister, ...deps]);
}
