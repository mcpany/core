/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import { MoreHorizontal, Plus, Search, Filter } from "lucide-react"

// Mock Data
const initialServices = [
  { id: "srv-1", name: "github-service", type: "mcp", status: "online", enabled: true, endpoint: "http://localhost:8080" },
  { id: "srv-2", name: "slack-integration", type: "http", status: "online", enabled: true, endpoint: "https://api.slack.com" },
  { id: "srv-3", name: "postgres-db", type: "sql", status: "online", enabled: true, endpoint: "postgres://user:pass@localhost:5432/db" },
  { id: "srv-4", name: "legacy-api", type: "openapi", status: "offline", enabled: false, endpoint: "http://legacy:3000/spec.json" },
]

export default function ServicesPage() {
  const [services, setServices] = useState(initialServices)
  const [filter, setFilter] = useState("")

  const filteredServices = services.filter(s =>
    s.name.toLowerCase().includes(filter.toLowerCase()) ||
    s.type.toLowerCase().includes(filter.toLowerCase())
  )

  const toggleService = (id: string) => {
    setServices(services.map(s =>
      s.id === id ? { ...s, enabled: !s.enabled } : s
    ))
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Services</h2>
          <p className="text-muted-foreground">Manage your upstream MCP services.</p>
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" /> Add Service
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search services..."
            className="pl-8"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>
        <Button variant="outline" size="icon">
            <Filter className="h-4 w-4" />
        </Button>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Endpoint</TableHead>
              <TableHead>Enabled</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredServices.map((service) => (
              <TableRow key={service.id}>
                <TableCell className="font-medium">{service.name}</TableCell>
                <TableCell>
                    <Badge variant="secondary" className="uppercase text-xs">{service.type}</Badge>
                </TableCell>
                <TableCell>
                    <div className="flex items-center gap-2">
                        <span className={`h-2 w-2 rounded-full ${service.status === 'online' ? 'bg-green-500' : 'bg-red-500'}`} />
                        <span className="capitalize">{service.status}</span>
                    </div>
                </TableCell>
                <TableCell className="text-muted-foreground font-mono text-xs truncate max-w-[200px]">
                    {service.endpoint}
                </TableCell>
                <TableCell>
                    <Switch
                        checked={service.enabled}
                        onCheckedChange={() => toggleService(service.id)}
                    />
                </TableCell>
                <TableCell className="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon">
                        <MoreHorizontal className="h-4 w-4" />
                        <span className="sr-only">Open menu</span>
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Actions</DropdownMenuLabel>
                      <DropdownMenuItem>View Details</DropdownMenuItem>
                      <DropdownMenuItem>Edit Configuration</DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-destructive">Delete Service</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
