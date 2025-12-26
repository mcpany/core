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
import { Textarea } from "@/components/ui/textarea";

export default function WebhooksPage() {
  const [testUrl, setTestUrl] = useState("");
  const [payload, setPayload] = useState('{\n  "event": "service_registered",\n  "data": {\n    "service_id": "svc-123"\n  }\n}');

  const sendTestWebhook = () => {
      // Logic to send test webhook
      alert(`Sending webhook to ${testUrl}`);
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
          <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader>
              <CardTitle>Configuration</CardTitle>
              <CardDescription>Configure global webhook settings.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
               <div className="grid gap-2">
                   <Label htmlFor="endpoint">Default Endpoint URL</Label>
                   <Input id="endpoint" placeholder="https://api.example.com/webhooks" />
               </div>
               <div className="grid gap-2">
                   <Label htmlFor="secret">Signing Secret</Label>
                   <Input id="secret" type="password" value="****************" disabled />
               </div>
               <Button>Save Configuration</Button>
            </CardContent>
          </Card>

          <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader>
              <CardTitle>Test Interface</CardTitle>
              <CardDescription>Send test events to your webhook endpoints.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="grid gap-2">
                   <Label htmlFor="test-url">Target URL</Label>
                   <Input id="test-url" value={testUrl} onChange={(e) => setTestUrl(e.target.value)} placeholder="https://..." />
               </div>
               <div className="grid gap-2">
                   <Label htmlFor="payload">JSON Payload</Label>
                   <Textarea
                        id="payload"
                        className="font-mono text-sm h-[200px]"
                        value={payload}
                        onChange={(e) => setPayload(e.target.value)}
                    />
               </div>
               <Button onClick={sendTestWebhook}>Send Test Event</Button>
            </CardContent>
          </Card>
      </div>
    </div>
  );
}
