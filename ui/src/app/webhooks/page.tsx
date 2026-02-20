/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Webhook, Plus, Play, Trash2, Edit } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter
} from "@/components/ui/dialog"
import { apiClient } from "@/lib/client";
import { SystemWebhook } from "@proto/config/v1/webhook";
import { useToast } from "@/hooks/use-toast";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<SystemWebhook[]>([]);
    const [loading, setLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    // Form State
    const [currentWebhook, setCurrentWebhook] = useState<Partial<SystemWebhook>>({});
    const [selectedEvents, setSelectedEvents] = useState<string[]>(["all"]);

    const { toast } = useToast();

    const fetchWebhooks = async () => {
        setLoading(true);
        try {
            const list = await apiClient.listSystemWebhooks();
            setWebhooks(list);
        } catch (e) {
            console.error("Failed to load webhooks", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load webhooks."
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchWebhooks();
    }, []);

    const handleSave = async () => {
        if (!currentWebhook.url) {
            toast({
                variant: "destructive",
                title: "Error",
                description: "URL is required."
            });
            return;
        }

        try {
            const webhook: SystemWebhook = {
                id: currentWebhook.id || "",
                url: currentWebhook.url,
                secret: currentWebhook.secret || "",
                events: selectedEvents,
                active: currentWebhook.active ?? true,
                createdAt: currentWebhook.createdAt || "",
                lastTriggeredAt: currentWebhook.lastTriggeredAt || "",
                lastStatus: currentWebhook.lastStatus || "",
                lastError: currentWebhook.lastError || ""
            };

            if (webhook.id) {
                await apiClient.updateSystemWebhook(webhook);
                toast({ title: "Webhook Updated", description: "Configuration saved." });
            } else {
                await apiClient.createSystemWebhook(webhook);
                toast({ title: "Webhook Created", description: "New webhook registered." });
            }
            setIsDialogOpen(false);
            fetchWebhooks();
        } catch (e) {
            console.error("Failed to save webhook", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save webhook."
            });
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure?")) return;
        try {
            await apiClient.deleteSystemWebhook(id);
            toast({ title: "Webhook Deleted", description: "Webhook removed." });
            fetchWebhooks();
        } catch (e) {
            console.error("Failed to delete webhook", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to delete webhook."
            });
        }
    };

    const handleTest = async (id: string) => {
        try {
            toast({ title: "Testing...", description: "Sending test payload..." });
            const res = await apiClient.testSystemWebhook(id);
            if (res.success) {
                toast({ title: "Success", description: res.message });
            } else {
                toast({ variant: "destructive", title: "Failed", description: res.message });
            }
            fetchWebhooks(); // Reload to update status
        } catch (e) {
            console.error("Failed to test webhook", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to test webhook."
            });
        }
    };

    const openNew = () => {
        setCurrentWebhook({ active: true });
        setSelectedEvents(["all"]);
        setIsDialogOpen(true);
    };

    const openEdit = (wh: SystemWebhook) => {
        setCurrentWebhook(wh);
        setSelectedEvents(wh.events);
        setIsDialogOpen(true);
    };

    const toggleActive = async (wh: SystemWebhook) => {
        try {
            const updated = { ...wh, active: !wh.active };
            await apiClient.updateSystemWebhook(updated);
            fetchWebhooks();
        } catch (e) {
            console.error("Failed to toggle webhook", e);
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
                    <p className="text-muted-foreground">Configure outbound webhooks for system events.</p>
                </div>
                <Button onClick={openNew}>
                    <Plus className="mr-2 h-4 w-4" /> New Webhook
                </Button>
            </div>

            <Card className="backdrop-blur-sm bg-background/50 flex-1 flex flex-col min-h-0">
                <CardHeader>
                    <CardTitle>Configured Webhooks</CardTitle>
                </CardHeader>
                <CardContent className="flex-1 overflow-auto">
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
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24">Loading...</TableCell>
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
                                                    onCheckedChange={() => toggleActive(hook)}
                                                />
                                                <div className="flex flex-col">
                                                    <span className={`text-xs ${hook.active ? 'text-foreground' : 'text-muted-foreground'}`}>
                                                        {hook.active ? "Active" : "Inactive"}
                                                    </span>
                                                    {hook.lastStatus && (
                                                        <span className={`text-[10px] uppercase ${hook.lastStatus === 'success' ? 'text-green-500' : 'text-red-500'}`}>
                                                            {hook.lastStatus}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-muted-foreground text-xs">
                                            <div>{hook.lastTriggeredAt ? new Date(hook.lastTriggeredAt).toLocaleString() : "Never"}</div>
                                            {hook.lastError && (
                                                <div className="text-red-500 truncate max-w-[200px]" title={hook.lastError}>{hook.lastError}</div>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex justify-end gap-2">
                                                <Button variant="ghost" size="icon" onClick={() => handleTest(hook.id)} title="Test Delivery">
                                                    <Play className="h-4 w-4" />
                                                </Button>
                                                <Button variant="ghost" size="icon" onClick={() => openEdit(hook)}>
                                                    <Edit className="h-4 w-4" />
                                                </Button>
                                                 <Button variant="ghost" size="icon" onClick={() => handleDelete(hook.id)} className="text-red-500 hover:text-red-600">
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

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{currentWebhook.id ? "Edit Webhook" : "New Webhook"}</DialogTitle>
                        <DialogDescription>
                            Enter the URL where events should be sent.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="url">Payload URL</Label>
                            <Input
                                id="url"
                                value={currentWebhook.url || ""}
                                onChange={(e) => setCurrentWebhook({...currentWebhook, url: e.target.value})}
                                placeholder="https://api.example.com/webhook"
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="secret">Secret (Optional)</Label>
                            <Input
                                id="secret"
                                type="password"
                                value={currentWebhook.secret || ""}
                                onChange={(e) => setCurrentWebhook({...currentWebhook, secret: e.target.value})}
                                placeholder="HMAC Secret Key"
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label>Events</Label>
                            {/* Simple multi-select via text for now or just check-boxes */}
                            <div className="flex gap-2">
                                <Badge
                                    variant={selectedEvents.includes("all") ? "default" : "outline"}
                                    className="cursor-pointer"
                                    onClick={() => setSelectedEvents(["all"])}
                                >
                                    All Events
                                </Badge>
                                <Badge
                                    variant={selectedEvents.includes("service.registered") ? "default" : "outline"}
                                    className="cursor-pointer"
                                    onClick={() => {
                                        const newEvents = selectedEvents.filter(e => e !== "all");
                                        if (newEvents.includes("service.registered")) {
                                            setSelectedEvents(newEvents.filter(e => e !== "service.registered"));
                                        } else {
                                            setSelectedEvents([...newEvents, "service.registered"]);
                                        }
                                    }}
                                >
                                    Service Registered
                                </Badge>
                                 <Badge
                                    variant={selectedEvents.includes("tool.invoked") ? "default" : "outline"}
                                    className="cursor-pointer"
                                    onClick={() => {
                                        const newEvents = selectedEvents.filter(e => e !== "all");
                                        if (newEvents.includes("tool.invoked")) {
                                            setSelectedEvents(newEvents.filter(e => e !== "tool.invoked"));
                                        } else {
                                            setSelectedEvents([...newEvents, "tool.invoked"]);
                                        }
                                    }}
                                >
                                    Tool Invoked
                                </Badge>
                            </div>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleSave}>Save Webhook</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
