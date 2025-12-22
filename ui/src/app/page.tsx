"use client"

import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { apiClient } from "@/lib/mock-client"
import { UpstreamServiceConfig } from "@/lib/types"
import { Activity, AlertCircle, CheckCircle2, Clock, Globe, Box, MessageSquare } from "lucide-react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"

// Mock real-time metrics for visualization
const generateHistory = (length: number) => Array.from({ length }, () => Math.floor(Math.random() * 100));

export default function DashboardPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const { services } = await apiClient.listServices()
        setServices(services)
      } catch (error) {
        console.error("Failed to fetch services", error)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  // Calculate high-level stats
  const activeServices = services.filter(s => !s.disable).length
  const totalServices = services.length
  // For demo purposes, counting tools/prompts/resources from mock data structure if available
  // In a real app, we'd aggregate this from all services
  const totalTools = services.reduce((acc, s) => acc + (s.grpc_service?.tools?.length || s.http_service?.tools?.length || s.command_line_service?.tools?.length || 0), 0)

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">Platform overview and real-time metrics.</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Services</CardTitle>
            <Globe className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeServices} / {totalServices}</div>
            <p className="text-xs text-muted-foreground">
              {totalServices - activeServices} disabled
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Tools</CardTitle>
            <Box className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalTools}</div>
            <p className="text-xs text-muted-foreground">
              Across all services
            </p>
          </CardContent>
        </Card>
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Global Request Rate</CardTitle>
                <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">~1.2k req/s</div>
                <p className="text-xs text-muted-foreground">
                    +12% from last hour
                </p>
            </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">42ms</div>
            <p className="text-xs text-muted-foreground">
              p99: 120ms
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>Service Status</CardTitle>
            <CardDescription>
              Live health checks for connected MCP servers.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Service</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead className="text-right">Latency</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {services.map((service) => (
                        <TableRow key={service.id}>
                            <TableCell className="font-medium">
                                <div className="flex flex-col">
                                    <span>{service.name}</span>
                                    <span className="text-xs text-muted-foreground">{service.version}</span>
                                </div>
                            </TableCell>
                            <TableCell>
                                <Badge variant="outline">
                                    {service.grpc_service ? "gRPC" :
                                     service.http_service ? "HTTP" :
                                     service.command_line_service ? "CMD" : "Other"}
                                </Badge>
                            </TableCell>
                            <TableCell>
                                {service.disable ? (
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <AlertCircle className="h-4 w-4" />
                                        <span>Disabled</span>
                                    </div>
                                ) : (
                                    <div className="flex items-center gap-2 text-green-600">
                                        <CheckCircle2 className="h-4 w-4" />
                                        <span>Healthy</span>
                                    </div>
                                )}
                            </TableCell>
                            <TableCell className="text-right">
                                {service.disable ? "-" : `${Math.floor(Math.random() * 50 + 10)}ms`}
                            </TableCell>
                        </TableRow>
                    ))}
                    {loading && (
                         <TableRow>
                             <TableCell colSpan={4} className="text-center h-24">
                                Loading services...
                             </TableCell>
                         </TableRow>
                    )}
                </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>
              Latest operations and errors.
            </CardDescription>
          </CardHeader>
          <CardContent>
             <div className="space-y-4">
                {[1,2,3,4,5].map((_, i) => (
                    <div key={i} className="flex items-start gap-4 text-sm">
                        <div className={`mt-1 h-2 w-2 rounded-full ${i === 2 ? 'bg-red-500' : 'bg-blue-500'}`} />
                        <div className="flex flex-col gap-1">
                            <span className="font-medium">{i === 2 ? 'Error: Connection Timeout' : 'Tool executed: payment-gateway/CreatePayment'}</span>
                            <span className="text-xs text-muted-foreground">{i * 2 + 1} mins ago</span>
                        </div>
                    </div>
                ))}
             </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
