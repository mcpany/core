"use client"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Plus, Webhook } from "lucide-react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"

export default function WebhooksPage() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
          <p className="text-muted-foreground">Configure and test outgoing webhooks.</p>
        </div>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Add Webhook
        </Button>
      </div>

      <Card>
          <CardHeader>
              <CardTitle>Configured Webhooks</CardTitle>
              <CardDescription>
                  List of active webhooks.
              </CardDescription>
          </CardHeader>
          <CardContent>
               <Table>
                <TableHeader>
                    <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Target URL</TableHead>
                    <TableHead>Events</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                     <TableRow>
                        <TableCell className="font-medium">Slack Notification</TableCell>
                        <TableCell>https://hooks.slack.com/services/...</TableCell>
                        <TableCell>
                            <div className="flex gap-1">
                                <Badge variant="secondary">service.down</Badge>
                                <Badge variant="secondary">error.critical</Badge>
                            </div>
                        </TableCell>
                        <TableCell>
                             <Badge variant="outline" className="text-green-600 border-green-600">Active</Badge>
                        </TableCell>
                        <TableCell className="text-right">
                            <Button variant="ghost" size="sm">Edit</Button>
                            <Button variant="ghost" size="sm">Test</Button>
                        </TableCell>
                    </TableRow>
                     <TableRow>
                        <TableCell className="font-medium">Audit Log Dump</TableCell>
                        <TableCell>https://audit.internal/ingest</TableCell>
                        <TableCell>
                             <Badge variant="secondary">*</Badge>
                        </TableCell>
                        <TableCell>
                             <Badge variant="outline" className="text-yellow-600 border-yellow-600">Paused</Badge>
                        </TableCell>
                        <TableCell className="text-right">
                            <Button variant="ghost" size="sm">Edit</Button>
                            <Button variant="ghost" size="sm">Test</Button>
                        </TableCell>
                    </TableRow>
                </TableBody>
               </Table>
          </CardContent>
      </Card>
    </div>
  )
}
