"use client"

import { useEffect, useState } from "react"
import { apiClient } from "@/lib/mock-client"
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Search } from "lucide-react"

interface ToolWithService extends ToolDefinition {
    serviceName: string;
    serviceId: string;
}

export default function ToolsPage() {
  const [tools, setTools] = useState<ToolWithService[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState("")

  useEffect(() => {
    const fetchTools = async () => {
      try {
        const { services } = await apiClient.listServices()
        const allTools: ToolWithService[] = []
        services.forEach(service => {
            const serviceTools = service.grpc_service?.tools || service.http_service?.tools || service.command_line_service?.tools || []
            serviceTools.forEach(tool => {
                allTools.push({
                    ...tool,
                    serviceName: service.name,
                    serviceId: service.id || ""
                })
            })
        })
        setTools(allTools)
      } catch (error) {
        console.error("Failed to fetch tools", error)
      } finally {
        setLoading(false)
      }
    }
    fetchTools()
  }, [])

  const filteredTools = tools.filter(tool =>
    tool.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    tool.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    tool.serviceName.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Tools</h1>
        <p className="text-muted-foreground">Browse and manage tools across all services.</p>
      </div>

      <div className="flex w-full max-w-sm items-center space-x-2">
        <Input
            placeholder="Search tools..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
        />
        <Button size="icon" variant="ghost">
             <Search className="h-4 w-4" />
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
              <TableHead>Source</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTools.map((tool, index) => (
              <TableRow key={`${tool.serviceId}-${tool.name}-${index}`}>
                <TableCell className="font-medium">{tool.name}</TableCell>
                <TableCell className="text-muted-foreground">{tool.description || "-"}</TableCell>
                <TableCell>
                    <Badge variant="secondary">{tool.serviceName}</Badge>
                </TableCell>
                <TableCell>
                    <Badge variant="outline">{tool.source || "configured"}</Badge>
                </TableCell>
              </TableRow>
            ))}
             {loading && (
                 <TableRow>
                     <TableCell colSpan={4} className="text-center h-24">
                        Loading tools...
                     </TableCell>
                 </TableRow>
            )}
             {!loading && filteredTools.length === 0 && (
                <TableRow>
                    <TableCell colSpan={4} className="text-center h-24">
                        No tools found.
                    </TableCell>
                </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
