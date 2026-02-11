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
import { AlertTriangle, FileText, Loader2, Save } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * WebhooksPage component.
 * Manages system-level webhooks for Alerts and Audit Logs.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const { toast } = useToast();

    // Alert Config State
    const [alertsEnabled, setAlertsEnabled] = useState(false);
    const [alertsUrl, setAlertsUrl] = useState("");

    // Audit Config State
    const [auditEnabled, setAuditEnabled] = useState(false);
    const [auditUrl, setAuditUrl] = useState("");
    // We need to preserve other audit settings when saving
    const [auditConfig, setAuditConfig] = useState<any>({});

    // Full settings object to preserve other fields
    const [fullSettings, setFullSettings] = useState<any>({});

    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        setLoading(true);
        try {
            const settings = await apiClient.getGlobalSettings();
            setFullSettings(settings);

            // Alerts
            if (settings.alerts) {
                setAlertsEnabled(settings.alerts.enabled || false);
                setAlertsUrl(settings.alerts.webhook_url || "");
            }

            // Audit
            if (settings.audit) {
                setAuditConfig(settings.audit);
                // Check if audit is enabled AND storage type is webhook (4) OR just if webhook url is present?
                // Ideally, we want to enable the "Webhook Channel" specifically.
                // But audit.enabled toggles ALL audit logging.
                // So we'll track if audit is enabled generally.
                // And if storage_type is 4, we assume webhook is the primary output.
                setAuditEnabled(settings.audit.enabled || false);
                setAuditUrl(settings.audit.webhook_url || "");
            }
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
            // Construct payload
            const payload = {
                ...fullSettings,
                alerts: {
                    ...fullSettings.alerts,
                    enabled: alertsEnabled,
                    webhook_url: alertsUrl
                },
                audit: {
                    ...auditConfig,
                    enabled: auditEnabled, // Global audit switch
                    webhook_url: auditUrl,
                    // Auto-set storage type to WEBHOOK (4) if URL is provided and we are enabling it?
                    // Or only if the user explicitly wants "Webhook Storage".
                    // For this feature to be useful, we assume if they set a URL, they want to use it.
                    // If they clear the URL, we might revert to default?
                    // Let's be explicit: If URL is set, use storage_type 4.
                    storage_type: auditUrl ? 4 : (auditConfig.storage_type || 0)
                }
            };

            await apiClient.saveGlobalSettings(payload);
            toast({
                title: "Configuration Saved",
                description: "System webhooks have been updated."
            });

            // Reload to ensure state sync
            loadSettings();
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

    if (loading) {
        return (
            <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-6 p-8 pt-6 max-w-5xl mx-auto">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Webhooks</h1>
                    <p className="text-muted-foreground mt-2">
                        Configure outbound notification channels for critical system events.
                    </p>
                </div>
                <Button onClick={handleSave} disabled={saving}>
                    {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    Save Changes
                </Button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Alerts Webhook */}
                <Card className="flex flex-col">
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <div className="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-lg">
                                <AlertTriangle className="h-5 w-5 text-orange-600 dark:text-orange-400" />
                            </div>
                            <div>
                                <CardTitle>Health Alerts</CardTitle>
                                <CardDescription>Notify on system health changes.</CardDescription>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-4 flex-1">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="alerts-enabled" className="flex flex-col gap-1">
                                <span>Enable Alerts</span>
                                <span className="font-normal text-xs text-muted-foreground">Send POST requests when health status changes.</span>
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
                                placeholder="https://api.example.com/alerts"
                                value={alertsUrl}
                                onChange={(e) => setAlertsUrl(e.target.value)}
                                disabled={!alertsEnabled}
                            />
                        </div>
                    </CardContent>
                </Card>

                {/* Audit Webhook */}
                <Card className="flex flex-col">
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                                <FileText className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                            </div>
                            <div>
                                <CardTitle>Audit Logging</CardTitle>
                                <CardDescription>Stream audit events to an external collector.</CardDescription>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-4 flex-1">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="audit-enabled" className="flex flex-col gap-1">
                                <span>Enable Audit Stream</span>
                                <span className="font-normal text-xs text-muted-foreground">Send POST requests for every audited action.</span>
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
                                placeholder="https://api.example.com/audit"
                                value={auditUrl}
                                onChange={(e) => setAuditUrl(e.target.value)}
                                disabled={!auditEnabled}
                            />
                        </div>
                        <div className="text-xs text-muted-foreground bg-muted p-2 rounded">
                            Note: Enabling this will set the audit storage type to <code>WEBHOOK</code>.
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
