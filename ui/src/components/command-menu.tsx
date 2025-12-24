"use client"

import * as React from "react"
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
    Logs,
    Terminal,
    Settings,
    Server,
    Wrench,
    FileText,
    Search,
    Github,
    Moon,
    Sun,
    Monitor
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
            <Server className="mr-2 h-4 w-4" />
            <span>Services</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
            <Wrench className="mr-2 h-4 w-4" />
            <span>Tools</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/playground"))}>
            <Terminal className="mr-2 h-4 w-4" />
            <span>Playground</span>
          </CommandItem>
           <CommandItem onSelect={() => runCommand(() => router.push("/logs"))}>
            <Logs className="mr-2 h-4 w-4" />
            <span>Logs</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Settings">
          <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
            <Settings className="mr-2 h-4 w-4" />
            <span>General Settings</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/settings/api-keys"))}>
            <PersonIcon className="mr-2 h-4 w-4" />
            <span>API Keys</span>
            <CommandShortcut>⌘P</CommandShortcut>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
         <CommandGroup heading="Theme">
          <CommandItem onSelect={() => runCommand(() => console.log("Set Light"))}>
            <Sun className="mr-2 h-4 w-4" />
            <span>Light</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => console.log("Set Dark"))}>
            <Moon className="mr-2 h-4 w-4" />
            <span>Dark</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => console.log("Set System"))}>
            <Monitor className="mr-2 h-4 w-4" />
            <span>System</span>
          </CommandItem>
        </CommandGroup>
         <CommandSeparator />
         <CommandGroup heading="Help">
             <CommandItem onSelect={() => runCommand(() => window.open("https://github.com/mcp-any/mcp-any", "_blank"))}>
                <Github className="mr-2 h-4 w-4" />
                <span>GitHub Repository</span>
             </CommandItem>
              <CommandItem onSelect={() => runCommand(() => window.open("https://mcp-any.com/docs", "_blank"))}>
                <FileText className="mr-2 h-4 w-4" />
                <span>Documentation</span>
             </CommandItem>
         </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
