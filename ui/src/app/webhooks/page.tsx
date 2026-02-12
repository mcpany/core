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
import { Badge } from "@/components/ui/badge";
import { AlertTriangle, Activity, Save, RefreshCw, Loader2, Info } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Separator } from "@/components/ui/separator";

// Constants from backend proto
const STORAGE_TYPE_UNSPECIFIED = 0;
const STORAGE_TYPE_WEBHOOK = 4;

interface WebhookState {
    enabled: boolean;
    url: string;
}

// Partial interface for GlobalSettings based on proto
interface GlobalSettings {
    mcp_listen_address?: string;
    alerts?: {
        enabled?: boolean;
        webhook_url?: string;
    };
    audit?: {
        enabled?: boolean;
        webhook_url?: string;
        storage_type?: number;
    };
    [key: string]: any; // Allow other properties to be preserved
}

/**
 * WebhooksPage component.
 * Allows configuration of system-level webhooks for Alerts and Audit Logs.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const { toast } = useToast();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    // Local state for form management
    const [alertsConfig, setAlertsConfig] = useState<WebhookState>({ enabled: false, url: "" });
    const [auditConfig, setAuditConfig] = useState<WebhookState>({ enabled: false, url: "" });

    // Store full settings object to merge back on save
    const [fullSettings, setFullSettings] = useState<GlobalSettings>({});

    const fetchSettings = async () => {
        setLoading(true);
        try {
            const settings = await apiClient.getGlobalSettings();
            setFullSettings(settings);

            // Map backend settings to local state
            if (settings.alerts) {
                setAlertsConfig({
                    enabled: settings.alerts.enabled || false,
                    url: settings.alerts.webhook_url || ""
                });
            }

            if (settings.audit) {
                // Check if audit is enabled AND storage type is WEBHOOK
                const isWebhookStorage = settings.audit.storage_type === STORAGE_TYPE_WEBHOOK;
                setAuditConfig({
                    enabled: (settings.audit.enabled || false) && isWebhookStorage,
                    url: settings.audit.webhook_url || ""
                });
            }
        } catch (e) {
            console.error("Failed to fetch settings", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load webhook configuration."
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSettings();
    }, []);

    const handleSave = async () => {
        setSaving(true);
        try {
            // Determine new audit settings
            // If the user enables the "Audit Stream" switch:
            //   - Enable audit logging
            //   - Set storage type to WEBHOOK
            // If the user disables the "Audit Stream" switch:
            //   - If storage type was WEBHOOK, we disable audit logging entirely to be safe.
            //   - If storage type was NOT WEBHOOK (e.g. File), we leave it alone (preserve original state).

            let newAuditSettings = { ...fullSettings.audit };

            if (auditConfig.enabled) {
                // Explicitly enabling webhook stream
                newAuditSettings.enabled = true;
                newAuditSettings.storage_type = STORAGE_TYPE_WEBHOOK;
                newAuditSettings.webhook_url = auditConfig.url;
            } else {
                // Explicitly disabling webhook stream
                if (fullSettings.audit?.storage_type === STORAGE_TYPE_WEBHOOK) {
                    // It was webhook, so now we turn it off.
                    // Reverting to STORAGE_TYPE_UNSPECIFIED (0) or just disabling.
                    newAuditSettings.enabled = false;
                    newAuditSettings.storage_type = STORAGE_TYPE_UNSPECIFIED;
                }
                // If it wasn't webhook (e.g. File), we don't touch it, effectively keeping the "Audit Stream" toggle
                // strictly about the Webhook channel.
            }

            // Merge local state into full settings object
            const newSettings: GlobalSettings = {
                ...fullSettings,
                alerts: {
                    ...fullSettings.alerts,
                    enabled: alertsConfig.enabled,
                    webhook_url: alertsConfig.url
                },
                audit: newAuditSettings
            };

            await apiClient.saveGlobalSettings(newSettings);
            toast({
                title: "Configuration Saved",
                description: "Webhook settings have been updated."
            });
            // Refresh to ensure sync
            fetchSettings();
        } catch (e) {
            console.error("Failed to save settings", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save configuration."
            });
        } finally {
            setSaving(false);
        }
    };

    if (loading && !fullSettings.mcp_listen_address) {
        return (
            <div className="flex flex-col items-center justify-center h-[calc(100vh-4rem)]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                <p className="mt-2 text-muted-foreground">Loading configuration...</p>
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-6 p-8 pt-6 max-w-5xl mx-auto">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Integrations</h1>
                    <p className="text-muted-foreground mt-2">
                        Configure outbound webhooks for system alerts and audit logs.
                    </p>
                </div>
                <div className="flex gap-2">
                     <Button variant="outline" onClick={fetchSettings} disabled={loading || saving}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        Refresh
                    </Button>
                    <Button onClick={handleSave} disabled={loading || saving}>
                        {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        Save Changes
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Alerts Webhook Card */}
                <Card className="flex flex-col">
                    <CardHeader>
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <div className="p-2 bg-red-100 dark:bg-red-900/20 rounded-lg text-red-600">
                                    <AlertTriangle className="h-5 w-5" />
                                </div>
                                <CardTitle>System Alerts</CardTitle>
                            </div>
                            <Badge variant={alertsConfig.enabled ? "default" : "secondary"}>
                                {alertsConfig.enabled ? "Active" : "Inactive"}
                            </Badge>
                        </div>
                        <CardDescription className="mt-2">
                            Receive notifications when system health status changes (e.g. Service Down).
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4 flex-1">
                        <div className="flex items-center justify-between space-x-2 border p-3 rounded-md bg-muted/20">
                            <Label htmlFor="alerts-enabled" className="flex flex-col space-y-1 cursor-pointer">
                                <span>Enable Notifications</span>
                                <span className="font-normal text-xs text-muted-foreground">Send POST requests on status change.</span>
                            </Label>
                            <Switch
                                id="alerts-enabled"
                                checked={alertsConfig.enabled}
                                onCheckedChange={(checked) => setAlertsConfig(prev => ({ ...prev, enabled: checked }))}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="alerts-url">Webhook URL</Label>
                            <Input
                                id="alerts-url"
                                type="url"
                                placeholder="https://api.pagerduty.com/..."
                                value={alertsConfig.url}
                                onChange={(e) => setAlertsConfig(prev => ({ ...prev, url: e.target.value }))}
                                disabled={!alertsConfig.enabled}
                            />
                            <p className="text-[10px] text-muted-foreground">
                                Requires a valid HTTPS URL. Payload includes `status`, `message`, and `timestamp`.
                            </p>
                        </div>
                    </CardContent>
                </Card>

                {/* Audit Logs Webhook Card */}
                <Card className="flex flex-col">
                     <CardHeader>
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <div className="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg text-blue-600">
                                    <Activity className="h-5 w-5" />
                                </div>
                                <CardTitle>Audit Stream</CardTitle>
                            </div>
                             <Badge variant={auditConfig.enabled ? "default" : "secondary"}>
                                {auditConfig.enabled ? "Streaming" : "Inactive"}
                            </Badge>
                        </div>
                        <CardDescription className="mt-2">
                            Stream all operational audit logs to an external collector in real-time.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4 flex-1">
                         <div className="flex items-center justify-between space-x-2 border p-3 rounded-md bg-muted/20">
                            <Label htmlFor="audit-enabled" className="flex flex-col space-y-1 cursor-pointer">
                                <span>Enable Streaming</span>
                                <span className="font-normal text-xs text-muted-foreground">Stream logs via HTTP POST.</span>
                            </Label>
                            <Switch
                                id="audit-enabled"
                                checked={auditConfig.enabled}
                                onCheckedChange={(checked) => setAuditConfig(prev => ({ ...prev, enabled: checked }))}
                            />
                        </div>

                         {auditConfig.enabled && (
                            <div className="flex items-start gap-2 p-2 bg-amber-50 dark:bg-amber-900/20 text-amber-800 dark:text-amber-200 text-xs rounded border border-amber-200 dark:border-amber-800">
                                <Info className="h-4 w-4 shrink-0 mt-0.5" />
                                <p>
                                    Enabling webhook streaming will override any existing File or Database audit logging configuration.
                                </p>
                            </div>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="audit-url">Collector URL</Label>
                            <Input
                                id="audit-url"
                                type="url"
                                placeholder="https://splunk-hec.example.com/..."
                                value={auditConfig.url}
                                onChange={(e) => setAuditConfig(prev => ({ ...prev, url: e.target.value }))}
                                disabled={!auditConfig.enabled}
                            />
                            <p className="text-[10px] text-muted-foreground">
                                High-volume stream. Ensure the endpoint can handle burst traffic.
                            </p>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <Separator className="my-6" />

            <div className="bg-muted/30 p-4 rounded-lg border">
                <h3 className="font-medium text-sm mb-2">Payload Specifications</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 text-xs text-muted-foreground">
                    <div>
                        <strong className="text-foreground">System Alerts Payload:</strong>
                        <pre className="mt-2 p-2 bg-background border rounded overflow-x-auto">
{`{
  "type": "alert",
  "status": "degraded",
  "message": "Service 'payments' is unreachable",
  "timestamp": "2024-03-20T10:00:00Z"
}`}
                        </pre>
                    </div>
                    <div>
                         <strong className="text-foreground">Audit Log Payload:</strong>
                        <pre className="mt-2 p-2 bg-background border rounded overflow-x-auto">
{`{
  "type": "audit_log",
  "event_id": "evt_123",
  "actor": "user_456",
  "action": "execute_tool",
  "resource": "calculator",
  "timestamp": "2024-03-20T10:00:00Z"
}`}
                        </pre>
                    </div>
                </div>
            </div>
        </div>
    );
}
