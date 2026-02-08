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
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";
import { Loader2, Save, RefreshCw, AlertTriangle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface AuditConfig {
    enabled: boolean;
    output_path?: string;
    log_arguments: boolean;
    log_results: boolean;
    storage_type: number | string; // Enum: 0=UNSPECIFIED, 1=FILE, 2=SQLITE, 3=POSTGRES, 4=WEBHOOK
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
}

interface GlobalSettings {
    audit?: AuditConfig;
    [key: string]: any;
}

const STORAGE_TYPE_WEBHOOK = 4;
const STORAGE_TYPE_WEBHOOK_STR = "STORAGE_TYPE_WEBHOOK";
const STORAGE_TYPE_UNSPECIFIED = 0;

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
    const [globalSettings, setGlobalSettings] = useState<GlobalSettings | null>(null);
    const [overwritingStorage, setOverwritingStorage] = useState<string | null>(null);

    const loadData = async () => {
        setLoading(true);
        try {
            // Load Alerts Webhook
            try {
                const alertRes = await apiClient.getWebhookURL();
                setAlertWebhookUrl(alertRes.url || "");
            } catch (e) {
                console.warn("Failed to fetch alert webhook", e);
            }

            // Load Global Settings (for Audit)
            const settings = await apiClient.getGlobalSettings();
            setGlobalSettings(settings);

            if (settings?.audit) {
                setAuditWebhookUrl(settings.audit.webhook_url || "");
                const type = settings.audit.storage_type;
                const isWebhook = type === STORAGE_TYPE_WEBHOOK || type === STORAGE_TYPE_WEBHOOK_STR;
                setAuditEnabled(isWebhook);

                // Detect if another storage type is active
                if (!isWebhook && type && type !== 0 && type !== "STORAGE_TYPE_UNSPECIFIED" && type !== 1 && type !== "STORAGE_TYPE_FILE") {
                    setOverwritingStorage(String(type));
                } else {
                    setOverwritingStorage(null);
                }
            }
        } catch (e) {
            console.error("Failed to load settings", e);
            toast({
                title: "Error",
                description: "Failed to load webhook settings.",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadData();
    }, []);

    const handleSaveAlerts = async () => {
        setSavingAlerts(true);
        try {
            await apiClient.saveWebhookURL(alertWebhookUrl);
            toast({
                title: "Saved",
                description: "Alert webhook URL updated successfully."
            });
        } catch (e) {
            console.error("Failed to save alert webhook", e);
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
        if (!globalSettings) return;
        setSavingAudit(true);
        try {
            // Deep clone to avoid state mutation
            const updatedSettings = structuredClone(globalSettings);
            if (!updatedSettings.audit) {
                updatedSettings.audit = {
                    enabled: false,
                    log_arguments: false,
                    log_results: false,
                    storage_type: STORAGE_TYPE_UNSPECIFIED
                };
            }

            updatedSettings.audit.webhook_url = auditWebhookUrl;

            if (auditEnabled) {
                updatedSettings.audit.storage_type = STORAGE_TYPE_WEBHOOK;
                // We ensure audit is enabled globally if using webhook
                updatedSettings.audit.enabled = true;
            } else {
                // Only unset if it was set to WEBHOOK
                const currentType = updatedSettings.audit.storage_type;
                if (currentType === STORAGE_TYPE_WEBHOOK || currentType === STORAGE_TYPE_WEBHOOK_STR) {
                    updatedSettings.audit.storage_type = STORAGE_TYPE_UNSPECIFIED;
                }
            }

            await apiClient.saveGlobalSettings(updatedSettings);
            // Reload to confirm server state
            await loadData();

            toast({
                title: "Saved",
                description: "Audit webhook settings updated successfully."
            });
        } catch (e) {
            console.error("Failed to save audit settings", e);
            toast({
                title: "Error",
                description: "Failed to save audit settings.",
                variant: "destructive"
            });
        } finally {
            setSavingAudit(false);
        }
    };

    if (loading && !globalSettings) {
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
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
                    <p className="text-muted-foreground">Configure outbound webhooks for system events.</p>
                </div>
                <Button variant="outline" size="sm" onClick={loadData} disabled={loading}>
                    <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                    Refresh
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Alert Webhook Card */}
                <Card>
                    <CardHeader>
                        <CardTitle>System Alerts</CardTitle>
                        <CardDescription>
                            Receive notifications when system health issues or errors occur.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="alert-url">Webhook URL</Label>
                            <Input
                                id="alert-url"
                                placeholder="https://api.pagerduty.com/..."
                                value={alertWebhookUrl}
                                onChange={(e) => setAlertWebhookUrl(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                We will send a JSON payload with error details to this URL.
                            </p>
                        </div>
                    </CardContent>
                    <CardFooter>
                        <Button onClick={handleSaveAlerts} disabled={savingAlerts}>
                            {savingAlerts && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            <Save className="mr-2 h-4 w-4" />
                            Save Configuration
                        </Button>
                    </CardFooter>
                </Card>

                {/* Audit Webhook Card */}
                <Card>
                    <CardHeader>
                        <CardTitle>Audit Log Stream</CardTitle>
                        <CardDescription>
                            Stream all system audit logs to an external collector in real-time.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="audit-enabled" className="flex flex-col space-y-1">
                                <span>Enable Webhook Stream</span>
                                <span className="font-normal text-xs text-muted-foreground">Sets audit storage type to Webhook</span>
                            </Label>
                            <Switch
                                id="audit-enabled"
                                checked={auditEnabled}
                                onCheckedChange={setAuditEnabled}
                            />
                        </div>

                        {overwritingStorage && !auditEnabled && (
                            <Alert variant="warning" className="py-2">
                                <AlertTriangle className="h-4 w-4" />
                                <AlertTitle className="text-xs">Warning</AlertTitle>
                                <AlertDescription className="text-xs">
                                    Another storage type is currently active ({overwritingStorage}). Enabling webhook will replace it.
                                </AlertDescription>
                            </Alert>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="audit-url">Collector URL</Label>
                            <Input
                                id="audit-url"
                                placeholder="https://splunk-hec.internal/..."
                                value={auditWebhookUrl}
                                onChange={(e) => setAuditWebhookUrl(e.target.value)}
                                disabled={!auditEnabled}
                            />
                        </div>
                    </CardContent>
                    <CardFooter>
                        <Button onClick={handleSaveAudit} disabled={savingAudit}>
                            {savingAudit && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            <Save className="mr-2 h-4 w-4" />
                            Save Configuration
                        </Button>
                    </CardFooter>
                </Card>
            </div>
        </div>
    );
}
