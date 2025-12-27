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
  Terminal,
  Network,
  Logs,
  FileText,
  Search,
  Moon,
  Sun,
  Laptop
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

export function CommandPalette() {
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
       <div className="fixed bottom-4 right-4 z-50 md:hidden">
        <button
          onClick={() => setOpen(true)}
          className="bg-primary text-primary-foreground p-3 rounded-full shadow-lg"
        >
          <Search className="h-6 w-6" />
        </button>
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
            <CommandItem onSelect={() => runCommand(() => router.push("/playground"))}>
              <Terminal className="mr-2 h-4 w-4" />
              <span>Playground</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/topology"))}>
              <Network className="mr-2 h-4 w-4" />
              <span>Topology</span>
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
              <span>Settings</span>
              <CommandShortcut>⌘S</CommandShortcut>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme("light"))}>
              <Sun className="mr-2 h-4 w-4" />
              <span>Light Mode</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme("dark"))}>
              <Moon className="mr-2 h-4 w-4" />
              <span>Dark Mode</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => setTheme("system"))}>
              <Laptop className="mr-2 h-4 w-4" />
              <span>System Mode</span>
            </CommandItem>
          </CommandGroup>
           <CommandGroup heading="Tools (Mock)">
            <CommandItem onSelect={() => runCommand(() => console.log("List Files"))}>
              <FileText className="mr-2 h-4 w-4" />
              <span>filesystem / list_files</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => console.log("Read File"))}>
              <FileText className="mr-2 h-4 w-4" />
              <span>filesystem / read_file</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  )
}
