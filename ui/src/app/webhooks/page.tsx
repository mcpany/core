/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { AuditConfig_StorageType } from "@proto/config/v1/config";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";
import { Loader2, Save } from "lucide-react";

/**
 * WebhooksPage component.
 * Allows configuration of system-wide webhooks for alerts and audit logging.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const { toast } = useToast();
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    // Local state for form
    const [alertsEnabled, setAlertsEnabled] = useState(false);
    const [alertsUrl, setAlertsUrl] = useState("");

    const [auditEnabled, setAuditEnabled] = useState(false);
    const [auditUrl, setAuditUrl] = useState("");
    const [auditStorageType, setAuditStorageType] = useState<AuditConfig_StorageType>(0);

    // Full settings object to preserve other fields when saving
    const [fullSettings, setFullSettings] = useState<any>(null);

    useEffect(() => {
        loadSettings();
    }, []);

    async function loadSettings() {
        setLoading(true);
        try {
            const settings = await apiClient.getGlobalSettings();
            setFullSettings(settings);

            if (settings.alerts) {
                setAlertsEnabled(settings.alerts.enabled || false);
                setAlertsUrl(settings.alerts.webhookUrl || "");
            }

            if (settings.audit) {
                setAuditEnabled(settings.audit.enabled || false);
                setAuditUrl(settings.audit.webhookUrl || "");
                setAuditStorageType(settings.audit.storageType || 0);
            }
        } catch (e) {
            console.error("Failed to load settings", e);
            toast({ variant: "destructive", description: "Failed to load system settings" });
        } finally {
            setLoading(false);
        }
    }

    async function handleSave() {
        setSaving(true);
        try {
            // Construct payload merging with existing settings
            const payload = { ...fullSettings };

            // Update Alerts
            payload.alerts = {
                ...payload.alerts,
                enabled: alertsEnabled,
                webhookUrl: alertsUrl
            };

            // Update Audit
            // If enabling webhook, force storage type to WEBHOOK
            let newStorageType = auditStorageType;
            if (auditEnabled) {
                newStorageType = AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK;
            }

            payload.audit = {
                ...payload.audit,
                enabled: auditEnabled,
                webhookUrl: auditUrl,
                storageType: newStorageType
            };

            await apiClient.saveGlobalSettings(payload);

            // Reload to confirm sync
            await loadSettings();

            toast({ title: "Settings Saved", description: "System webhooks configuration updated." });
        } catch (e) {
            console.error("Failed to save settings", e);
            toast({ variant: "destructive", description: "Failed to save settings" });
        } finally {
            setSaving(false);
        }
    }

    if (loading && !fullSettings) {
        return <div className="flex justify-center p-8"><Loader2 className="h-8 w-8 animate-spin" /></div>;
    }

    return (
        <div className="flex-1 space-y-6 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Webhooks</h1>
                    <p className="text-muted-foreground">Configure outbound notifications for system events.</p>
                </div>
                <Button onClick={handleSave} disabled={saving}>
                    {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    Save Changes
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>System Alerts</CardTitle>
                        <CardDescription>
                            Receive notifications when system health status changes (e.g. Healthy to Degraded).
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="alerts-toggle" className="flex flex-col gap-1">
                                <span>Enable Alerts</span>
                                <span className="font-normal text-xs text-muted-foreground">Send POST request on status change</span>
                            </Label>
                            <Switch
                                id="alerts-toggle"
                                checked={alertsEnabled}
                                onCheckedChange={setAlertsEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="alerts-url">Webhook URL</Label>
                            <Input
                                id="alerts-url"
                                placeholder="https://..."
                                value={alertsUrl}
                                onChange={(e) => setAlertsUrl(e.target.value)}
                                disabled={!alertsEnabled}
                            />
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Audit Logging</CardTitle>
                        <CardDescription>
                            Stream all audit logs (tool executions, config changes) to an external endpoint.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="audit-toggle" className="flex flex-col gap-1">
                                <span>Enable Audit Stream</span>
                                <span className="font-normal text-xs text-muted-foreground">Send POST request for every audit event</span>
                            </Label>
                            <Switch
                                id="audit-toggle"
                                data-testid="audit-switch"
                                checked={auditEnabled}
                                onCheckedChange={setAuditEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="audit-url">Webhook URL</Label>
                            <Input
                                id="audit-url"
                                placeholder="https://audit-logs.example.com/..."
                                value={auditUrl}
                                onChange={(e) => setAuditUrl(e.target.value)}
                                disabled={!auditEnabled}
                            />
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
