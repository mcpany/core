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
import { apiClient } from "@/lib/client";
import { toast } from "sonner";

interface WebhookConfig {
    id: string;
    url: string;
    events: string[];
    active: boolean;
    type: 'alerts' | 'audit';
}

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<WebhookConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [newUrl, setNewUrl] = useState("");
    const [newType, setNewType] = useState<'alerts' | 'audit'>('alerts');
    const [settings, setSettings] = useState<any>(null);

    useEffect(() => {
        loadWebhooks();
    }, []);

    const loadWebhooks = async () => {
        try {
            const data = await apiClient.getGlobalSettings();
            setSettings(data);
            const hooks: WebhookConfig[] = [];
            if (data.alerts?.webhookUrl) {
                hooks.push({
                    id: 'alerts',
                    url: data.alerts.webhookUrl,
                    events: ['alerts'],
                    active: data.alerts.enabled !== false,
                    type: 'alerts'
                });
            }
            if (data.audit?.webhookUrl) {
                hooks.push({
                    id: 'audit',
                    url: data.audit.webhookUrl,
                    events: ['audit.log'],
                    active: data.audit.enabled !== false,
                    type: 'audit'
                });
            }
            setWebhooks(hooks);
        } catch (err) {
            console.error("Failed to load settings", err);
            toast.error("Failed to load webhook configuration");
        } finally {
            setLoading(false);
        }
    };

    const toggleWebhook = async (id: string) => {
        const hook = webhooks.find(h => h.id === id);
        if (!hook || !settings) return;

        const newActive = !hook.active;
        const newSettings = { ...settings };

        if (hook.type === 'alerts') {
            if (!newSettings.alerts) newSettings.alerts = {};
            newSettings.alerts.enabled = newActive;
        } else if (hook.type === 'audit') {
            if (!newSettings.audit) newSettings.audit = {};
            newSettings.audit.enabled = newActive;
        }

        try {
            await apiClient.saveGlobalSettings(newSettings);
            // Optimistic update
            setWebhooks(webhooks.map(w => w.id === id ? { ...w, active: newActive } : w));
            setSettings(newSettings);
            toast.success(`${hook.type} webhook ${newActive ? 'enabled' : 'disabled'}`);
        } catch (err) {
            console.error("Failed to save settings", err);
            toast.error("Failed to update webhook status");
        }
    };

    const deleteWebhook = async (id: string) => {
        const hook = webhooks.find(h => h.id === id);
        if (!hook || !settings) return;

        const newSettings = { ...settings };
        if (hook.type === 'alerts') {
            if (newSettings.alerts) {
                newSettings.alerts.webhookUrl = "";
                newSettings.alerts.enabled = false;
            }
        } else if (hook.type === 'audit') {
            if (newSettings.audit) {
                newSettings.audit.webhookUrl = "";
                // Don't disable audit entirely, just clear webhook?
                // Audit config has storage_type. If it was webhook, we might need to change it?
                // For now just clear URL.
            }
        }

        try {
            await apiClient.saveGlobalSettings(newSettings);
            setWebhooks(webhooks.filter(w => w.id !== id));
            setSettings(newSettings);
            toast.success("Webhook removed");
        } catch (err) {
            console.error("Failed to delete webhook", err);
            toast.error("Failed to delete webhook");
        }
    };

    const addWebhook = async () => {
        if (!newUrl || !settings) return;

        const newSettings = { ...settings };
        // Check if already exists
        if (newType === 'alerts') {
            if (!newSettings.alerts) newSettings.alerts = {};
            newSettings.alerts.webhookUrl = newUrl;
            newSettings.alerts.enabled = true;
        } else if (newType === 'audit') {
            if (!newSettings.audit) newSettings.audit = {};
            newSettings.audit.webhookUrl = newUrl;
            // newSettings.audit.storageType = "webhook"; // Ideally set storage type
        }

        try {
            await apiClient.saveGlobalSettings(newSettings);
            setNewUrl("");
            setIsDialogOpen(false);
            loadWebhooks(); // Reload to reflect changes
            toast.success("Webhook added");
        } catch (err) {
            console.error("Failed to add webhook", err);
            toast.error("Failed to add webhook");
        }
    }

    const testWebhook = (id: string) => {
        // Trigger test via API if available, or just toast
        toast.info(`Test event sent to ${id} (simulated)`);
        // Actual test API call could go here
    }

    if (loading) {
        return (
            <div className="flex h-screen items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
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
                                <Label htmlFor="type" className="text-right">Type</Label>
                                <div className="col-span-3 flex gap-2">
                                    <Button
                                        variant={newType === 'alerts' ? 'default' : 'outline'}
                                        size="sm"
                                        onClick={() => setNewType('alerts')}
                                    >
                                        Alerts
                                    </Button>
                                    <Button
                                        variant={newType === 'audit' ? 'default' : 'outline'}
                                        size="sm"
                                        onClick={() => setNewType('audit')}
                                    >
                                        Audit Logs
                                    </Button>
                                </div>
                            </div>
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
                                <TableHead>Type</TableHead>
                                <TableHead>Events</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {webhooks.map((hook) => (
                                <TableRow key={hook.id}>
                                    <TableCell className="font-mono text-xs max-w-[200px] truncate" title={hook.url}>{hook.url}</TableCell>
                                    <TableCell>
                                        <Badge variant="outline">{hook.type}</Badge>
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
                                                onCheckedChange={() => toggleWebhook(hook.id)}
                                            />
                                            <span className={`text-xs ${hook.active ? 'text-green-500' : 'text-muted-foreground'}`}>
                                                {hook.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
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
