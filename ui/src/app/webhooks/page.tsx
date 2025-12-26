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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Webhook, Plus, Play, Trash2 } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter
} from "@/components/ui/dialog"

interface WebhookConfig {
    id: string;
    url: string;
    events: string[];
    active: boolean;
    lastTriggered?: string;
    status?: "success" | "failure" | "pending";
}

const INITIAL_WEBHOOKS: WebhookConfig[] = [
    { id: "wh-1", url: "https://api.example.com/events", events: ["service.registered", "tool.invoked"], active: true, lastTriggered: "2 mins ago", status: "success" },
    { id: "wh-2", url: "https://hooks.slack.com/...", events: ["error.occurred"], active: true, lastTriggered: "1 day ago", status: "success" },
];

export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<WebhookConfig[]>(INITIAL_WEBHOOKS);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [newUrl, setNewUrl] = useState("");

    const toggleWebhook = (id: string) => {
        setWebhooks(webhooks.map(w => w.id === id ? { ...w, active: !w.active } : w));
    };

    const deleteWebhook = (id: string) => {
        setWebhooks(webhooks.filter(w => w.id !== id));
    };

    const addWebhook = () => {
        if (!newUrl) return;
        setWebhooks([...webhooks, {
            id: `wh-${Date.now()}`,
            url: newUrl,
            events: ["all"],
            active: true
        }]);
        setNewUrl("");
        setIsDialogOpen(false);
    }

    const testWebhook = (id: string) => {
        // Simulate test
        alert(`Testing webhook ${id}...`);
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
                    <p className="text-muted-foreground">Configure outbound webhooks for system events.</p>
                </div>
                <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                    <DialogTrigger asChild>
                         <Button>
                            <Plus className="mr-2 h-4 w-4" /> New Webhook
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Add Webhook</DialogTitle>
                            <DialogDescription>
                                Enter the URL where events should be sent.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid grid-cols-4 items-center gap-4">
                                <Label htmlFor="url" className="text-right">
                                    Payload URL
                                </Label>
                                <Input
                                    id="url"
                                    value={newUrl}
                                    onChange={(e) => setNewUrl(e.target.value)}
                                    placeholder="https://..."
                                    className="col-span-3"
                                />
                            </div>
                        </div>
                        <DialogFooter>
                            <Button onClick={addWebhook}>Add Webhook</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle>Configured Webhooks</CardTitle>
                </CardHeader>
                <CardContent>
                     <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>URL</TableHead>
                                <TableHead>Events</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Last Triggered</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {webhooks.map((hook) => (
                                <TableRow key={hook.id}>
                                    <TableCell className="font-mono text-xs">{hook.url}</TableCell>
                                    <TableCell>
                                        <div className="flex gap-1 flex-wrap">
                                            {hook.events.map(e => <Badge key={e} variant="secondary" className="text-xs">{e}</Badge>)}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            <Switch
                                                checked={hook.active}
                                                onCheckedChange={() => toggleWebhook(hook.id)}
                                            />
                                            <span className={`text-xs ${hook.status === 'success' ? 'text-green-500' : hook.status === 'failure' ? 'text-red-500' : 'text-muted-foreground'}`}>
                                                {hook.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-xs">
                                        {hook.lastTriggered || "Never"}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button variant="ghost" size="icon" onClick={() => testWebhook(hook.id)} title="Test Delivery">
                                                <Play className="h-4 w-4" />
                                            </Button>
                                             <Button variant="ghost" size="icon" onClick={() => deleteWebhook(hook.id)} className="text-red-500 hover:text-red-600">
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))}
                            {webhooks.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                        No webhooks configured.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
