/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import {
  CreditCard,
  Settings,
  User,
  LayoutDashboard,
  Server,
  Database,
  Terminal,
  PlusCircle,
  RefreshCw,
  Box,
  MessageSquare
} from "lucide-react"

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command"

export function CommandMenu() {
  const router = useRouter()
  const [open, setOpen] = React.useState(false)

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }

    document.addEventListener("keydown", down)
    return () => document.removeEventListener("keydown", down)
  }, [])

  const runCommand = React.useCallback((command: () => unknown) => {
    setOpen(false)
    command()
  }, [])

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Suggestions">
          <CommandItem onSelect={() => runCommand(() => router.push("/"))}>
            <LayoutDashboard className="mr-2 h-4 w-4" />
            <span>Dashboard</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/services"))}>
            <Server className="mr-2 h-4 w-4" />
            <span>Services</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
            <Terminal className="mr-2 h-4 w-4" />
            <span>Tools</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/resources"))}>
            <Box className="mr-2 h-4 w-4" />
            <span>Resources</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/prompts"))}>
            <MessageSquare className="mr-2 h-4 w-4" />
            <span>Prompts</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Tools">
          <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
            <CreditCard className="mr-2 h-4 w-4" />
            <span>stripe_charge</span>
            <CommandShortcut>Payment</CommandShortcut>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
            <User className="mr-2 h-4 w-4" />
            <span>get_user</span>
            <CommandShortcut>User</CommandShortcut>
          </CommandItem>
           <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
            <Database className="mr-2 h-4 w-4" />
            <span>search_docs</span>
            <CommandShortcut>Search</CommandShortcut>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Actions">
          <CommandItem onSelect={() => runCommand(() => router.push("/services"))}>
            <PlusCircle className="mr-2 h-4 w-4" />
            <span>Create Service</span>
          </CommandItem>
           <CommandItem onSelect={() => runCommand(() => window.location.reload())}>
            <RefreshCw className="mr-2 h-4 w-4" />
            <span>Reload Window</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Settings">
          <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
            <Settings className="mr-2 h-4 w-4" />
            <span>Settings</span>
            <CommandShortcut>⌘S</CommandShortcut>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
