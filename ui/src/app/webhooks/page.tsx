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
import { Plus, Trash2, RefreshCw, Save } from "lucide-react";
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
    lastTriggered?: string;
    status?: "success" | "failure" | "pending";
}

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<WebhookConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [newUrl, setNewUrl] = useState("");
    const [newType, setNewType] = useState<"alerts" | "audit">("alerts");

    // We store the full settings to preserve other fields on save
    const [globalSettings, setGlobalSettings] = useState<any>(null);

    useEffect(() => {
        fetchWebhooks();
    }, []);

    const fetchWebhooks = async () => {
        setLoading(true);
        try {
            const settings = await apiClient.getGlobalSettings();
            setGlobalSettings(settings);

            const hooks: WebhookConfig[] = [];

            // Alerts Webhook
            if (settings.alerts?.webhookUrl) {
                hooks.push({
                    id: "alerts",
                    url: settings.alerts.webhookUrl,
                    events: ["alerts"],
                    active: settings.alerts.enabled !== false, // Default true if not set
                    status: "success" // Placeholder
                });
            } else if (settings.alerts) {
                 // Exist but no URL?
                 // UI logic: if URL is empty, we don't show it or show as empty?
                 // Let's show it if enabled?
            }

            // Audit Webhook
            if (settings.audit?.webhookUrl) {
                hooks.push({
                    id: "audit",
                    url: settings.audit.webhookUrl,
                    events: ["audit.logs"],
                    active: settings.audit.enabled !== false,
                    status: "success" // Placeholder
                });
            }

            setWebhooks(hooks);
        } catch (err) {
            console.error(err);
            toast.error("Failed to load webhooks");
        } finally {
            setLoading(false);
        }
    };

    const saveChanges = async (updatedWebhooks: WebhookConfig[]) => {
        setSaving(true);
        try {
            const newSettings = { ...globalSettings };

            // Update Alerts
            const alertHook = updatedWebhooks.find(h => h.id === "alerts");
            if (!newSettings.alerts) newSettings.alerts = {};
            if (alertHook) {
                newSettings.alerts.webhookUrl = alertHook.url;
                newSettings.alerts.enabled = alertHook.active;
            } else {
                newSettings.alerts.webhookUrl = "";
                newSettings.alerts.enabled = false;
            }

            // Update Audit
            const auditHook = updatedWebhooks.find(h => h.id === "audit");
            if (!newSettings.audit) newSettings.audit = {};
            if (auditHook) {
                newSettings.audit.webhookUrl = auditHook.url;
                newSettings.audit.enabled = auditHook.active;
            } else {
                newSettings.audit.webhookUrl = "";
                newSettings.audit.enabled = false;
            }

            await apiClient.saveGlobalSettings(newSettings);
            toast.success("Webhooks saved");
            fetchWebhooks();
        } catch (err) {
            console.error(err);
            toast.error("Failed to save webhooks");
        } finally {
            setSaving(false);
        }
    };

    const toggleWebhook = (id: string) => {
        const updated = webhooks.map(w => w.id === id ? { ...w, active: !w.active } : w);
        setWebhooks(updated);
        // We defer save to explicit button or auto-save?
        // Let's use explicit save for consistency with Middleware page
    };

    const deleteWebhook = (id: string) => {
        const updated = webhooks.filter(w => w.id !== id);
        setWebhooks(updated);
    };

    const addWebhook = () => {
        if (!newUrl) return;

        // Check if type already exists
        if (webhooks.find(w => w.id === newType)) {
            toast.error(`${newType} webhook already exists. Modify the existing one.`);
            return;
        }

        const newHook: WebhookConfig = {
            id: newType,
            url: newUrl,
            events: newType === "alerts" ? ["alerts"] : ["audit.logs"],
            active: true
        };

        setWebhooks([...webhooks, newHook]);
        setNewUrl("");
        setIsDialogOpen(false);
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
                    <p className="text-muted-foreground">Configure outbound webhooks for system events.</p>
                </div>
                <div className="flex gap-2">
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
                                    Select the event type and enter the URL.
                                </DialogDescription>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label className="text-right">Type</Label>
                                    <select
                                        className="col-span-3 flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                        value={newType}
                                        onChange={(e) => setNewType(e.target.value as "alerts" | "audit")}
                                    >
                                        <option value="alerts">Alerts</option>
                                        <option value="audit">Audit Logs</option>
                                    </select>
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
                    <Button variant="outline" onClick={fetchWebhooks} disabled={loading || saving}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} /> Refresh
                    </Button>
                    <Button onClick={() => saveChanges(webhooks)} disabled={loading || saving}>
                        <Save className="mr-2 h-4 w-4" /> Save Changes
                    </Button>
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
                            {webhooks.map((hook) => (
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
                                            <span className={`text-xs ${hook.active ? 'text-green-500' : 'text-muted-foreground'}`}>
                                                {hook.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-xs">
                                        {hook.lastTriggered || "Never"}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
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
