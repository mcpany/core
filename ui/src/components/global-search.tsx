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
  Wrench,
  Search,
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
      apiClient.listServices().then((res) => {
        setServices(res.services)
      }).catch((err) => {
        console.error("Failed to fetch services for global search", err)
      })
    }
  }, [open])

  const runCommand = React.useCallback((command: () => unknown) => {
    setOpen(false)
    command()
  }, [])

  return (
    <>
      <div
        onClick={() => setOpen(true)}
        className="fixed bottom-4 right-4 z-50 flex h-12 w-12 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg transition-transform hover:scale-110 md:hidden"
      >
        <Search className="h-6 w-6" />
      </div>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Suggestions">
            <CommandItem
              onSelect={() => {
                runCommand(() => router.push("/"))
              }}
            >
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Dashboard</span>
            </CommandItem>
            <CommandItem
              onSelect={() => {
                runCommand(() => router.push("/settings"))
              }}
            >
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Services">
            {services.map((service) => (
              <CommandItem
                key={service.name}
                onSelect={() => {
                  runCommand(() => router.push(`/service/${encodeURIComponent(service.name)}`))
                }}
              >
                <Server className="mr-2 h-4 w-4" />
                <span>{service.name}</span>
                <CommandShortcut>SVC</CommandShortcut>
              </CommandItem>
            ))}
             {services.length === 0 && <CommandItem disabled>No services found</CommandItem>}
          </CommandGroup>
           <CommandSeparator />
          <CommandGroup heading="Tools">
             {/* Placeholder for tools */}
            <CommandItem
              disabled
            >
              <Wrench className="mr-2 h-4 w-4" />
              <span>Tools (Coming Soon)</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  )
}
