/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";

interface Webhook {
  id: string;
  url: string;
  events: string[];
  active: boolean;
}

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);

  useEffect(() => {
    async function fetchWebhooks() {
      const res = await fetch("/api/webhooks");
      if (res.ok) {
        setWebhooks(await res.json());
      }
    }
    fetchWebhooks();
  }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
        <Button>Add Webhook</Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Configured Webhooks</CardTitle>
          <CardDescription>Receive notifications when events occur.</CardDescription>
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
              {webhooks.map((wh) => (
                <TableRow key={wh.id}>
                  <TableCell className="font-mono text-xs">{wh.url}</TableCell>
                  <TableCell>
                      <div className="flex gap-1">
                          {wh.events.map(e => <Badge key={e} variant="secondary" className="text-xs">{e}</Badge>)}
                      </div>
                  </TableCell>
                  <TableCell>
                      <Switch checked={wh.active} />
                  </TableCell>
                  <TableCell className="text-right">
                      <Button variant="ghost" size="sm">Test</Button>
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
