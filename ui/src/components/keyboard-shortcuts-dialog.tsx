/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Keyboard, RotateCcw, X, Edit2 } from "lucide-react"
import { useKeyboardShortcuts, ShortcutDefinition } from "@/contexts/keyboard-shortcuts-context"
import { cn } from "@/lib/utils"

/**
 * Props for the KeyboardShortcutsDialog component.
 */
interface KeyboardShortcutsDialogProps {
  /** Whether the dialog is open. */
  open: boolean
  /** Callback to handle dialog open/close state changes. */
  onOpenChange: (open: boolean) => void
}

/**
 * A dialog that displays and allows customization of keyboard shortcuts.
 */
export function KeyboardShortcutsDialog({ open, onOpenChange }: KeyboardShortcutsDialogProps) {
  const { shortcuts, overrides, updateOverride, resetOverride } = useKeyboardShortcuts()
  const [editingId, setEditingId] = React.useState<string | null>(null)
  const [recordedKeys, setRecordedKeys] = React.useState<string[]>([])

  // Group shortcuts by category
  const groupedShortcuts = React.useMemo(() => {
    const groups: Record<string, ShortcutDefinition[]> = {}
    Object.values(shortcuts).forEach(s => {
      const cat = s.category || "General"
      if (!groups[cat]) groups[cat] = []
      groups[cat].push(s)
    })
    return groups
  }, [shortcuts])

  const formatKeyForDisplay = (keyDef: string) => {
    const parts = keyDef.toLowerCase().split("+")
    return parts.map(p => {
      if (p === "meta" || p === "cmd" || p === "command") return "⌘"
      if (p === "ctrl" || p === "control") return "⌃"
      if (p === "alt" || p === "option") return "⌥"
      if (p === "shift") return "⇧"
      if (p === "enter") return "↵"
      if (p === "backspace") return "⌫"
      if (p === "delete") return "⌦"
      if (p === "escape") return "Esc"
      return p.toUpperCase()
    }).join("")
  }

  const handleStartEditing = (id: string) => {
    setEditingId(id)
    setRecordedKeys([])
  }

  const handleStopEditing = () => {
    setEditingId(null)
    setRecordedKeys([])
  }

  const handleReset = (id: string) => {
    resetOverride(id)
  }

  const handleSave = (id: string) => {
     if (recordedKeys.length > 0) {
         updateOverride(id, recordedKeys)
     }
     handleStopEditing()
  }

  const handleKeyDown = React.useCallback((e: React.KeyboardEvent) => {
    if (!editingId) return

    e.preventDefault()
    e.stopPropagation()

    // Don't record just modifiers
    if (["Meta", "Control", "Alt", "Shift"].includes(e.key)) return

    const modifiers = []
    if (e.metaKey) modifiers.push("meta")
    if (e.ctrlKey) modifiers.push("ctrl")
    if (e.altKey) modifiers.push("alt")
    if (e.shiftKey) modifiers.push("shift")

    // Normalize key
    let key = e.key.toLowerCase()
    if (key === " ") key = "space"

    const combo = [...modifiers, key].join("+")
    setRecordedKeys([combo])
  }, [editingId])

  return (
    <Dialog open={open} onOpenChange={(val) => {
        if (!val) handleStopEditing()
        onOpenChange(val)
    }}>
      <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col p-0 gap-0 overflow-hidden">
        <DialogHeader className="p-6 border-b">
          <DialogTitle className="flex items-center gap-2">
            <Keyboard className="h-5 w-5" />
            Keyboard Shortcuts
          </DialogTitle>
          <DialogDescription>
            View and customize keyboard shortcuts. Click on a shortcut to edit it.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto p-6">
            {Object.keys(groupedShortcuts).length === 0 && (
                <div className="text-center text-muted-foreground py-8">
                    No shortcuts registered yet.
                </div>
            )}

            {Object.entries(groupedShortcuts).map(([category, items]) => (
                <div key={category} className="mb-6 last:mb-0">
                    <h3 className="text-sm font-medium text-muted-foreground mb-3 uppercase tracking-wider">{category}</h3>
                    <div className="grid gap-2">
                        {items.map(item => {
                            const isEditing = editingId === item.id
                            const activeKeys = overrides[item.id] || item.defaultKeys
                            const hasOverride = !!overrides[item.id]

                            return (
                                <div key={item.id} className={cn(
                                    "flex items-center justify-between p-3 rounded-lg border bg-card text-card-foreground shadow-sm transition-colors",
                                    isEditing && "border-primary ring-1 ring-primary"
                                )}>
                                    <div className="flex flex-col">
                                        <span className="font-medium text-sm">{item.label}</span>
                                        {hasOverride && (
                                            <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                Default: {item.defaultKeys.map(formatKeyForDisplay).join(", ")}
                                            </span>
                                        )}
                                    </div>

                                    <div className="flex items-center gap-2">
                                        {isEditing ? (
                                            <div className="flex items-center gap-2">
                                                <div
                                                    className="h-8 min-w-[100px] px-3 flex items-center justify-center bg-muted rounded border border-input text-sm font-mono focus:outline-none animate-pulse"
                                                    onKeyDown={handleKeyDown}
                                                    tabIndex={0}
                                                    autoFocus
                                                >
                                                    {recordedKeys.length > 0 ? recordedKeys.map(formatKeyForDisplay).join(", ") : "Press keys..."}
                                                </div>
                                                <Button size="sm" onClick={() => handleSave(item.id)} disabled={recordedKeys.length === 0}>Save</Button>
                                                <Button size="sm" variant="ghost" onClick={handleStopEditing}><X className="h-4 w-4" /></Button>
                                            </div>
                                        ) : (
                                            <div className="flex items-center gap-2">
                                                <div className="flex gap-1">
                                                    {activeKeys.map((k, i) => (
                                                        <kbd key={i} className="pointer-events-none inline-flex h-6 select-none items-center gap-1 rounded border bg-muted px-2 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
                                                            {formatKeyForDisplay(k)}
                                                        </kbd>
                                                    ))}
                                                </div>
                                                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleStartEditing(item.id)}>
                                                    <Edit2 className="h-3 w-3" />
                                                </Button>
                                                {hasOverride && (
                                                    <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-foreground" onClick={() => handleReset(item.id)} title="Reset to default">
                                                        <RotateCcw className="h-3 w-3" />
                                                    </Button>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            )
                        })}
                    </div>
                </div>
            ))}
        </div>
      </DialogContent>
    </Dialog>
  )
}
