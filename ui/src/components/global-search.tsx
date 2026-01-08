/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { useTheme } from "next-themes"
import {
  Settings,
  Calculator,
  LayoutDashboard,
  Server,
  FileText,
  Terminal,
  Moon,
  Sun,
  Laptop,
  Database,
  Wrench,
  FileBox,
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
} from "@/components/ui/command"

import { apiClient, ToolDefinition, ResourceDefinition, PromptDefinition, UpstreamServiceConfig } from "@/lib/client"

export function GlobalSearch() {
  const [open, setOpen] = React.useState(false)
  const [query, setQuery] = React.useState("")
  const router = useRouter()
  const { setTheme } = useTheme()

  const [services, setServices] = React.useState<UpstreamServiceConfig[]>([])
  const [tools, setTools] = React.useState<ToolDefinition[]>([])
  const [resources, setResources] = React.useState<ResourceDefinition[]>([])
  const [prompts, setPrompts] = React.useState<PromptDefinition[]>([])
  const [loading, setLoading] = React.useState(false)
  const lastFetched = React.useRef(0)

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.key === "k" || e.key === "K") && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }

    document.addEventListener("keydown", down)
    return () => document.removeEventListener("keydown", down)
  }, [])

  React.useEffect(() => {
    if (open) {
      // ⚡ Bolt Optimization: Prevent redundant API calls if data was fetched recently (< 1 min)
      // This reduces 4 concurrent requests every time the search dialog is opened.
      const now = Date.now()
      if (now - lastFetched.current < 60000 && lastFetched.current > 0) {
        return
      }

      setLoading(true)
      Promise.all([
        apiClient.listServices().catch(() => ({ services: [] })),
        apiClient.listTools().catch(() => ({ tools: [] })),
        apiClient.listResources().catch(() => ({ resources: [] })),
        apiClient.listPrompts().catch(() => ({ prompts: [] }))
      ]).then(([servicesData, toolsData, resourcesData, promptsData]) => {
         setServices(Array.isArray(servicesData) ? servicesData : servicesData.services || [])
         setTools(toolsData.tools || [])
         setResources(resourcesData.resources || [])
         setPrompts(promptsData.prompts || [])
         lastFetched.current = Date.now()
      }).finally(() => {
        setLoading(false)
      })
    }
  }, [open])

  const runCommand = React.useCallback((command: () => unknown) => {
    setOpen(false)
    command()
  }, [])

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="inline-flex items-center gap-2 whitespace-nowrap transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 border border-input hover:bg-accent hover:text-accent-foreground px-4 py-2 relative h-8 w-full justify-start rounded-[0.5rem] bg-muted/50 text-sm font-normal text-muted-foreground shadow-none sm:pr-12 md:w-40 lg:w-64"
      >
        <span className="hidden lg:inline-flex">Search feature...</span>
        <span className="inline-flex lg:hidden">Search...</span>
        <kbd className="pointer-events-none absolute right-[0.3rem] top-[0.3rem] hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
          <span className="text-xs">⌘</span>K
        </kbd>
      </button>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." value={query} onValueChange={setQuery} />
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
             <CommandItem onSelect={() => runCommand(() => router.push("/resources"))}>
              <FileBox className="mr-2 h-4 w-4" />
              <span>Resources</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => router.push("/prompts"))}>
              <MessageSquare className="mr-2 h-4 w-4" />
              <span>Prompts</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/logs"))}>
              <FileText className="mr-2 h-4 w-4" />
              <span>Logs</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/playground"))}>
              <Terminal className="mr-2 h-4 w-4" />
              <span>Playground</span>
            </CommandItem>
             <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />

          {services.length > 0 && (
             <CommandGroup heading="Services">
               {services.map((service) => (
                 <CommandItem key={service.id || service.name} value={`service ${service.name}`} onSelect={() => runCommand(() => router.push(`/services?id=${service.id}`))}>
                   <Database className="mr-2 h-4 w-4" />
                   <span>{service.name}</span>
                   {service.version && <span className="ml-2 text-xs text-muted-foreground">v{service.version}</span>}
                 </CommandItem>
               ))}
             </CommandGroup>
          )}

          {tools.length > 0 && (
             <CommandGroup heading="Tools">
               {tools.map((tool) => (
                 <CommandItem key={tool.name} value={`tool ${tool.name}`} onSelect={() => runCommand(() => router.push(`/tools?name=${tool.name}`))}>
                   <Calculator className="mr-2 h-4 w-4" />
                   <span>{tool.name}</span>
                   <span className="ml-2 text-xs text-muted-foreground truncate max-w-[200px]">{tool.description}</span>
                 </CommandItem>
               ))}
             </CommandGroup>
          )}

           {resources.length > 0 && (
             <CommandGroup heading="Resources">
               {resources.map((resource) => (
                 <CommandItem key={resource.uri} value={`resource ${resource.name}`} onSelect={() => runCommand(() => router.push(`/resources?uri=${encodeURIComponent(resource.uri)}`))}>
                   <FileBox className="mr-2 h-4 w-4" />
                   <span>{resource.name}</span>
                 </CommandItem>
               ))}
             </CommandGroup>
          )}

           {prompts.length > 0 && (
             <CommandGroup heading="Prompts">
               {prompts.map((prompt) => (
                 <CommandItem key={prompt.name} value={`prompt ${prompt.name}`} onSelect={() => runCommand(() => router.push(`/prompts?name=${prompt.name}`))}>
                   <MessageSquare className="mr-2 h-4 w-4" />
                   <span>{prompt.name}</span>
                 </CommandItem>
               ))}
             </CommandGroup>
          )}

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
