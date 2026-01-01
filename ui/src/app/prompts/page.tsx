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
import { Search, MessageSquare } from "lucide-react"

const initialPrompts = [
  { id: "prm-1", name: "summarize_logs", description: "Summarize the last N lines of logs", service: "system" },
  { id: "prm-2", name: "explain_schema", description: "Explain the database schema", service: "postgres-db" },
]

export default function PromptsPage() {
  const [prompts] = useState(initialPrompts)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Prompts</h2>
          <p className="text-muted-foreground">Manage and view prompts.</p>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Search prompts..." className="pl-8" />
        </div>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
              <TableHead className="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {prompts.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-medium flex items-center gap-2">
                    <MessageSquare className="h-4 w-4 text-muted-foreground" />
                    {p.name}
                </TableCell>
                <TableCell className="text-muted-foreground">{p.description}</TableCell>
                <TableCell>{p.service}</TableCell>
                <TableCell className="text-right">
                    <Button variant="outline" size="sm">Execute</Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
