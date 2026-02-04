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
  MessageSquare,
  RefreshCw,
  Copy,
  RotateCcw,
  Keyboard
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
import { useToast } from "@/hooks/use-toast"
import { useShortcut, useKeyboardShortcuts } from "@/contexts/keyboard-shortcuts-context"
import { KeyboardShortcutsDialog } from "@/components/keyboard-shortcuts-dialog"

/**
 * Global search component that provides quick access to navigation, tools, services, and actions.
 * It is triggered by a keyboard shortcut (Cmd/Ctrl+K) or by clicking the search button.
 *
 * @returns The rendered GlobalSearch component (button + dialog).
 */
export function GlobalSearch() {
  const [open, setOpen] = React.useState(false)
  const [showShortcuts, setShowShortcuts] = React.useState(false)
  const [query, setQuery] = React.useState("")
  const router = useRouter()
  const { setTheme } = useTheme()
  const { toast } = useToast()
  const { getKeys } = useKeyboardShortcuts()

  const [services, setServices] = React.useState<UpstreamServiceConfig[]>([])
  const [tools, setTools] = React.useState<ToolDefinition[]>([])
  const [resources, setResources] = React.useState<ResourceDefinition[]>([])
  const [prompts, setPrompts] = React.useState<PromptDefinition[]>([])
  const [recentTools, setRecentTools] = React.useState<ToolDefinition[]>([])
  const [loading, setLoading] = React.useState(false)
  const lastFetched = React.useRef(0)

  useShortcut("search.toggle", ["meta+k", "ctrl+k"], () => setOpen((open) => !open), {
    label: "Toggle Search",
    category: "Navigation"
  })

  const fetchData = React.useCallback(async () => {
      setLoading(true)
      try {
        const [servicesData, toolsData, resourcesData, promptsData] = await Promise.all([
            apiClient.listServices().catch(() => ({ services: [] })),
            apiClient.listTools().catch(() => ({ tools: [] })),
            apiClient.listResources().catch(() => ({ resources: [] })),
            apiClient.listPrompts().catch(() => ({ prompts: [] }))
        ])

        setServices(servicesData && (Array.isArray(servicesData) ? servicesData : servicesData.services) || [])
        setTools(toolsData?.tools || [])
        setResources(resourcesData?.resources || [])
        setPrompts(promptsData?.prompts || [])
        lastFetched.current = Date.now()
      } finally {
        setLoading(false)
      }
  }, [])

  React.useEffect(() => {
    if (open) {
      // Load recent tools from localStorage
      const saved = JSON.parse(localStorage.getItem("recent_tools") || "[]") as ToolDefinition[]
      setRecentTools(saved)

      // ⚡ Bolt Optimization: Prevent redundant API calls if data was fetched recently (< 1 min)
      // This reduces 4 concurrent requests every time the search dialog is opened.
      const now = Date.now()
      if (now - lastFetched.current < 60000 && lastFetched.current > 0) {
        return
      }
      fetchData()
    }
  }, [open, fetchData])

  const runCommand = React.useCallback((command: () => unknown) => {
    setOpen(false)
    command()
  }, [])

  const selectTool = React.useCallback((tool: ToolDefinition) => {
    runCommand(() => {
        // Save to recent tools
        const updated = [
            tool,
            ...recentTools.filter(t => t.name !== tool.name)
        ].slice(0, 5)
        localStorage.setItem("recent_tools", JSON.stringify(updated))
        setRecentTools(updated)

        router.push(`/tools?name=${tool.name}`)
    })
  }, [runCommand, router, recentTools])

  const copyToClipboard = React.useCallback((text: string, label: string) => {
      runCommand(() => {
          navigator.clipboard.writeText(text)
          toast({
              title: "Copied to clipboard",
              description: `Copied ${label} to clipboard.`
          })
      })
  }, [runCommand, toast])

  const formatKey = (keys: string[]) => {
      if (!keys || keys.length === 0) return ""
      const key = keys[0] // Just show first one
      const parts = key.toLowerCase().split("+")
      return parts.map(p => {
          if (p === "meta" || p === "cmd") return "⌘"
          if (p === "ctrl") return "⌃"
          if (p === "shift") return "⇧"
          if (p === "alt") return "⌥"
          return p.toUpperCase()
      }).join("")
  }

  const restartService = React.useCallback(async (serviceName: string) => {
      runCommand(async () => {
          toast({
              title: "Restarting Service",
              description: `Restarting ${serviceName}...`
          })
          try {
              await apiClient.setServiceStatus(serviceName, true) // Disable
              // Small delay to ensure it stops? Or just immediately enable
              await new Promise(r => setTimeout(r, 1000))
              await apiClient.setServiceStatus(serviceName, false) // Enable
               toast({
                  title: "Service Restarted",
                  description: `${serviceName} has been restarted.`
              })
          } catch (e) {
              toast({
                  variant: "destructive",
                  title: "Restart Failed",
                  description: `Failed to restart ${serviceName}.`
              })
          }
      })
  }, [runCommand, toast])

    const reloadWindow = React.useCallback(() => {
        runCommand(() => {
            window.location.reload()
        })
    }, [runCommand])


  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="inline-flex items-center gap-2 whitespace-nowrap transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 border border-input hover:bg-accent hover:text-accent-foreground px-4 py-2 relative h-8 w-full justify-start rounded-[0.5rem] bg-muted/50 text-sm font-normal text-muted-foreground shadow-none sm:pr-12 md:w-40 lg:w-64"
      >
        <span className="hidden lg:inline-flex">Search or type &gt; for actions...</span>
        <span className="inline-flex lg:hidden">Search...</span>
        <kbd className="pointer-events-none absolute right-[0.3rem] top-[0.3rem] hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
          {formatKey(getKeys("search.toggle"))}
        </kbd>
      </button>
      <KeyboardShortcutsDialog open={showShortcuts} onOpenChange={setShowShortcuts} />
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." value={query} onValueChange={setQuery} />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Suggestions">
            <CommandItem onSelect={() => runCommand(() => router.push("/"))}>
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Dashboard</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => router.push("/upstream-services"))}>
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

           <CommandGroup heading="System Actions">
              <CommandItem value="keyboard shortcuts" onSelect={() => runCommand(() => setShowShortcuts(true))}>
                  <Keyboard className="mr-2 h-4 w-4" />
                  <span>Keyboard Shortcuts</span>
              </CommandItem>
              <CommandItem value="reload window" onSelect={reloadWindow}>
                  <RotateCcw className="mr-2 h-4 w-4" />
                  <span>Reload Window</span>
              </CommandItem>
               <CommandItem value="refresh data" onSelect={() => {
                   lastFetched.current = 0; // Force invalidate
                   fetchData();
                   toast({ title: "Refreshing Data..." });
               }}>
                  <RefreshCw className="mr-2 h-4 w-4" />
                  <span>Refresh Data</span>
              </CommandItem>
              <CommandItem value="copy current url" onSelect={() => copyToClipboard(window.location.href, "Current URL")}>
                  <Copy className="mr-2 h-4 w-4" />
                  <span>Copy Current URL</span>
              </CommandItem>
           </CommandGroup>
           <CommandSeparator />

           {recentTools.length > 0 && query.length === 0 && (
             <CommandGroup heading="Recent Tools">
               {recentTools.map((tool) => (
                 <CommandItem key={`recent-${tool.name}`} value={`recent tool ${tool.name}`} onSelect={() => selectTool(tool)}>
                   <RotateCcw className="mr-2 h-4 w-4 text-muted-foreground" />
                   <span>{tool.name}</span>
                 </CommandItem>
               ))}
             </CommandGroup>
           )}

          {services.length > 0 && (
             <CommandGroup heading="Services">
               {services.map((service) => (
                 <CommandItem key={service.id || service.name} value={`service ${service.name}`} onSelect={() => runCommand(() => router.push(`/upstream-services?id=${service.id}`))}>
                   <Database className="mr-2 h-4 w-4" />
                   <span>{service.name}</span>
                   {service.version && <span className="ml-2 text-xs text-muted-foreground">v{service.version}</span>}
                 </CommandItem>
               ))}
               {/* Actions for Services - only shown if searching for "restart" or service name */}
               {/* ⚡ Bolt Optimization: Only render these heavy actions when the user has typed something.
                   This significantly reduces the number of DOM nodes when the dialog is first opened. */}
               {query.length > 0 && services.map((service) => (
                   <CommandItem key={`restart-${service.name}`} value={`restart service ${service.name}`} onSelect={() => restartService(service.name)}>
                       <RefreshCw className="mr-2 h-4 w-4 text-orange-500" />
                       <span>Restart {service.name}</span>
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
                {/* ⚡ Bolt Optimization: Only render copy actions when searching to reduce DOM nodes */}
                {query.length > 0 && resources.map((resource) => (
                 <CommandItem key={`copy-${resource.uri}`} value={`copy uri ${resource.name}`} onSelect={() => copyToClipboard(resource.uri, "Resource URI")}>
                   <Copy className="mr-2 h-4 w-4 text-blue-500" />
                   <span>Copy URI: {resource.name}</span>
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
