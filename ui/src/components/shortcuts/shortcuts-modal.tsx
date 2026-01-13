/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import React from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useShortcuts } from "./shortcuts-provider"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Keyboard } from "lucide-react"

export function ShortcutsModal() {
  const { isModalOpen, closeModal, shortcuts } = useShortcuts()

  // Group by category
  const groupedShortcuts = React.useMemo(() => {
    const groups: Record<string, typeof shortcuts> = {}
    shortcuts.forEach((s) => {
      if (!groups[s.category]) {
        groups[s.category] = []
      }
      groups[s.category].push(s)
    })
    return groups
  }, [shortcuts])

  // Sort categories
  const sortedCategories = Object.keys(groupedShortcuts).sort()

  return (
    <Dialog open={isModalOpen} onOpenChange={(open) => !open && closeModal()}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
             <Keyboard className="h-5 w-5 text-muted-foreground" />
             Keyboard Shortcuts
          </DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[60vh] pr-4">
          <div className="space-y-6">
            {sortedCategories.map((category) => (
              <div key={category}>
                <h4 className="mb-2 text-sm font-medium text-muted-foreground uppercase tracking-wider">
                  {category}
                </h4>
                <div className="grid grid-cols-1 gap-2">
                  {groupedShortcuts[category].map((shortcut) => (
                    <div
                      key={shortcut.id}
                      className="flex items-center justify-between rounded-md border p-2 text-sm"
                    >
                      <span>{shortcut.description}</span>
                      <div className="flex gap-1">
                          {shortcut.key.split('+').map((k, i) => (
                             <kbd key={i} className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100 uppercase">
                                {k === "Cmd" ? "⌘" : k}
                             </kbd>
                          ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}

            {shortcuts.length === 0 && (
                <div className="text-center text-muted-foreground py-8">
                    No shortcuts registered.
                </div>
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
