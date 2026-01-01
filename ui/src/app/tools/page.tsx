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
import { Badge } from "@/components/ui/badge"
import { Search, Filter, Wrench } from "lucide-react"

// Mock Data
const initialTools = [
  { id: "tool-1", name: "github-service/list_repos", description: "List repositories for a user", service: "github-service", tags: ["git", "source-control"] },
  { id: "tool-2", name: "github-service/create_issue", description: "Create a new issue", service: "github-service", tags: ["git", "issue-tracking"] },
  { id: "tool-3", name: "postgres-db/query", description: "Execute a SQL query", service: "postgres-db", tags: ["database", "sql"] },
  { id: "tool-4", name: "slack-integration/send_message", description: "Send a message to a channel", service: "slack-integration", tags: ["communication"] },
]

export default function ToolsPage() {
  const [tools] = useState(initialTools)
  const [filter, setFilter] = useState("")

  const filteredTools = tools.filter(t =>
    t.name.toLowerCase().includes(filter.toLowerCase()) ||
    t.description.toLowerCase().includes(filter.toLowerCase())
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Tools</h2>
          <p className="text-muted-foreground">Explore available MCP tools.</p>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search tools..."
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
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
              <TableHead>Tags</TableHead>
              <TableHead className="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTools.map((tool) => (
              <TableRow key={tool.id}>
                <TableCell className="font-medium flex items-center gap-2">
                    <Wrench className="h-4 w-4 text-muted-foreground" />
                    {tool.name}
                </TableCell>
                <TableCell className="text-muted-foreground">{tool.description}</TableCell>
                <TableCell>{tool.service}</TableCell>
                <TableCell>
                    <div className="flex gap-1 flex-wrap">
                        {tool.tags.map(tag => (
                            <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
                        ))}
                    </div>
                </TableCell>
                <TableCell className="text-right">
                    <Button variant="outline" size="sm">Test</Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
