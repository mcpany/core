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
import { useToast } from "@/hooks/use-toast";
import { Save, Bell, FileText, Loader2 } from "lucide-react";

/**
 * WebhooksPage component.
 * Allows configuring global alert webhooks and audit log streaming webhooks.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const { toast } = useToast();
    const [loading, setLoading] = useState(true);
    const [savingAlerts, setSavingAlerts] = useState(false);
    const [savingAudit, setSavingAudit] = useState(false);

    // Alert Webhook State
    const [alertWebhookUrl, setAlertWebhookUrl] = useState("");

    // Audit Webhook State
    const [auditWebhookUrl, setAuditWebhookUrl] = useState("");
    const [auditEnabled, setAuditEnabled] = useState(false);
    const [auditConfig, setAuditConfig] = useState<any>(null); // Keep full audit config to preserve other fields

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            try {
                const [webhookRes, settingsRes] = await Promise.all([
                    apiClient.getWebhookURL(),
                    apiClient.getGlobalSettings()
                ]);

                if (webhookRes && webhookRes.url) {
                    setAlertWebhookUrl(webhookRes.url);
                }

                if (settingsRes && settingsRes.audit) {
                    setAuditConfig(settingsRes.audit);
                    setAuditWebhookUrl(settingsRes.audit.webhook_url || "");
                    // Check if enabled AND storage type includes WEBHOOK (enum 4)
                    // Note: API might return string or number for enum.
                    // Ideally we check storage_type === 'STORAGE_TYPE_WEBHOOK' or 4.
                    const storageType = settingsRes.audit.storage_type;
                    const isWebhook = storageType === 4 || storageType === 'STORAGE_TYPE_WEBHOOK';
                    setAuditEnabled(settingsRes.audit.enabled && isWebhook);
                }
            } catch (error) {
                console.error("Failed to load webhook configuration", error);
                toast({
                    title: "Error",
                    description: "Failed to load configuration from server.",
                    variant: "destructive"
                });
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [toast]);

    const handleSaveAlerts = async () => {
        setSavingAlerts(true);
        try {
            await apiClient.saveWebhookURL(alertWebhookUrl);
            toast({
                title: "Saved",
                description: "Alert webhook URL has been updated."
            });
        } catch (error) {
            console.error("Failed to save alert webhook", error);
            toast({
                title: "Error",
                description: "Failed to save alert webhook.",
                variant: "destructive"
            });
        } finally {
            setSavingAlerts(false);
        }
    };

    const handleSaveAudit = async () => {
        setSavingAudit(true);
        try {
            // Get current global settings first to ensure we don't overwrite other things?
            // Actually getGlobalSettings was called on mount.
            // But we should probably fetch fresh or just patch 'audit'.
            // saveGlobalSettings takes the WHOLE settings object usually?
            // Let's fetch fresh to be safe.
            const currentSettings = await apiClient.getGlobalSettings();
            const currentAudit = currentSettings.audit || {};

            const newAuditConfig = {
                ...currentAudit,
                webhook_url: auditWebhookUrl,
                enabled: auditEnabled,
                storage_type: auditEnabled ? 4 : (currentAudit.storage_type === 4 ? 0 : currentAudit.storage_type),
            };

            // If we disable the toggle, and it was enabled, we should probably disable audit or change type.
            // Let's assume this switch controls "Enable Audit Logging to Webhook".
            // If disabled, we don't touch global enabled state if it's logging to file?
            // But for this UI, let's simplify: Toggle = Enable Audit via Webhook.

            if (auditEnabled) {
                newAuditConfig.enabled = true;
            }

            await apiClient.saveGlobalSettings({
                ...currentSettings,
                audit: newAuditConfig
            });

            toast({
                title: "Saved",
                description: "Audit webhook configuration updated."
            });
        } catch (error) {
            console.error("Failed to save audit webhook", error);
            toast({
                title: "Error",
                description: "Failed to save audit configuration.",
                variant: "destructive"
            });
        } finally {
            setSavingAudit(false);
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
        <div className="flex-1 space-y-6 p-8 pt-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Integrations & Webhooks</h1>
                <p className="text-muted-foreground">Configure external integrations and event streams.</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Alert Webhooks */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Bell className="h-5 w-5" />
                            Alert Notifications
                        </CardTitle>
                        <CardDescription>
                            Send system alerts and health notifications to an external URL (e.g., Slack, PagerDuty).
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="alert-url">Webhook URL</Label>
                            <Input
                                id="alert-url"
                                placeholder="https://api.example.com/alerts"
                                value={alertWebhookUrl}
                                onChange={(e) => setAlertWebhookUrl(e.target.value)}
                                data-testid="alert-webhook-input"
                            />
                        </div>
                    </CardContent>
                    <CardFooter className="justify-end">
                        <Button onClick={handleSaveAlerts} disabled={savingAlerts} data-testid="alert-webhook-save">
                            {savingAlerts && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            <Save className="mr-2 h-4 w-4" />
                            Save Configuration
                        </Button>
                    </CardFooter>
                </Card>

                {/* Audit Webhooks */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <FileText className="h-5 w-5" />
                            Audit Log Streaming
                        </CardTitle>
                        <CardDescription>
                            Stream all system audit logs to an external HTTP endpoint in real-time.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="audit-enable" className="flex flex-col space-y-1">
                                <span>Enable Audit Streaming</span>
                                <span className="font-normal text-xs text-muted-foreground">Automatically sends JSON logs on every action.</span>
                            </Label>
                            <Switch
                                id="audit-enable"
                                checked={auditEnabled}
                                onCheckedChange={setAuditEnabled}
                                data-testid="audit-webhook-switch"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="audit-url">Destination URL</Label>
                            <Input
                                id="audit-url"
                                placeholder="https://logs.example.com/ingest"
                                value={auditWebhookUrl}
                                onChange={(e) => setAuditWebhookUrl(e.target.value)}
                                disabled={!auditEnabled}
                                data-testid="audit-webhook-input"
                            />
                        </div>
                    </CardContent>
                    <CardFooter className="justify-end">
                        <Button onClick={handleSaveAudit} disabled={savingAudit} data-testid="audit-webhook-save">
                            {savingAudit && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            <Save className="mr-2 h-4 w-4" />
                            Save Settings
                        </Button>
                    </CardFooter>
                </Card>
            </div>
        </div>
    );
}
