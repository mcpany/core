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
import { Play, Trash2, Edit, Loader2 } from "lucide-react";
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
    id: string; // 'alerts' or 'audit'
    name: string;
    url: string;
    events: string[];
    active: boolean;
    status?: "success" | "failure" | "pending"; // Not strictly available from backend config, implied
}

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<WebhookConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingWebhook, setEditingWebhook] = useState<WebhookConfig | null>(null);
    const [editUrl, setEditUrl] = useState("");

    useEffect(() => {
        loadWebhooks();
    }, []);

    const loadWebhooks = async () => {
        try {
            const settings = await apiClient.getGlobalSettings();
            const hooks: WebhookConfig[] = [];

            // Alerts Webhook
            const alertUrl = settings.alerts?.webhookUrl;
            const alertEnabled = settings.alerts?.enabled !== false; // Default enabled? Proto defaults are usually false/empty.
            // Actually, proto bool defaults to false.
            // Let's assume if URL is present, it's configured. Enabled flag is explicit.
            if (alertUrl) {
                hooks.push({
                    id: 'alerts',
                    name: 'System Alerts',
                    url: alertUrl,
                    events: ['alert.triggered', 'health.check.failed'],
                    active: settings.alerts?.enabled || false,
                    status: 'success' // Placeholder
                });
            } else {
                // Add placeholder for creating one?
                // Or we always show rows for available slots (Alerts, Audit)
                hooks.push({
                    id: 'alerts',
                    name: 'System Alerts',
                    url: '',
                    events: ['alert.triggered'],
                    active: false
                });
            }

            // Audit Webhook
            const auditUrl = settings.audit?.webhookUrl;
            if (auditUrl) {
                hooks.push({
                    id: 'audit',
                    name: 'Audit Logs',
                    url: auditUrl,
                    events: ['audit.log.created'],
                    active: settings.audit?.enabled || false,
                    status: 'success'
                });
            } else {
                 hooks.push({
                    id: 'audit',
                    name: 'Audit Logs',
                    url: '',
                    events: ['audit.log.created'],
                    active: false
                });
            }

            setWebhooks(hooks);
        } catch (err) {
            console.error(err);
            toast.error("Failed to load settings");
        } finally {
            setLoading(false);
        }
    };

    const saveSettings = async (newWebhooks: WebhookConfig[]) => {
        try {
            const current = await apiClient.getGlobalSettings();

            // Map webhooks back to settings
            const alertsHook = newWebhooks.find(w => w.id === 'alerts');
            const auditHook = newWebhooks.find(w => w.id === 'audit');

            const newSettings = {
                ...current,
                alerts: {
                    ...current.alerts,
                    enabled: alertsHook?.active || false,
                    webhookUrl: alertsHook?.url || ''
                },
                audit: {
                    ...current.audit,
                    enabled: auditHook?.active || false,
                    webhookUrl: auditHook?.url || ''
                }
            };

            await apiClient.saveGlobalSettings(newSettings);
            toast.success("Webhooks updated");
            loadWebhooks(); // Reload to confirm
        } catch (err) {
            console.error(err);
            toast.error("Failed to save changes");
        }
    }

    const toggleWebhook = (id: string) => {
        const updated = webhooks.map(w => w.id === id ? { ...w, active: !w.active } : w);
        setWebhooks(updated); // Optimistic
        saveSettings(updated);
    };

    const openEdit = (hook: WebhookConfig) => {
        setEditingWebhook(hook);
        setEditUrl(hook.url);
        setIsDialogOpen(true);
    };

    const saveEdit = () => {
        if (!editingWebhook) return;
        const updated = webhooks.map(w => w.id === editingWebhook.id ? { ...w, url: editUrl, active: !!editUrl } : w);
        setWebhooks(updated);
        saveSettings(updated);
        setIsDialogOpen(false);
    };

    const deleteWebhook = (id: string) => {
        // Just clear URL and disable
        const updated = webhooks.map(w => w.id === id ? { ...w, url: '', active: false } : w);
        setWebhooks(updated);
        saveSettings(updated);
    };

    const testWebhook = (id: string) => {
        // TODO: Implement test endpoint in backend?
        alert(`Test for ${id} not implemented yet`);
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
            </div>

            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle>Configured Webhooks</CardTitle>
                </CardHeader>
                <CardContent>
                     <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Type</TableHead>
                                <TableHead>URL</TableHead>
                                <TableHead>Events</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {webhooks.map((hook) => (
                                <TableRow key={hook.id}>
                                    <TableCell className="font-medium">{hook.name}</TableCell>
                                    <TableCell className="font-mono text-xs max-w-[200px] truncate" title={hook.url}>
                                        {hook.url || <span className="text-muted-foreground italic">Not configured</span>}
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
                                                disabled={!hook.url}
                                            />
                                            <span className={`text-xs ${hook.active ? 'text-green-500' : 'text-muted-foreground'}`}>
                                                {hook.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button variant="ghost" size="icon" onClick={() => openEdit(hook)} title="Edit URL">
                                                <Edit className="h-4 w-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" onClick={() => testWebhook(hook.id)} title="Test Delivery" disabled={!hook.url}>
                                                <Play className="h-4 w-4" />
                                            </Button>
                                             <Button variant="ghost" size="icon" onClick={() => deleteWebhook(hook.id)} className="text-red-500 hover:text-red-600" disabled={!hook.url}>
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

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Configure Webhook</DialogTitle>
                        <DialogDescription>
                            Enter the URL where events should be sent for {editingWebhook?.name}.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor="url" className="text-right">
                                Payload URL
                            </Label>
                            <Input
                                id="url"
                                value={editUrl}
                                onChange={(e) => setEditUrl(e.target.value)}
                                placeholder="https://..."
                                className="col-span-3"
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button onClick={saveEdit}>Save</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
