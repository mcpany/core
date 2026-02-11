/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Webhook, Save, Loader2, AlertTriangle, Info } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

/**
 * SystemIntegrationsPage component.
 * Replaces the mocked Webhooks page with real configuration for System Alerts and Audit Logging.
 * @returns The rendered component.
 */
export default function SystemIntegrationsPage() {
    const [settings, setSettings] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        setLoading(true);
        try {
            const data = await apiClient.getGlobalSettings();
            setSettings(data);
        } catch (e) {
            console.error("Failed to load settings", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load global settings."
            });
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            // Prepare payload
            // We need to send the full settings object or at least the parts we modified.
            // apiClient.saveGlobalSettings expects the full object structure usually.
            // But we only want to update alerts and audit.
            // Let's assume we send back the modified settings object.

            // Ensure audit storage type is set if webhook is enabled
            const newSettings = { ...settings };
            if (newSettings.audit?.enabled && newSettings.audit?.webhook_url) {
                // Set storage type to WEBHOOK (4) if not already set or if explicitly enabling webhook
                // We don't want to override if user has set it to something else but still wants webhook?
                // Wait, Proto says `storage_type` is an enum. It can only be one value.
                // So if we enable webhook here, we imply storage_type = WEBHOOK.
                newSettings.audit.storage_type = 4; // STORAGE_TYPE_WEBHOOK
            }

            await apiClient.saveGlobalSettings(newSettings);
            toast({ title: "Settings Saved", description: "System integration settings updated." });
            // Reload to confirm
            loadSettings();
        } catch (e) {
            console.error("Failed to save settings", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save settings."
            });
        } finally {
            setSaving(false);
        }
    };

    const updateAlerts = (updates: any) => {
        setSettings({
            ...settings,
            alerts: { ...settings.alerts, ...updates }
        });
    };

    const updateAudit = (updates: any) => {
        setSettings({
            ...settings,
            audit: { ...settings.audit, ...updates }
        });
    };

    if (loading && !settings) {
        return (
            <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 max-w-4xl mx-auto">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Integrations</h1>
                    <p className="text-muted-foreground">Configure outbound webhooks for alerts and auditing.</p>
                </div>
                <Button onClick={handleSave} disabled={saving}>
                    {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    Save Changes
                </Button>
            </div>

            <div className="grid gap-6">
                {/* System Alerts */}
                <Card>
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <Webhook className="h-5 w-5 text-primary" />
                            <CardTitle>System Alerts</CardTitle>
                        </div>
                        <CardDescription>
                            Receive notifications when critical system events occur (e.g., service downtime, high error rates).
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="alerts-enabled" className="flex flex-col space-y-1">
                                <span>Enable Alerts</span>
                                <span className="font-normal text-xs text-muted-foreground">Send notifications to an external URL.</span>
                            </Label>
                            <Switch
                                id="alerts-enabled"
                                checked={settings?.alerts?.enabled || false}
                                onCheckedChange={(checked) => updateAlerts({ enabled: checked })}
                            />
                        </div>
                        {settings?.alerts?.enabled && (
                            <div className="grid gap-2 animate-in fade-in slide-in-from-top-1 duration-200">
                                <Label htmlFor="alerts-url">Webhook URL</Label>
                                <Input
                                    id="alerts-url"
                                    placeholder="https://hooks.slack.com/services/..."
                                    value={settings?.alerts?.webhook_url || ""}
                                    onChange={(e) => updateAlerts({ webhook_url: e.target.value })}
                                />
                                <p className="text-[10px] text-muted-foreground">
                                    Supports generic JSON payloads compatible with Slack, Discord, and Teams.
                                </p>
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* Audit Logging */}
                <Card>
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <Info className="h-5 w-5 text-primary" />
                            <CardTitle>Audit Logging</CardTitle>
                        </div>
                        <CardDescription>
                            Stream security and operational logs to an external compliance system.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <Alert variant="default" className="bg-muted/50 border-primary/20">
                            <Info className="h-4 w-4" />
                            <AlertTitle>Audit Storage</AlertTitle>
                            <AlertDescription>
                                Enabling this webhook will set the Audit Storage Type to <strong>WEBHOOK</strong>.
                                Previous storage settings (e.g. File or Database) may be overridden.
                            </AlertDescription>
                        </Alert>

                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="audit-enabled" className="flex flex-col space-y-1">
                                <span>Enable Audit Stream</span>
                                <span className="font-normal text-xs text-muted-foreground">Send every request/response log to an endpoint.</span>
                            </Label>
                            <Switch
                                id="audit-enabled"
                                checked={settings?.audit?.enabled || false}
                                onCheckedChange={(checked) => updateAudit({ enabled: checked })}
                            />
                        </div>
                         {settings?.audit?.enabled && (
                            <div className="grid gap-2 animate-in fade-in slide-in-from-top-1 duration-200">
                                <Label htmlFor="audit-url">Webhook URL</Label>
                                <Input
                                    id="audit-url"
                                    placeholder="https://collector.example.com/v1/audit"
                                    value={settings?.audit?.webhook_url || ""}
                                    onChange={(e) => updateAudit({ webhook_url: e.target.value })}
                                />
                                <div className="flex items-center gap-2">
                                     <Label htmlFor="log-args" className="text-xs font-normal text-muted-foreground flex items-center gap-2">
                                        <Switch
                                            id="log-args"
                                            className="scale-75"
                                            checked={settings?.audit?.log_arguments || false}
                                            onCheckedChange={(c) => updateAudit({ log_arguments: c })}
                                        />
                                        Include Arguments (Sensitive)
                                     </Label>
                                     <Label htmlFor="log-results" className="text-xs font-normal text-muted-foreground flex items-center gap-2">
                                        <Switch
                                            id="log-results"
                                            className="scale-75"
                                            checked={settings?.audit?.log_results || false}
                                            onCheckedChange={(c) => updateAudit({ log_results: c })}
                                        />
                                        Include Results (Sensitive)
                                     </Label>
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
