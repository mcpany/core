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
import { Search, FileText } from "lucide-react"

const initialResources = [
  { id: "res-1", name: "system-logs", description: "System logs for the last 24h", type: "text/plain", service: "system" },
  { id: "res-2", name: "api-schema", description: "OpenAPI schema for the legacy API", type: "application/json", service: "legacy-api" },
]

export default function ResourcesPage() {
  const [resources] = useState(initialResources)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Resources</h2>
          <p className="text-muted-foreground">Manage and view resources.</p>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Search resources..." className="pl-8" />
        </div>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Service</TableHead>
              <TableHead className="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {resources.map((res) => (
              <TableRow key={res.id}>
                <TableCell className="font-medium flex items-center gap-2">
                    <FileText className="h-4 w-4 text-muted-foreground" />
                    {res.name}
                </TableCell>
                <TableCell className="text-muted-foreground">{res.description}</TableCell>
                <TableCell className="font-mono text-xs">{res.type}</TableCell>
                <TableCell>{res.service}</TableCell>
                <TableCell className="text-right">
                    <Button variant="outline" size="sm">View</Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
