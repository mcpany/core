/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

export default function SettingsPage() {
  const [webhooks, setWebhooks] = useState([
    { id: 1, url: "https://slack.com/api/webhooks/xyz", events: ["service.up", "service.down"], active: true }
  ]);
  const [newWebhookUrl, setNewWebhookUrl] = useState("");

  const addWebhook = () => {
    if (newWebhookUrl) {
      setWebhooks([...webhooks, { id: Date.now(), url: newWebhookUrl, events: ["all"], active: true }]);
      setNewWebhookUrl("");
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
      </div>
      <Tabs defaultValue="profiles" className="space-y-4">
        <TabsList>
          <TabsTrigger value="profiles">Profiles</TabsTrigger>
          <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
          <TabsTrigger value="middleware">Middleware</TabsTrigger>
        </TabsList>

        {/* PROFILES */}
        <TabsContent value="profiles" className="space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle>Execution Profiles</CardTitle>
                    <CardDescription>
                        Manage environments like Development, Production, and Debug.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between border p-4 rounded-lg">
                        <div>
                            <div className="flex items-center gap-2">
                                <p className="font-medium">Development</p>
                                <Badge variant="secondary">Active</Badge>
                            </div>
                            <p className="text-sm text-muted-foreground">Verbose logging, auto-reload, local mock services.</p>
                        </div>
                        <Button variant="outline">Edit</Button>
                    </div>
                    <div className="flex items-center justify-between border p-4 rounded-lg opacity-60">
                        <div>
                            <p className="font-medium">Production</p>
                            <p className="text-sm text-muted-foreground">Optimized for performance, strict security, audit logging.</p>
                        </div>
                        <Button variant="outline">Edit</Button>
                    </div>
                </CardContent>
            </Card>
             <Card>
                <CardHeader>
                    <CardTitle>Global Configuration</CardTitle>
                    <CardDescription>Server-wide settings.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                     <div className="grid grid-cols-2 gap-4">
                         <div className="space-y-2">
                            <Label>Log Level</Label>
                            <Select defaultValue="info">
                                <SelectTrigger>
                                    <SelectValue placeholder="Select log level" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="debug">DEBUG</SelectItem>
                                    <SelectItem value="info">INFO</SelectItem>
                                    <SelectItem value="warn">WARN</SelectItem>
                                    <SelectItem value="error">ERROR</SelectItem>
                                </SelectContent>
                            </Select>
                         </div>
                          <div className="space-y-2">
                            <Label>Audit Logging</Label>
                             <div className="flex items-center space-x-2 pt-2">
                                <Switch id="audit-mode" defaultChecked />
                                <Label htmlFor="audit-mode">Enabled</Label>
                             </div>
                         </div>
                     </div>
                </CardContent>
            </Card>
        </TabsContent>

        {/* WEBHOOKS */}
        <TabsContent value="webhooks" className="space-y-4">
             <Card>
                <CardHeader>
                    <CardTitle>Webhooks</CardTitle>
                    <CardDescription>
                        Configure and test webhooks for event notifications.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex gap-4 items-end border-b pb-6">
                         <div className="grid w-full max-w-sm items-center gap-1.5">
                            <Label htmlFor="webhook-url">Endpoint URL</Label>
                            <Input
                                type="url"
                                id="webhook-url"
                                placeholder="https://api.example.com/webhook"
                                value={newWebhookUrl}
                                onChange={(e) => setNewWebhookUrl(e.target.value)}
                            />
                        </div>
                        <Button onClick={addWebhook}>Add Webhook</Button>
                    </div>

                    <div className="space-y-4">
                        {webhooks.map(webhook => (
                             <div key={webhook.id} className="flex items-center justify-between border p-4 rounded-lg">
                                <div>
                                    <p className="font-mono text-sm">{webhook.url}</p>
                                    <div className="flex gap-2 mt-2">
                                        {webhook.events.map(ev => <Badge key={ev} variant="outline">{ev}</Badge>)}
                                    </div>
                                </div>
                                <div className="flex items-center gap-4">
                                     <Switch checked={webhook.active} />
                                     <Button variant="ghost" size="sm" className="text-red-500">Remove</Button>
                                </div>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </TabsContent>

        {/* MIDDLEWARE */}
        <TabsContent value="middleware" className="space-y-4">
             <Card>
                <CardHeader>
                    <CardTitle>Middleware Pipeline</CardTitle>
                    <CardDescription>
                        Visual management of the request processing pipeline. Drag to reorder (simulated).
                    </CardDescription>
                </CardHeader>
                <CardContent>
                     {/* Visual representation of a pipeline */}
                    <div className="relative flex items-center justify-between p-8 bg-muted/20 rounded-xl overflow-x-auto">
                         <div className="absolute left-0 right-0 top-1/2 h-1 bg-muted-foreground/20 -z-10 transform -translate-y-1/2"></div>

                         {["Authentication", "Rate Limit", "Logging", "Routing", "Transformation"].map((step, i) => (
                             <div key={step} className="flex flex-col items-center gap-2 z-10 min-w-[120px]">
                                 <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold shadow-lg ring-4 ring-background">
                                     {i + 1}
                                 </div>
                                 <div className="bg-card border p-3 rounded-lg shadow-sm w-full text-center hover:shadow-md transition-shadow cursor-grab active:cursor-grabbing">
                                     <span className="font-medium text-sm">{step}</span>
                                 </div>
                             </div>
                         ))}
                    </div>

                    <div className="mt-8 grid grid-cols-2 gap-4">
                        <div className="border p-4 rounded-lg">
                            <h4 className="font-semibold mb-2">Available Middleware</h4>
                            <div className="space-y-2">
                                <div className="flex justify-between items-center p-2 bg-muted/50 rounded text-sm">
                                    <span>Compression (Gzip)</span>
                                    <Button size="sm" variant="ghost">+</Button>
                                </div>
                                <div className="flex justify-between items-center p-2 bg-muted/50 rounded text-sm">
                                    <span>IP Whitelist</span>
                                    <Button size="sm" variant="ghost">+</Button>
                                </div>
                            </div>
                        </div>
                        <div className="border p-4 rounded-lg">
                             <h4 className="font-semibold mb-2">Configuration</h4>
                             <p className="text-sm text-muted-foreground">Select a middleware node above to configure it.</p>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
