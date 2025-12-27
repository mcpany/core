"use client"

import * as React from "react"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { StatusBadge } from "@/components/ui-custom/status-badge"
import { Switch } from "@/components/ui/switch"
import { Wrench } from "lucide-react"
import { apiClient, ToolDefinition } from "@/lib/client"
import { useToast } from "@/components/ui/use-toast"
import { Input } from "@/components/ui/input"

export default function ToolsPage() {
  const [tools, setTools] = React.useState<ToolDefinition[]>([])
  const [loading, setLoading] = React.useState(true)
  const { toast } = useToast()
  const [search, setSearch] = React.useState("")

  React.useEffect(() => {
    fetchTools()
  }, [])

  const fetchTools = async () => {
    try {
      const data = await apiClient.listTools()
      setTools(data)
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch tools",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleToggle = async (tool: ToolDefinition) => {
    try {
      setTools(tools.map(t => t.name === tool.name ? { ...t, enabled: !tool.enabled } : t))
      await apiClient.setToolStatus(tool.name, !tool.enabled)
      toast({ title: "Success", description: "Tool status updated" })
    } catch (error) {
       setTools(tools.map(t => t.name === tool.name ? { ...t, enabled: tool.enabled } : t))
       toast({
        title: "Error",
        description: "Failed to update tool",
        variant: "destructive",
      })
    }
  }

  const filteredTools = tools.filter(t => t.name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
          <p className="text-muted-foreground mt-2">
            Manage available tools from upstream services.
          </p>
        </div>
        <div className="w-64">
             <Input placeholder="Search tools..." value={search} onChange={e => setSearch(e.target.value)} />
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
           <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredTools.map((tool) => (
            <GlassCard key={tool.name} className="p-6 flex flex-col justify-between space-y-4" hoverEffect>
              <div className="flex items-start justify-between">
                 <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-lg bg-indigo-500/10 flex items-center justify-center text-indigo-500">
                        <Wrench className="w-5 h-5" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-base">{tool.name}</h3>
                        <p className="text-xs text-muted-foreground">{tool.serviceName}</p>
                    </div>
                 </div>
                 <StatusBadge status={tool.enabled || false} />
              </div>

              <p className="text-sm text-muted-foreground line-clamp-2 min-h-[40px]">
                  {tool.description}
              </p>

              <div className="flex items-center justify-between pt-2 border-t">
                  <span className="text-xs text-muted-foreground font-mono">Schema: JSON</span>
                  <Switch
                        checked={tool.enabled}
                        onCheckedChange={() => handleToggle(tool)}
                    />
              </div>
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  )
}
