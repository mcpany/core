/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
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
    const { toast } = useToast();

    // Fetch Global Settings
    useEffect(() => {
        const fetchSettings = async () => {
            try {
                const settings = await apiClient.getGlobalSettings();
                if (settings.alerts) {
                    setWebhookUrl(settings.alerts.webhook_url || "");
                    setEnabled(settings.alerts.enabled || false);
                }
            } catch (err) {
                toast({ title: "Error", description: "Failed to load settings", variant: "destructive" });
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, [toast]);

    const handleSave = async () => {
        setSaving(true);
        try {
            // Re-fetch to get latest
            const currentSettings = await apiClient.getGlobalSettings();

            await apiClient.saveGlobalSettings({
                ...currentSettings,
                alerts: {
                    ...currentSettings.alerts,
                    webhook_url: webhookUrl,
                    enabled: enabled
                }
            });
            toast({ title: "Settings Saved", description: "Webhook settings updated." });
        } catch (err) {
            console.error(err);
            toast({ title: "Error", description: "Failed to save settings", variant: "destructive" });
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return <div className="flex items-center justify-center h-full"><Loader2 className="animate-spin" /></div>;
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
                        <Link href="/settings">Secrets & Keys</Link>
                    </TabsTrigger>
                    <TabsTrigger value="auth" asChild>
                        <Link href="/settings">Authentication</Link>
                    </TabsTrigger>
                    <TabsTrigger value="general" asChild>
                        <Link href="/settings">General</Link>
                    </TabsTrigger>
                </TabsList>
                <TabsContent value="webhooks" className="space-y-4">
                    <div className="flex items-center justify-between">
                         <div>
                            <h3 className="text-lg font-medium">Global Webhook</h3>
                            <p className="text-sm text-muted-foreground">Receive notifications when system health status changes.</p>
                         </div>
                    </div>
                    <Card>
                        <CardHeader>
                            <CardTitle>Configuration</CardTitle>
                            <CardDescription>Configure where alerts should be sent.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center space-x-2">
                                <Switch id="enabled" checked={enabled} onCheckedChange={setEnabled} />
                                <Label htmlFor="enabled">Enable Webhook Notifications</Label>
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="url">Webhook URL</Label>
                                <Input
                                    id="url"
                                    placeholder="https://api.example.com/webhooks/mcp-alerts"
                                    value={webhookUrl}
                                    onChange={(e) => setWebhookUrl(e.target.value)}
                                    disabled={!enabled}
                                />
                            </div>
                            <div className="flex justify-end">
                                <Button onClick={handleSave} disabled={saving}>
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
