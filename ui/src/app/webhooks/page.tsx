/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Bell, FileText, Save, Loader2, RefreshCw, AlertTriangle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * WebhooksPage component.
 * Allows configuration of system-wide webhooks for Alerts and Audit Logs.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const { toast } = useToast();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [settings, setSettings] = useState<any>(null);

    // Alert State
    const [alertsEnabled, setAlertsEnabled] = useState(false);
    const [alertsUrl, setAlertsUrl] = useState("");

    // Audit State
    const [auditEnabled, setAuditEnabled] = useState(false);
    const [auditUrl, setAuditUrl] = useState("");
    // We track storage type just to know if we should override it
    const [auditStorageType, setAuditStorageType] = useState<number>(0);

    const loadSettings = async () => {
        setLoading(true);
        try {
            const data = await apiClient.getGlobalSettings();
            setSettings(data);

            // Alerts
            if (data.alerts) {
                setAlertsEnabled(data.alerts.enabled || false);
                setAlertsUrl(data.alerts.webhook_url || "");
            }

            // Audit
            if (data.audit) {
                // Determine if audit is "enabled" in UI sense if storage type is webhook OR if general enabled flag is set
                // But specifically for this page, we care about the Webhook configuration.
                // If storage_type is 4 (WEBHOOK), we consider the webhook active.
                setAuditEnabled(data.audit.enabled && data.audit.storage_type === 4);
                setAuditUrl(data.audit.webhook_url || "");
                setAuditStorageType(data.audit.storage_type || 0);
            }
        } catch (e) {
            console.error("Failed to load settings", e);
            toast({
                title: "Error",
                description: "Failed to load global settings.",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadSettings();
    }, []);

    const handleSave = async () => {
        if (!settings) return;
        setSaving(true);
        try {
            // Prepare update payload
            // We need to be careful not to overwrite other settings if the backend does a full replace.
            // But usually saveGlobalSettings merges or we send the full object back.
            // Let's assume we send the full object back with modifications.
            const newSettings = { ...settings };

            // Update Alerts
            newSettings.alerts = {
                ...newSettings.alerts,
                enabled: alertsEnabled,
                webhook_url: alertsUrl
            };

            // Update Audit
            // If enabled via this UI, we enforce storage_type = 4 (WEBHOOK)
            if (auditEnabled) {
                newSettings.audit = {
                    ...newSettings.audit,
                    enabled: true,
                    webhook_url: auditUrl,
                    storage_type: 4 // STORAGE_TYPE_WEBHOOK
                };
            } else {
                // If disabling webhook, do we disable audit entirely?
                // Or just switch storage type?
                // Let's assume we disable audit entirely if this is the main control,
                // OR just switch storage type to FILE (1) if it was WEBHOOK (4).
                // But to be safe and simple: disable audit.
                // Users can re-enable File logging in Global Settings.
                if (auditStorageType === 4) {
                     newSettings.audit = {
                        ...newSettings.audit,
                        enabled: false,
                        storage_type: 0
                    };
                }
                // If it wasn't webhook, we don't touch it (this UI only manages Webhook aspect).
            }

            await apiClient.saveGlobalSettings(newSettings);
            toast({
                title: "Settings Saved",
                description: "Webhook configurations have been updated."
            });
            // Reload to confirm
            loadSettings();

        } catch (e) {
             console.error("Failed to save settings", e);
            toast({
                title: "Error",
                description: "Failed to save settings.",
                variant: "destructive"
            });
        } finally {
            setSaving(false);
        }
    };

    if (loading && !settings) {
        return (
            <div className="flex flex-col items-center justify-center h-[calc(100vh-4rem)]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                <p className="text-muted-foreground mt-2">Loading configuration...</p>
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] overflow-y-auto">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Integrations</h1>
                    <p className="text-muted-foreground">Configure outbound webhooks for system events and audit logs.</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" onClick={loadSettings} disabled={loading || saving}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        Refresh
                    </Button>
                    <Button onClick={handleSave} disabled={loading || saving}>
                        {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        Save Changes
                    </Button>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Alerts Webhook Card */}
                <Card className="flex flex-col">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Bell className="h-5 w-5 text-amber-500" />
                            System Alerts
                        </CardTitle>
                        <CardDescription>
                            Receive notifications when system health changes or critical errors occur.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4 flex-1">
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="alerts-enabled" className="flex flex-col space-y-1">
                                <span>Enable Alerts Webhook</span>
                                <span className="font-normal text-xs text-muted-foreground">
                                    Send JSON payloads for health status changes.
                                </span>
                            </Label>
                            <Switch
                                id="alerts-enabled"
                                checked={alertsEnabled}
                                onCheckedChange={setAlertsEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="alerts-url">Webhook URL</Label>
                            <Input
                                id="alerts-url"
                                placeholder="https://api.example.com/webhooks/alerts"
                                value={alertsUrl}
                                onChange={(e) => setAlertsUrl(e.target.value)}
                                disabled={!alertsEnabled}
                            />
                        </div>
                         <div className="rounded-md bg-muted p-3">
                            <div className="flex items-center gap-2 text-sm font-medium mb-1">
                                <AlertTriangle className="h-4 w-4 text-yellow-500" />
                                Example Payload
                            </div>
                            <pre className="text-[10px] font-mono text-muted-foreground overflow-x-auto">
{`{
  "event": "health_check_failed",
  "status": "degraded",
  "timestamp": "2024-03-20T10:00:00Z",
  "details": { "check": "database", "error": "connection timeout" }
}`}
                            </pre>
                        </div>
                    </CardContent>
                </Card>

                {/* Audit Logs Webhook Card */}
                <Card className="flex flex-col">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <FileText className="h-5 w-5 text-blue-500" />
                            Audit Logging
                        </CardTitle>
                        <CardDescription>
                            Stream all system audit logs to an external HTTP endpoint.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4 flex-1">
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="audit-enabled" className="flex flex-col space-y-1">
                                <span>Enable Audit Stream</span>
                                <span className="font-normal text-xs text-muted-foreground">
                                    Send detailed logs of every user action and tool execution.
                                </span>
                            </Label>
                            <Switch
                                id="audit-enabled"
                                checked={auditEnabled}
                                onCheckedChange={setAuditEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="audit-url">Webhook URL</Label>
                            <Input
                                id="audit-url"
                                placeholder="https://splunk-hec.example.com/services/collector"
                                value={auditUrl}
                                onChange={(e) => setAuditUrl(e.target.value)}
                                disabled={!auditEnabled}
                            />
                        </div>
                        <div className="rounded-md bg-muted p-3">
                            <div className="flex items-center gap-2 text-sm font-medium mb-1">
                                <FileText className="h-4 w-4 text-blue-500" />
                                Example Payload
                            </div>
                            <pre className="text-[10px] font-mono text-muted-foreground overflow-x-auto">
{`{
  "id": "audit-12345",
  "action": "execute_tool",
  "actor": "user@example.com",
  "target": "github-service",
  "meta": { "tool": "create_issue" }
}`}
                            </pre>
                        </div>
                    </CardContent>
                    <CardFooter>
                         <p className="text-xs text-muted-foreground">
                            Note: Enabling this will set the audit storage type to <strong>WEBHOOK</strong>.
                            Ensure the endpoint can handle high volume.
                        </p>
                    </CardFooter>
                </Card>
            </div>
        </div>
    );
}
