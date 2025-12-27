"use client"

import * as React from "react"

export function SearchTrigger() {
  return (
    <div
      className="hidden md:flex items-center text-sm text-muted-foreground bg-muted/50 px-2 py-1 rounded-md border mr-2 cursor-pointer hover:bg-muted"
      onClick={() => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }))}
    >
      <span className="text-xs mr-2">Search</span>
      <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
        <span className="text-xs">⌘</span>K
      </kbd>
    </div>
  )
}
