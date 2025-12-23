"use client"

import * as React from "react"
import {
  Calculator,
  Calendar,
  CreditCard,
  Settings,
  Smile,
  User,
  LayoutDashboard,
  Server,
  Wrench,
  Search,
  ExternalLink,
  Laptop
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
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"

export function GlobalSearch() {
  const [open, setOpen] = React.useState(false)
  const router = useRouter()

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
    <>
        <Button
            variant="outline"
            className="fixed bottom-4 right-4 z-50 flex h-10 w-10 items-center justify-center rounded-full p-0 shadow-lg md:h-12 md:w-12"
            onClick={() => setOpen(true)}
        >
            <Search className="h-5 w-5" />
            <span className="sr-only">Open Search</span>
        </Button>
        <div className="hidden md:flex fixed top-4 right-4 z-50">
           <Button
            variant="outline"
            className="relative h-9 w-full justify-start rounded-[0.5rem] text-sm text-muted-foreground sm:pr-12 md:w-40 lg:w-64"
            onClick={() => setOpen(true)}
          >
            <span className="hidden lg:inline-flex">Search...</span>
            <span className="inline-flex lg:hidden">Search...</span>
            <kbd className="pointer-events-none absolute right-1.5 top-1.5 hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
              <span className="text-xs">⌘</span>K
            </kbd>
          </Button>
        </div>

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
              <Wrench className="mr-2 h-4 w-4" />
              <span>Tools</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Tools">
            <CommandItem onSelect={() => runCommand(() => router.push("/tools?q=calculator"))}>
              <Calculator className="mr-2 h-4 w-4" />
              <span>Calculator</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/tools?q=filesystem"))}>
              <Laptop className="mr-2 h-4 w-4" />
              <span>FileSystem</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => router.push("/tools?q=websearch"))}>
              <Search className="mr-2 h-4 w-4" />
              <span>Web Search</span>
            </CommandItem>
          </CommandGroup>
           <CommandSeparator />
          <CommandGroup heading="System">
            <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
              <CommandShortcut>⌘S</CommandShortcut>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => window.open("https://github.com/mcp-any/mcp-any", "_blank"))}>
              <ExternalLink className="mr-2 h-4 w-4" />
              <span>Documentation</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  )
}
