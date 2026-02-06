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
import { Webhook as WebhookIcon, Plus, Play, Trash2, RefreshCw } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter
} from "@/components/ui/dialog"
import { apiClient, Webhook } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<Webhook[]>([]);
    const [loading, setLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [newUrl, setNewUrl] = useState("");
    const { toast } = useToast();

    const fetchWebhooks = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listWebhooks();
            setWebhooks(data);
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to load webhooks", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchWebhooks();
    }, []);

    const toggleWebhook = async (webhook: Webhook) => {
        try {
            await apiClient.updateWebhook({ ...webhook, active: !webhook.active });
            fetchWebhooks();
        } catch (e) {
            toast({ title: "Error", description: "Failed to update webhook", variant: "destructive" });
        }
    };

    const deleteWebhook = async (id: string) => {
        if (!confirm("Are you sure you want to delete this webhook?")) return;
        try {
            await apiClient.deleteWebhook(id);
            fetchWebhooks();
            toast({ title: "Success", description: "Webhook deleted" });
        } catch (e) {
            toast({ title: "Error", description: "Failed to delete webhook", variant: "destructive" });
        }
    };

    const addWebhook = async () => {
        if (!newUrl) return;
        try {
            await apiClient.createWebhook({
                url: newUrl,
                events: ["all"],
                active: true
            });
            setNewUrl("");
            setIsDialogOpen(false);
            fetchWebhooks();
            toast({ title: "Success", description: "Webhook created" });
        } catch (e) {
            toast({ title: "Error", description: "Failed to create webhook", variant: "destructive" });
        }
    }

    const testWebhook = (id: string) => {
        // Simulate test for now as we don't have a specific test endpoint
        // Or we could trigger a test alert?
        toast({ title: "Test Triggered", description: "A test event has been sent (simulation)." });
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
                    <p className="text-muted-foreground">Configure outbound webhooks for system events.</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="icon" onClick={fetchWebhooks}>
                        <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                    </Button>
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
                                <Button onClick={addWebhook} disabled={!newUrl}>Add Webhook</Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
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
                            {loading && webhooks.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                        Loading webhooks...
                                    </TableCell>
                                </TableRow>
                            )}
                            {!loading && webhooks.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                        No webhooks configured.
                                    </TableCell>
                                </TableRow>
                            )}
                            {webhooks.map((hook) => (
                                <TableRow key={hook.id}>
                                    <TableCell className="font-mono text-xs max-w-[300px] truncate" title={hook.url}>
                                        {hook.url}
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex gap-1 flex-wrap">
                                            {hook.events.map(e => <Badge key={e} variant="secondary" className="text-xs">{e}</Badge>)}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            <Switch
                                                checked={hook.active}
                                                onCheckedChange={() => toggleWebhook(hook)}
                                            />
                                            <span className={`text-xs ${hook.active ? 'text-green-500' : 'text-muted-foreground'}`}>
                                                {hook.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-xs">
                                        {hook.lastTriggered ? new Date(hook.lastTriggered).toLocaleString() : "Never"}
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
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
