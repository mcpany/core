"use client"

import * as React from "react"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { StatusBadge } from "@/components/ui-custom/status-badge"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { Plus, MoreVertical, Edit, Trash, Server } from "lucide-react"
import Link from "next/link"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { apiClient, UpstreamServiceConfig } from "@/lib/client"
import { useToast } from "@/components/ui/use-toast"

export default function ServicesPage() {
  const [services, setServices] = React.useState<UpstreamServiceConfig[]>([])
  const [loading, setLoading] = React.useState(true)
  const { toast } = useToast()

  React.useEffect(() => {
    fetchServices()
  }, [])

  const fetchServices = async () => {
    try {
      const data = await apiClient.listServices()
      // Adapt the response to match the expected array structure if needed
      // The mock returns { services: [...] }
      if (data.services) {
          setServices(data.services)
      } else if (Array.isArray(data)) {
          setServices(data)
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch services",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleToggle = async (service: UpstreamServiceConfig) => {
    const newState = !service.disable // disable = true means disabled. So if currently disabled (true), new state is false (enabled).
    // Wait, the toggle UI usually shows "Enabled".
    // If disable=false, UI is ON.
    // If user clicks ON -> OFF, we set disable=true.
    // If user clicks OFF -> ON, we set disable=false.

    try {
       // Optimistic update
      setServices(services.map(s => s.name === service.name ? { ...s, disable: !service.disable } : s))

      await apiClient.setServiceStatus(service.name, !service.disable)
       toast({
        title: "Success",
        description: `Service ${!service.disable ? 'disabled' : 'enabled'} successfully`,
      })
    } catch (error) {
       // Revert
       setServices(services.map(s => s.name === service.name ? { ...s, disable: service.disable } : s))
       toast({
        title: "Error",
        description: "Failed to update status",
        variant: "destructive",
      })
    }
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Services</h2>
          <p className="text-muted-foreground mt-2">
            Manage your upstream MCP services and connections.
          </p>
        </div>
        <Link href="/services/new">
          <Button className="gap-2">
            <Plus className="w-4 h-4" /> Add Service
          </Button>
        </Link>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
           <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div className="grid gap-6">
          {services.map((service) => (
            <GlassCard key={service.id || service.name} className="p-6 flex items-center justify-between" hoverEffect>
              <div className="flex items-center gap-4">
                 <div className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
                    <Server className="w-6 h-6" />
                 </div>
                 <div>
                    <h3 className="font-semibold text-lg">{service.name}</h3>
                    <div className="flex items-center gap-3 text-sm text-muted-foreground mt-1">
                        <StatusBadge status={!service.disable} text={!service.disable ? "Enabled" : "Disabled"} />
                        <span>v{service.version || "1.0.0"}</span>
                        <span>•</span>
                        <span className="font-mono text-xs bg-slate-100 px-2 py-0.5 rounded border">
                            {service.http_service?.address || service.grpc_service?.address || service.command_line_service?.command || "Managed"}
                        </span>
                    </div>
                 </div>
              </div>

              <div className="flex items-center gap-6">
                 <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-muted-foreground">Active</span>
                    <Switch
                        checked={!service.disable}
                        onCheckedChange={() => handleToggle(service)}
                    />
                 </div>

                 <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0">
                      <span className="sr-only">Open menu</span>
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                    <Link href={`/services/${service.id || service.name}`}>
                         <DropdownMenuItem>
                            <Edit className="mr-2 h-4 w-4" /> Edit Configuration
                        </DropdownMenuItem>
                    </Link>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-red-600">
                        <Trash className="mr-2 h-4 w-4" /> Delete Service
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </GlassCard>
          ))}

          {services.length === 0 && (
              <div className="text-center py-20 text-muted-foreground">
                  No services found. Add one to get started.
              </div>
          )}
        </div>
      )}
    </div>
  )
}
