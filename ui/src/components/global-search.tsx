"use client"

import * as React from "react"
import {
  Settings,
  LayoutDashboard,
  Terminal,
  Play,
  Briefcase,
  Moon,
  Sun,
  Laptop,
  Wrench,
  FileText,
  MessageSquare,
  Workflow,
  Webhook
} from "lucide-react"
import { useRouter } from "next/navigation"
import { useTheme } from "next-themes"

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

export function GlobalSearch() {
  const [open, setOpen] = React.useState(false)
  const router = useRouter()
  const { setTheme } = useTheme()

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
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Navigation">
            <CommandItem
              onSelect={() => runCommand(() => router.push("/"))}
            >
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Dashboard</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runCommand(() => router.push("/services"))}
            >
              <Briefcase className="mr-2 h-4 w-4" />
              <span>Services</span>
            </CommandItem>
             <CommandItem
              onSelect={() => runCommand(() => router.push("/tools"))}
            >
              <Wrench className="mr-2 h-4 w-4" />
              <span>Tools</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runCommand(() => router.push("/resources"))}
            >
              <FileText className="mr-2 h-4 w-4" />
              <span>Resources</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runCommand(() => router.push("/prompts"))}
            >
              <MessageSquare className="mr-2 h-4 w-4" />
              <span>Prompts</span>
            </CommandItem>
             <CommandItem
              onSelect={() => runCommand(() => router.push("/middleware"))}
            >
              <Workflow className="mr-2 h-4 w-4" />
              <span>Middleware</span>
            </CommandItem>
             <CommandItem
              onSelect={() => runCommand(() => router.push("/webhooks"))}
            >
              <Webhook className="mr-2 h-4 w-4" />
              <span>Webhooks</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runCommand(() => router.push("/logs"))}
            >
              <Terminal className="mr-2 h-4 w-4" />
              <span>Logs</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runCommand(() => router.push("/playground"))}
            >
              <Play className="mr-2 h-4 w-4" />
              <span>Playground</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runCommand(() => router.push("/settings"))}
            >
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Theme">
            <CommandItem onSelect={() => runCommand(() => setTheme("light"))}>
              <Sun className="mr-2 h-4 w-4" />
              <span>Light</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme("dark"))}>
              <Moon className="mr-2 h-4 w-4" />
              <span>Dark</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme("system"))}>
              <Laptop className="mr-2 h-4 w-4" />
              <span>System</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  )
}
