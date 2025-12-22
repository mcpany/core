"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import {
  Calculator,
  Calendar,
  CreditCard,
  Settings,
  Smile,
  User,
  LayoutDashboard,
  Server,
  Box,
  MessageSquare,
  Wrench,
  Moon,
  Sun,
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
import { apiClient } from "@/lib/client"
import { UpstreamServiceConfig } from "@/lib/types"

export function GlobalSearch() {
  const [open, setOpen] = React.useState(false)
  const [services, setServices] = React.useState<UpstreamServiceConfig[]>([])
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

  React.useEffect(() => {
    if (open) {
      apiClient.listServices()
        .then((res) => setServices(res.services))
        .catch((err) => console.error("Failed to load services for search", err))
    }
  }, [open])

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
          <CommandGroup heading="Suggestions">
            <CommandItem onSelect={() => runCommand(() => router.push("/"))}>
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Dashboard</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/services"))}>
              <Server className="mr-2 h-4 w-4" />
              <span>Services</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/resources"))}>
              <Box className="mr-2 h-4 w-4" />
              <span>Resources</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => router.push("/prompts"))}>
              <MessageSquare className="mr-2 h-4 w-4" />
              <span>Prompts</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
              <Wrench className="mr-2 h-4 w-4" />
              <span>Tools</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
              <CommandShortcut>⌘S</CommandShortcut>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          {services.length > 0 && (
             <CommandGroup heading="Services">
              {services.map((service) => (
                <CommandItem
                  key={service.name}
                  onSelect={() => runCommand(() => router.push(`/services/${service.name}`))}
                >
                  <Server className="mr-2 h-4 w-4" />
                  <span>{service.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          )}
           <CommandSeparator />
          <CommandGroup heading="Theme">
             <CommandItem onSelect={() => runCommand(() => console.log("Set theme to light"))}>
              <Sun className="mr-2 h-4 w-4" />
              <span>Light</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => console.log("Set theme to dark"))}>
              <Moon className="mr-2 h-4 w-4" />
              <span>Dark</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => console.log("Set theme to system"))}>
              <Laptop className="mr-2 h-4 w-4" />
              <span>System</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  )
}
