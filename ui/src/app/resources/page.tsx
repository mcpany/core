"use client"

import * as React from "react"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { StatusBadge } from "@/components/ui-custom/status-badge"
import { Switch } from "@/components/ui/switch"
import { FileText } from "lucide-react"
import { apiClient, ResourceDefinition } from "@/lib/client"
import { useToast } from "@/components/ui/use-toast"
import { Input } from "@/components/ui/input"

export default function ResourcesPage() {
  const [resources, setResources] = React.useState<ResourceDefinition[]>([])
  const [loading, setLoading] = React.useState(true)
  const { toast } = useToast()
  const [search, setSearch] = React.useState("")

  React.useEffect(() => {
    fetchResources()
  }, [])

  const fetchResources = async () => {
    try {
      const data = await apiClient.listResources()
      setResources(data)
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch resources",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleToggle = async (resource: ResourceDefinition) => {
    try {
      setResources(resources.map(r => r.uri === resource.uri ? { ...r, enabled: !resource.enabled } : r))
      await apiClient.setResourceStatus(resource.uri, !resource.enabled)
      toast({ title: "Success", description: "Resource status updated" })
    } catch (error) {
       setResources(resources.map(r => r.uri === resource.uri ? { ...r, enabled: resource.enabled } : r))
       toast({
        title: "Error",
        description: "Failed to update resource",
        variant: "destructive",
      })
    }
  }

  const filteredResources = resources.filter(r => r.name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
          <p className="text-muted-foreground mt-2">
            Manage exposed resources (files, database tables, etc).
          </p>
        </div>
        <div className="w-64">
             <Input placeholder="Search resources..." value={search} onChange={e => setSearch(e.target.value)} />
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
           <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {filteredResources.map((resource) => (
            <GlassCard key={resource.uri} className="p-6 flex flex-col justify-between space-y-4" hoverEffect>
              <div className="flex items-start justify-between">
                 <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-lg bg-orange-500/10 flex items-center justify-center text-orange-500">
                        <FileText className="w-5 h-5" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-base">{resource.name}</h3>
                         <p className="text-xs text-muted-foreground">{resource.serviceName}</p>
                    </div>
                 </div>
                 <StatusBadge status={resource.enabled || false} />
              </div>

               <div className="bg-slate-100 p-2 rounded text-xs font-mono text-slate-600 truncate">
                   {resource.uri}
               </div>

              <p className="text-sm text-muted-foreground">
                  {resource.description}
              </p>

              <div className="flex items-center justify-between pt-2 border-t">
                  <span className="text-xs text-muted-foreground">{resource.mimeType || "application/octet-stream"}</span>
                  <Switch
                        checked={resource.enabled}
                        onCheckedChange={() => handleToggle(resource)}
                    />
              </div>
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  )
}
