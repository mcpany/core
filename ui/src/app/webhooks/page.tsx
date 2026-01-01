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
import { Search, Webhook } from "lucide-react"

const initialWebhooks = [
  { id: "wh-1", url: "https://api.example.com/hooks/pre-call", type: "PreCall", status: "active", failures: 0 },
  { id: "wh-2", url: "https://audit.example.com/log", type: "PostCall", status: "active", failures: 2 },
]

export default function WebhooksPage() {
  const [webhooks] = useState(initialWebhooks)

  return (
    <div className="space-y-4">
       <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Webhooks</h2>
          <p className="text-muted-foreground">Manage and test webhook integrations.</p>
        </div>
        <Button>Add Webhook</Button>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Search webhooks..." className="pl-8" />
        </div>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>URL</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Failures</TableHead>
              <TableHead className="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {webhooks.map((wh) => (
              <TableRow key={wh.id}>
                <TableCell className="font-medium flex items-center gap-2">
                    <Webhook className="h-4 w-4 text-muted-foreground" />
                    {wh.url}
                </TableCell>
                <TableCell>
                    <Badge variant="outline">{wh.type}</Badge>
                </TableCell>
                <TableCell>
                     <span className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ring-1 ring-inset ${wh.status === 'active' ? 'bg-green-50 text-green-700 ring-green-600/20' : 'bg-red-50 text-red-700 ring-red-600/20'}`}>
                        {wh.status}
                     </span>
                </TableCell>
                <TableCell>{wh.failures}</TableCell>
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
