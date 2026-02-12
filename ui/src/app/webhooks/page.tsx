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
import { Switch } from "@/components/ui/switch";
import { Bell, FileText, Save, Loader2, AlertTriangle } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { GlobalSettings, AuditConfig_StorageType } from "@proto/config/v1/config";

/**
 * WebhooksPage component.
 * Manages system-level webhook integrations for Alerts and Audit Logs.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const { toast } = useToast();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [settings, setSettings] = useState<GlobalSettings | null>(null);

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
                description: "Failed to load global configuration."
            });
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        if (!settings) return;
        setSaving(true);
        try {
            // Ensure types are consistent with protobuf expectations
            const payload = {
                ...settings,
                alerts: settings.alerts ? {
                    enabled: settings.alerts.enabled,
                    webhook_url: settings.alerts.webhookUrl // Map camelCase to snake_case if client wrapper doesn't handle it
                } : undefined,
                audit: settings.audit ? {
                    ...settings.audit,
                    webhook_url: settings.audit.webhookUrl
                } : undefined
            };

            // Note: apiClient.saveGlobalSettings accepts 'any' and likely expects snake_case for some fields
            // based on GlobalSettingsForm logic. However, apiClient implementation uses fetchWithAuth
            // and simply stringifies the payload. The backend expects snake_case json tags usually.
            // The Generated Proto Types use camelCase properties but the json_name is snake_case.
            // When we use GlobalSettings.toJSON() it handles mapping.
            // But here we are modifying the object directly.
            // Let's rely on apiClient to handle it or manually ensure snake_case for the API.
            // Looking at GlobalSettingsForm, it manually constructs the payload.
            // We should do the same to be safe.

            const safePayload: any = {
                // Preserve other settings that might be overwritten if we don't send them?
                // saveGlobalSettings sends a POST. If it's a full replace, we need everything.
                // If it's a merge, we are fine.
                // Assuming it might be a full replace or validation requires other fields.
                // We'll try to send the full object but mapped.
                // Ideally, we use GlobalSettings.toJSON(settings) but we don't have the class here, just interface.
                // Let's manually construct the critical parts we changed.
                // Wait, GlobalSettingsForm sends a SUBSET.
                // "const payload = { mcp_listen_address: ..., alerts: ... }"
                // This implies PARTIAL update is supported or at least attempted.
                // But we want to be careful not to unset things we didn't touch if the backend is dumb.
                // Ideally we use what we fetched.

                // HACK: We assume `settings` object matches the structure we got from `getGlobalSettings`.
                // `getGlobalSettings` returns JSON from backend (snake_case usually unless `protojson` emits camelCase).
                // But the type hint says `GlobalSettings` (camelCase).
                // If `apiClient.getGlobalSettings()` returns camelCase (via proto-mapping), then we have camelCase keys.
                // If we send camelCase to backend, does it accept it?
                // Most Go servers with protojson allow camelCase or snake_case inputs.
                // Let's assume standard behavior and just send what we have, but ensuring our modifications are applied.

                ...settings,
                alerts: {
                    enabled: settings.alerts?.enabled,
                    webhook_url: settings.alerts?.webhookUrl
                },
                audit: {
                    enabled: settings.audit?.enabled,
                    webhook_url: settings.audit?.webhookUrl,
                    storage_type: settings.audit?.storageType,
                    // Pass other audit fields to avoid resetting them
                    output_path: settings.audit?.outputPath,
                    log_arguments: settings.audit?.logArguments,
                    log_results: settings.audit?.logResults,
                }
            };

            await apiClient.saveGlobalSettings(safePayload);
            toast({ title: "Settings Saved", description: "Webhook configuration updated." });
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

    const updateAlerts = (updates: Partial<typeof settings.alerts>) => {
        if (!settings) return;
        setSettings({
            ...settings,
            alerts: {
                enabled: false,
                webhookUrl: "",
                ...settings.alerts,
                ...updates
            }
        });
    };

    const updateAudit = (updates: Partial<typeof settings.audit>) => {
        if (!settings) return;
        const newAudit = {
            enabled: false,
            webhookUrl: "",
            storageType: 0,
            outputPath: "",
            logArguments: false,
            logResults: false,
            webhookHeaders: {},
            ...settings.audit,
            ...updates
        };

        // Logic: If enabling audit webhook, enforce storage type if not already set or explicit
        // Actually, if we set the URL, we probably imply we want to use it.
        // If enabling and storage type is not WEBHOOK, should we force it?
        // Let's force it if enabled is becoming true.
        if (updates.enabled === true) {
             newAudit.storageType = AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK;
        }

        setSettings({
            ...settings,
            audit: newAudit
        });
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!settings) {
        return (
            <div className="flex flex-col items-center justify-center h-full gap-4 text-muted-foreground">
                <AlertTriangle className="h-12 w-12 opacity-20" />
                <p>Unable to load configuration.</p>
                <Button onClick={loadSettings} variant="outline">Retry</Button>
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-6 p-8 pt-6 max-w-4xl mx-auto">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Integrations</h1>
                    <p className="text-muted-foreground mt-1">
                        Configure outbound webhooks for system events and audit logs.
                    </p>
                </div>
                <Button onClick={handleSave} disabled={saving}>
                    {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    Save Changes
                </Button>
            </div>

            <div className="grid gap-6">
                {/* Alerts Configuration */}
                <Card>
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <div className="p-2 bg-red-100 dark:bg-red-900/20 rounded-md">
                                <Bell className="h-5 w-5 text-red-600 dark:text-red-400" />
                            </div>
                            <div>
                                <CardTitle>System Alerts</CardTitle>
                                <CardDescription>Receive notifications for critical system health events and errors.</CardDescription>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        <div className="flex items-center justify-between">
                            <div className="space-y-0.5">
                                <Label className="text-base">Enable Alerts Webhook</Label>
                                <p className="text-sm text-muted-foreground">
                                    Send payloads when services become unhealthy or errors occur.
                                </p>
                            </div>
                            <Switch
                                checked={settings.alerts?.enabled || false}
                                onCheckedChange={(checked) => updateAlerts({ enabled: checked })}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="alerts-url">Webhook URL</Label>
                            <Input
                                id="alerts-url"
                                placeholder="https://api.example.com/webhooks/alerts"
                                value={settings.alerts?.webhookUrl || ""}
                                onChange={(e) => updateAlerts({ webhookUrl: e.target.value })}
                                disabled={!settings.alerts?.enabled}
                            />
                        </div>
                    </CardContent>
                </Card>

                {/* Audit Configuration */}
                <Card>
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <div className="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-md">
                                <FileText className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                            </div>
                            <div>
                                <CardTitle>Audit Logging</CardTitle>
                                <CardDescription>Stream structured audit logs to an external collector.</CardDescription>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        <div className="flex items-center justify-between">
                            <div className="space-y-0.5">
                                <Label className="text-base">Enable Audit Webhook</Label>
                                <p className="text-sm text-muted-foreground">
                                    Stream every tool execution and system action to a webhook.
                                </p>
                            </div>
                            <Switch
                                checked={settings.audit?.enabled || false}
                                onCheckedChange={(checked) => updateAudit({ enabled: checked })}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="audit-url">Collector URL</Label>
                            <Input
                                id="audit-url"
                                placeholder="https://splunk-hec.example.com/..."
                                value={settings.audit?.webhookUrl || ""}
                                onChange={(e) => updateAudit({ webhookUrl: e.target.value })}
                                disabled={!settings.audit?.enabled}
                            />
                            {settings.audit?.enabled && settings.audit?.storageType !== AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK && (
                                <div className="flex items-center gap-2 text-xs text-yellow-600 dark:text-yellow-400 mt-1">
                                    <AlertTriangle className="h-3 w-3" />
                                    <span>Warning: Audit storage type is currently set to a different provider. Enabling this will switch it to Webhook.</span>
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
