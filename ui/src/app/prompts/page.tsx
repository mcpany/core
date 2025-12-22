"use client"

import { useEffect, useState } from "react"
import { apiClient } from "@/lib/mock-client"
import { UpstreamServiceConfig, PromptDefinition } from "@/lib/types"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"

interface PromptWithService extends PromptDefinition {
    serviceName: string;
    serviceId: string;
}

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<PromptWithService[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchPrompts = async () => {
      try {
        const { services } = await apiClient.listServices()
        const allPrompts: PromptWithService[] = []
        services.forEach(service => {
            const servicePrompts = service.grpc_service?.prompts || service.http_service?.prompts || service.command_line_service?.prompts || []
            servicePrompts.forEach(prompt => {
                allPrompts.push({
                    ...prompt,
                    serviceName: service.name,
                    serviceId: service.id || ""
                })
            })
        })
        setPrompts(allPrompts)
      } catch (error) {
        console.error("Failed to fetch prompts", error)
      } finally {
        setLoading(false)
      }
    }
    fetchPrompts()
  }, [])

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Prompts</h1>
        <p className="text-muted-foreground">Manage AI prompts and templates.</p>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {prompts.map((prompt, index) => (
              <TableRow key={`${prompt.serviceId}-${prompt.name}-${index}`}>
                <TableCell className="font-medium">{prompt.name}</TableCell>
                <TableCell className="text-muted-foreground">{prompt.description || "-"}</TableCell>
                <TableCell>
                    <Badge variant="secondary">{prompt.serviceName}</Badge>
                </TableCell>
              </TableRow>
            ))}
             {loading && (
                 <TableRow>
                     <TableCell colSpan={3} className="text-center h-24">
                        Loading prompts...
                     </TableCell>
                 </TableRow>
            )}
             {!loading && prompts.length === 0 && (
                <TableRow>
                    <TableCell colSpan={3} className="text-center h-24">
                        No prompts found.
                    </TableCell>
                </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
