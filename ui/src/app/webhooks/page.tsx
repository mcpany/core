"use client"

import * as React from "react"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Webhook, Plus } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { Checkbox } from "@/components/ui/checkbox"

interface WebhookItem {
    id: string
    url: string
    events: string[]
    active: boolean
}

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = React.useState<WebhookItem[]>([])
  const [newUrl, setNewUrl] = React.useState("")
  const { toast } = useToast()

  React.useEffect(() => {
    fetchWebhooks()
  }, [])

  const fetchWebhooks = async () => {
      try {
        const res = await fetch('/api/webhooks')
        const data = await res.json()
        setWebhooks(data)
      } catch (e) {}
  }

  const handleAdd = async () => {
      if (!newUrl) return
      try {
          await fetch('/api/webhooks', {
              method: 'POST',
              body: JSON.stringify({ url: newUrl, events: ["all"] })
          })
          setNewUrl("")
          fetchWebhooks()
          toast({ title: "Webhook added" })
      } catch (e) {
          toast({ title: "Error", variant: "destructive" })
      }
  }

  return (
    <div className="space-y-8 max-w-4xl">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
        <p className="text-muted-foreground mt-2">
          Configure external event notifications.
        </p>
      </div>

      <GlassCard className="p-6">
          <div className="flex gap-4 items-end">
              <div className="flex-1 space-y-2">
                  <Label>Endpoint URL</Label>
                  <Input placeholder="https://api.example.com/webhook" value={newUrl} onChange={e => setNewUrl(e.target.value)} />
              </div>
              <Button onClick={handleAdd}>Add Webhook</Button>
          </div>
      </GlassCard>

      <div className="grid gap-4">
          {webhooks.map(wh => (
              <GlassCard key={wh.id} className="p-6 flex items-start justify-between">
                  <div className="flex items-start gap-4">
                      <div className="h-10 w-10 rounded-lg bg-pink-500/10 flex items-center justify-center text-pink-500">
                          <Webhook className="w-5 h-5" />
                      </div>
                      <div>
                          <h3 className="font-mono text-sm font-medium">{wh.url}</h3>
                          <div className="flex gap-2 mt-2">
                              {wh.events.map(e => (
                                  <span key={e} className="text-xs bg-slate-100 px-2 py-0.5 rounded border text-muted-foreground">{e}</span>
                              ))}
                          </div>
                      </div>
                  </div>
                  <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600 hover:bg-red-50">Delete</Button>
              </GlassCard>
          ))}
      </div>
    </div>
  )
}
