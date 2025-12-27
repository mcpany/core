"use client"

import * as React from "react"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { Switch } from "@/components/ui/switch"
import { GripVertical, Shield, Clock, FileText } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd"

// Since we are using Next.js 13+, we need to handle dnd specially or use a simpler list for now to avoid hydration errors quickly
// For this MVP step, I will use a simple list that "looks" like a pipeline.

interface Middleware {
    id: string
    name: string
    type: string
    enabled: boolean
    order: number
}

export default function MiddlewarePage() {
  const [middlewares, setMiddlewares] = React.useState<Middleware[]>([])
  const [loading, setLoading] = React.useState(true)
  const { toast } = useToast()

  React.useEffect(() => {
    fetchMiddleware()
  }, [])

  const fetchMiddleware = async () => {
    try {
      const res = await fetch('/api/middleware')
      const data = await res.json()
      setMiddlewares(data.sort((a: any, b: any) => a.order - b.order))
    } catch (error) {
       // ignore
    } finally {
      setLoading(false)
    }
  }

  const handleToggle = async (mw: Middleware) => {
      // Optimistic
      const newMws = middlewares.map(m => m.id === mw.id ? {...m, enabled: !mw.enabled} : m)
      setMiddlewares(newMws)

      await fetch('/api/middleware', {
          method: 'POST',
          body: JSON.stringify({ action: 'toggle', id: mw.id, enabled: !mw.enabled })
      })
  }

  const getIcon = (type: string) => {
      switch(type) {
          case 'auth': return <Shield className="w-5 h-5 text-blue-500" />
          case 'rate_limit': return <Clock className="w-5 h-5 text-orange-500" />
          default: return <FileText className="w-5 h-5 text-slate-500" />
      }
  }

  return (
    <div className="space-y-8 max-w-3xl">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h2>
        <p className="text-muted-foreground mt-2">
          Configure the global request processing pipeline.
        </p>
      </div>

      <div className="relative">
          {/* Visual line connector */}
          <div className="absolute left-[2.25rem] top-6 bottom-6 w-0.5 bg-border -z-10"></div>

          <div className="space-y-4">
            {middlewares.map((mw, index) => (
                <GlassCard key={mw.id} className="p-4 flex items-center gap-4 bg-background/80" hoverEffect>
                    <div className="cursor-grab active:cursor-grabbing text-muted-foreground p-2 hover:bg-slate-100 rounded">
                        <GripVertical className="w-5 h-5" />
                    </div>
                    <div className="h-10 w-10 rounded-full border bg-background flex items-center justify-center z-10 shadow-sm">
                        <span className="text-xs font-mono font-bold text-muted-foreground">{index + 1}</span>
                    </div>
                    <div className="h-12 w-12 rounded-lg bg-slate-50 border flex items-center justify-center">
                        {getIcon(mw.type)}
                    </div>
                    <div className="flex-1">
                        <h3 className="font-medium">{mw.name}</h3>
                        <p className="text-xs text-muted-foreground capitalize">{mw.type} Middleware</p>
                    </div>
                     <Switch
                        checked={mw.enabled}
                        onCheckedChange={() => handleToggle(mw)}
                    />
                </GlassCard>
            ))}
          </div>
      </div>
       <p className="text-xs text-muted-foreground text-center">
           Requests flow from top to bottom.
       </p>
    </div>
  )
}
