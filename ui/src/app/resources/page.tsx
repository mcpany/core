"use client"

import { useEffect, useState } from "react"
import { apiClient } from "@/lib/mock-client"
import { UpstreamServiceConfig, ResourceDefinition } from "@/lib/types"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"

interface ResourceWithService extends ResourceDefinition {
    serviceName: string;
    serviceId: string;
}

export default function ResourcesPage() {
  const [resources, setResources] = useState<ResourceWithService[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchResources = async () => {
      try {
        const { services } = await apiClient.listServices()
        const allResources: ResourceWithService[] = []
        services.forEach(service => {
            const serviceResources = service.grpc_service?.resources || service.http_service?.resources || service.command_line_service?.resources || []
            serviceResources.forEach(res => {
                allResources.push({
                    ...res,
                    serviceName: service.name,
                    serviceId: service.id || ""
                })
            })
        })
        setResources(allResources)
      } catch (error) {
        console.error("Failed to fetch resources", error)
      } finally {
        setLoading(false)
      }
    }
    fetchResources()
  }, [])

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Resources</h1>
        <p className="text-muted-foreground">Manage data resources exposed by services.</p>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Service</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {resources.map((res, index) => (
              <TableRow key={`${res.serviceId}-${res.name}-${index}`}>
                <TableCell className="font-medium">{res.name}</TableCell>
                <TableCell>{res.type || "Unknown"}</TableCell>
                <TableCell>
                    <Badge variant="secondary">{res.serviceName}</Badge>
                </TableCell>
              </TableRow>
            ))}
             {loading && (
                 <TableRow>
                     <TableCell colSpan={3} className="text-center h-24">
                        Loading resources...
                     </TableCell>
                 </TableRow>
            )}
             {!loading && resources.length === 0 && (
                <TableRow>
                    <TableCell colSpan={3} className="text-center h-24">
                        No resources found.
                    </TableCell>
                </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
