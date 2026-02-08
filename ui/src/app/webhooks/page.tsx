/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Bell, FileText, Save, Loader2, AlertCircle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";
import { GlobalSettings, AuditConfig_StorageType } from "@proto/config/v1/config";

/**
 * WebhooksPage component.
 * Allows configuration of system-wide webhooks for Alerts and Audit logs.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [settings, setSettings] = useState<GlobalSettings | null>(null);
    const { toast } = useToast();

    // Local state for form inputs
    const [alertUrl, setAlertUrl] = useState("");
    const [alertEnabled, setAlertEnabled] = useState(false);
    const [auditUrl, setAuditUrl] = useState("");
    const [auditEnabled, setAuditEnabled] = useState(false);

    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        setLoading(true);
        try {
            const data = await apiClient.getGlobalSettings();
            // Data from API is plain JSON (snake_case potentially if raw, but apiClient usually returns json).
            // apiClient.getGlobalSettings returns `res.json()`.
            // The proto-generated `fromJSON` handles the mapping from wire format (snake_case) to domain object (camelCase).
            const typedSettings = GlobalSettings.fromJSON(data);
            setSettings(typedSettings);

            // Initialize local state
            if (typedSettings.alerts) {
                setAlertUrl(typedSettings.alerts.webhookUrl || "");
                setAlertEnabled(typedSettings.alerts.enabled || false);
            }
            if (typedSettings.audit) {
                setAuditUrl(typedSettings.audit.webhookUrl || "");
                // Audit enabled + storage_type WEBHOOK implies webhook enabled
                setAuditEnabled(typedSettings.audit.enabled && typedSettings.audit.storageType === AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK);
            }
        } catch (e) {
            console.error("Failed to load settings", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load webhook settings."
            });
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        if (!settings) return;
        setSaving(true);

        try {
            // Clone settings to modify using fromPartial since settings is already typed (camelCase)
            const newSettings = GlobalSettings.fromPartial(settings);

            // Update Alerts
            if (!newSettings.alerts) newSettings.alerts = { enabled: false, webhookUrl: "" };
            newSettings.alerts.webhookUrl = alertUrl;
            newSettings.alerts.enabled = alertEnabled;

            // Update Audit
            if (!newSettings.audit) newSettings.audit = {
                enabled: false,
                webhookUrl: "",
                storageType: AuditConfig_StorageType.STORAGE_TYPE_UNSPECIFIED,
                outputPath: "",
                logArguments: false,
                logResults: false,
                webhookHeaders: {},
                splunk: undefined,
                datadog: undefined
            };
            newSettings.audit.webhookUrl = auditUrl;

            // If audit webhook is enabled, we must ensure storage_type is WEBHOOK
            if (auditEnabled) {
                newSettings.audit.enabled = true;
                newSettings.audit.storageType = AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK;
            } else {
                if (newSettings.audit.storageType === AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK) {
                    newSettings.audit.enabled = false;
                }
            }

            // toJSON converts back to wire format (snake_case)
            await apiClient.saveGlobalSettings(GlobalSettings.toJSON(newSettings));

            toast({
                title: "Settings Saved",
                description: "Webhook configuration has been updated."
            });

            // Reload to ensure sync
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

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-6 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Integrations</h1>
                    <p className="text-muted-foreground">Configure external integrations and webhooks.</p>
                </div>
                <Button onClick={handleSave} disabled={saving || loading}>
                    {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    {!saving && <Save className="mr-2 h-4 w-4" />}
                    Save Changes
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Alert Webhook */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Bell className="h-5 w-5 text-primary" />
                            Alerts Webhook
                        </CardTitle>
                        <CardDescription>
                            Receive system alerts and health notifications via HTTP POST.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="alert-enabled" className="flex flex-col gap-1">
                                <span>Enable Alerts</span>
                                <span className="font-normal text-xs text-muted-foreground">Send notifications when system health degrades.</span>
                            </Label>
                            <Switch
                                id="alert-enabled"
                                checked={alertEnabled}
                                onCheckedChange={setAlertEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="alert-url">Webhook URL</Label>
                            <Input
                                id="alert-url"
                                placeholder="https://api.slack.com/webhook/..."
                                value={alertUrl}
                                onChange={(e) => setAlertUrl(e.target.value)}
                                disabled={!alertEnabled}
                            />
                        </div>
                    </CardContent>
                    <CardFooter className="bg-muted/10 text-xs text-muted-foreground p-4">
                        <AlertCircle className="h-3 w-3 mr-2" />
                        Payload includes timestamp, severity, and error details.
                    </CardFooter>
                </Card>

                {/* Audit Webhook */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <FileText className="h-5 w-5 text-primary" />
                            Audit Log Stream
                        </CardTitle>
                        <CardDescription>
                            Stream security and access logs to an external collector.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="audit-enabled" className="flex flex-col gap-1">
                                <span>Enable Stream</span>
                                <span className="font-normal text-xs text-muted-foreground">Forward audit events as they happen.</span>
                            </Label>
                            <Switch
                                id="audit-enabled"
                                checked={auditEnabled}
                                onCheckedChange={setAuditEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="audit-url">Collector URL</Label>
                            <Input
                                id="audit-url"
                                placeholder="https://collector.splunk.com/..."
                                value={auditUrl}
                                onChange={(e) => setAuditUrl(e.target.value)}
                                disabled={!auditEnabled}
                            />
                        </div>
                    </CardContent>
                    <CardFooter className="bg-muted/10 text-xs text-muted-foreground p-4">
                        <AlertCircle className="h-3 w-3 mr-2" />
                        Requires a high-throughput endpoint. Storage type will be set to WEBHOOK.
                    </CardFooter>
                </Card>
            </div>
        </div>
    );
}
