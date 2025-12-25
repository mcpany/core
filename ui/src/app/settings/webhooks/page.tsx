/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Plus, Webhook as WebhookIcon, Play } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

interface Webhook {
  id: string;
  url: string;
  events: string[];
  status: "Active" | "Inactive";
}

const initialWebhooks: Webhook[] = [
  { id: "1", url: "https://api.slack.com/webhook/...", events: ["service.down", "error.critical"], status: "Active" },
  { id: "2", url: "https://pagerduty.com/...", events: ["service.down"], status: "Active" },
];

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = useState(initialWebhooks);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
         <Dialog>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Add Webhook
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                <DialogTitle>Add Webhook</DialogTitle>
                <DialogDescription>
                    Configure a new webhook endpoint to receive events.
                </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="url" className="text-right">
                    URL
                    </Label>
                    <Input id="url" placeholder="https://..." className="col-span-3" />
                </div>
                 <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="events" className="text-right">
                    Events
                    </Label>
                    <Input id="events" placeholder="service.down, ..." className="col-span-3" />
                </div>
                </div>
                <DialogFooter>
                <Button type="submit">Save changes</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>

      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Configured Webhooks</CardTitle>
          <CardDescription>Manage outbound webhooks for system events.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>URL</TableHead>
                <TableHead>Events</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {webhooks.map((webhook) => (
                <TableRow key={webhook.id}>
                  <TableCell className="font-mono text-xs">{webhook.url}</TableCell>
                  <TableCell>
                      <div className="flex gap-1 flex-wrap">
                        {webhook.events.map(e => <Badge key={e} variant="secondary" className="text-xs">{e}</Badge>)}
                      </div>
                  </TableCell>
                  <TableCell>
                      <Badge variant={webhook.status === "Active" ? "default" : "secondary"}>{webhook.status}</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="outline" size="sm">
                        <Play className="mr-2 h-3 w-3" /> Test
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
