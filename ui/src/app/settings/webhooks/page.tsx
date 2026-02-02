/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";
import Link from "next/link";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhookUrl, setWebhookUrl] = useState("");
    const [enabled, setEnabled] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [fetchError, setFetchError] = useState(false);
    const { toast } = useToast();
    // Keep full settings to merge updates
    const [fullSettings, setFullSettings] = useState<any>({});

    useEffect(() => {
        async function fetchSettings() {
            try {
                const settings = await apiClient.getGlobalSettings();
                setFullSettings(settings);
                if (settings.alerts) {
                    // Handle snake_case from backend (UseProtoNames: true)
                    setWebhookUrl(settings.alerts.webhook_url || settings.alerts.webhookUrl || "");
                    setEnabled(settings.alerts.enabled || false);
                }
                setFetchError(false);
            } catch (err) {
                console.error(err);
                setFetchError(true);
                toast({
                    title: "Error",
                    description: "Failed to load settings. Saving is disabled to prevent data loss.",
                    variant: "destructive"
                });
            } finally {
                setLoading(false);
            }
        }
        fetchSettings();
    }, [toast]);

    const handleSave = async () => {
        if (fetchError) return;
        setSaving(true);
        try {
            // Construct update payload preserving other settings
            const updatedSettings = {
                ...fullSettings,
                alerts: {
                    ...(fullSettings.alerts || {}),
                    webhook_url: webhookUrl,
                    enabled: enabled
                }
            };

            await apiClient.saveGlobalSettings(updatedSettings);
            setFullSettings(updatedSettings);
            toast({
                title: "Settings Saved",
                description: "Global webhook configuration updated."
            });
        } catch (err) {
            console.error(err);
            toast({
                title: "Error",
                description: "Failed to save settings",
                variant: "destructive"
            });
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="flex-1 p-8 pt-6 h-[calc(100vh-4rem)] flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
            </div>

            <Tabs defaultValue="webhooks" className="space-y-4 flex-1 flex flex-col">
                <TabsList>
                    <TabsTrigger value="profiles" asChild>
                        <Link href="/settings">Profiles</Link>
                    </TabsTrigger>
                    <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
                    <TabsTrigger value="secrets" asChild>
                        <Link href="/settings/secrets">Secrets & Keys</Link>
                    </TabsTrigger>
                    <TabsTrigger value="auth" asChild>
                        <Link href="/settings/auth">Authentication</Link>
                    </TabsTrigger>
                    <TabsTrigger value="general" asChild>
                        <Link href="/settings/general">General</Link>
                    </TabsTrigger>
                </TabsList>
                <TabsContent value="webhooks" className="space-y-4">
                    <div className="flex items-center justify-between">
                         <div>
                            <h3 className="text-lg font-medium">Global Webhook</h3>
                            <p className="text-sm text-muted-foreground">Configure a global webhook to receive system health alerts.</p>
                         </div>
                    </div>
                    <Card className="backdrop-blur-sm bg-background/50">
                        <CardHeader>
                            <CardTitle>Health Alerts</CardTitle>
                            <CardDescription>
                                Receive POST requests when service health status changes (e.g. Healthy to Degraded).
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center space-x-2">
                                <Switch id="enabled" checked={enabled} onCheckedChange={setEnabled} disabled={fetchError} />
                                <Label htmlFor="enabled">Enable Webhook Notifications</Label>
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="url">Webhook URL</Label>
                                <Input
                                    id="url"
                                    value={webhookUrl}
                                    onChange={e => setWebhookUrl(e.target.value)}
                                    placeholder="https://your-monitoring-system.com/webhook"
                                    disabled={!enabled || fetchError}
                                />
                            </div>
                            <div className="flex justify-end">
                                <Button onClick={handleSave} disabled={saving || fetchError}>
                                    {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                    Save Changes
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
