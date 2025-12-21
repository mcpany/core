"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import {
  CalendarIcon,
  EnvelopeClosedIcon,
  FaceIcon,
  GearIcon,
  PersonIcon,
  RocketIcon,
} from "@radix-ui/react-icons"
import {
    LayoutDashboard,
    Cpu,
    Database,
    Wrench,
    MessageSquare,
    Settings,
    Search
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
            <Cpu className="mr-2 h-4 w-4" />
            <span>Services</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/resources"))}>
            <Database className="mr-2 h-4 w-4" />
            <span>Resources</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
            <Wrench className="mr-2 h-4 w-4" />
            <span>Tools</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/prompts"))}>
            <MessageSquare className="mr-2 h-4 w-4" />
            <span>Prompts</span>
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
