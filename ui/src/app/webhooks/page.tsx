/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Webhook, Plus, Play, Trash2, Loader2 } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter
} from "@/components/ui/dialog"
import { toast } from "sonner";

interface WebhookConfig {
    id: string;
    url: string;
    events: string[];
    active: boolean;
    last_triggered?: string;
    status?: string;
}

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<WebhookConfig[]>([]);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [newUrl, setNewUrl] = useState("");
    const [isLoading, setIsLoading] = useState(false);

    const fetchWebhooks = async () => {
        setIsLoading(true);
        try {
            const res = await fetch("/api/v1/webhooks");
            if (res.ok) {
                const data = await res.json();
                setWebhooks(data);
            } else {
                toast.error("Failed to load webhooks");
            }
        } catch (error) {
            toast.error("Failed to load webhooks");
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchWebhooks();
    }, []);

    const toggleWebhook = async (id: string) => {
        // Toggle active status not implemented in backend yet, just placeholder for UI interaction
        // In real impl, this would be a PATCH or PUT
        toast.info("Toggle active status not yet implemented in backend");
    };

    const deleteWebhook = async (id: string) => {
        try {
            const res = await fetch(`/api/v1/webhooks/${id}`, {
                method: "DELETE"
            });
            if (res.ok) {
                toast.success("Webhook deleted");
                fetchWebhooks();
            } else {
                toast.error("Failed to delete webhook");
            }
        } catch (error) {
            toast.error("Failed to delete webhook");
        }
    };

    const addWebhook = async () => {
        if (!newUrl) return;
        try {
            const res = await fetch("/api/v1/webhooks", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    url: newUrl,
                    events: ["all"], // Default event
                    active: true
                })
            });
            if (res.ok) {
                toast.success("Webhook added");
                setNewUrl("");
                setIsDialogOpen(false);
                fetchWebhooks();
            } else {
                toast.error("Failed to add webhook");
            }
        } catch (error) {
            toast.error("Failed to add webhook");
        }
    }

    const testWebhook = async (id: string) => {
        const loadingToast = toast.loading("Testing webhook...");
        try {
            const res = await fetch(`/api/v1/webhooks/${id}/test`, {
                method: "POST"
            });
            toast.dismiss(loadingToast);
            if (res.ok) {
                toast.success("Webhook test delivered successfully");
                fetchWebhooks(); // Refresh status
            } else {
                toast.error("Webhook test failed");
                fetchWebhooks(); // Refresh status
            }
        } catch (error) {
            toast.dismiss(loadingToast);
            toast.error("Webhook test error");
        }
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
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
                            {isLoading && webhooks.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24">
                                        <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                                    </TableCell>
                                </TableRow>
                            ) : webhooks.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                        No webhooks configured.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                webhooks.map((hook) => (
                                    <TableRow key={hook.id}>
                                        <TableCell className="font-mono text-xs max-w-[300px] truncate" title={hook.url}>{hook.url}</TableCell>
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
                                                    {hook.status ? hook.status.toUpperCase() : (hook.active ? "ACTIVE" : "INACTIVE")}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-muted-foreground text-xs">
                                            {hook.last_triggered ? new Date(hook.last_triggered).toLocaleString() : "Never"}
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
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
