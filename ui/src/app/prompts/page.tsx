"use client"

import * as React from "react"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { StatusBadge } from "@/components/ui-custom/status-badge"
import { Switch } from "@/components/ui/switch"
import { MessageSquare } from "lucide-react"
import { apiClient, PromptDefinition } from "@/lib/client"
import { useToast } from "@/components/ui/use-toast"
import { Input } from "@/components/ui/input"

export default function PromptsPage() {
  const [prompts, setPrompts] = React.useState<PromptDefinition[]>([])
  const [loading, setLoading] = React.useState(true)
  const { toast } = useToast()
  const [search, setSearch] = React.useState("")

  React.useEffect(() => {
    fetchPrompts()
  }, [])

  const fetchPrompts = async () => {
    try {
      const data = await apiClient.listPrompts()
      setPrompts(data)
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch prompts",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleToggle = async (prompt: PromptDefinition) => {
    try {
      setPrompts(prompts.map(p => p.name === prompt.name ? { ...p, enabled: !prompt.enabled } : p))
      await apiClient.setPromptStatus(prompt.name, !prompt.enabled)
      toast({ title: "Success", description: "Prompt status updated" })
    } catch (error) {
       setPrompts(prompts.map(p => p.name === prompt.name ? { ...p, enabled: prompt.enabled } : p))
       toast({
        title: "Error",
        description: "Failed to update prompt",
        variant: "destructive",
      })
    }
  }

  const filteredPrompts = prompts.filter(p => p.name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
          <p className="text-muted-foreground mt-2">
            Manage reusable prompts and templates.
          </p>
        </div>
        <div className="w-64">
             <Input placeholder="Search prompts..." value={search} onChange={e => setSearch(e.target.value)} />
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
           <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredPrompts.map((prompt) => (
            <GlassCard key={prompt.name} className="p-6 flex flex-col justify-between space-y-4" hoverEffect>
              <div className="flex items-start justify-between">
                 <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-lg bg-pink-500/10 flex items-center justify-center text-pink-500">
                        <MessageSquare className="w-5 h-5" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-base">{prompt.name}</h3>
                         <p className="text-xs text-muted-foreground">{prompt.serviceName}</p>
                    </div>
                 </div>
                 <StatusBadge status={prompt.enabled || false} />
              </div>

              <p className="text-sm text-muted-foreground line-clamp-2">
                  {prompt.description}
              </p>

              <div className="flex items-center justify-between pt-2 border-t">
                  <span className="text-xs text-muted-foreground">Args: {prompt.arguments?.length || 0}</span>
                  <Switch
                        checked={prompt.enabled}
                        onCheckedChange={() => handleToggle(prompt)}
                    />
              </div>
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  )
}
