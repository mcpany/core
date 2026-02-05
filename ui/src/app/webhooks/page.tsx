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
import { Webhook, Zap, Save } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [url, setUrl] = useState("");
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [testing, setTesting] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        loadWebhook();
    }, []);

    const loadWebhook = async () => {
        setLoading(true);
        try {
            const data = await apiClient.getWebhookURL();
            setUrl(data.url);
        } catch (e) {
            console.error("Failed to load webhook", e);
            toast({
                title: "Error",
                description: "Failed to load webhook configuration.",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await apiClient.saveWebhookURL(url);
            toast({
                title: "Configuration Saved",
                description: "Global webhook URL has been updated."
            });
        } catch (e) {
            console.error("Failed to save webhook", e);
            toast({
                title: "Save Failed",
                description: "Could not save webhook configuration.",
                variant: "destructive"
            });
        } finally {
            setSaving(false);
        }
    };

    const handleTest = async () => {
        if (!url) {
            toast({
                title: "Missing URL",
                description: "Please enter and save a webhook URL first.",
                variant: "destructive"
            });
            return;
        }

        setTesting(true);
        try {
            // First ensure it's saved
            await apiClient.saveWebhookURL(url);

            // Then trigger a test alert
            await apiClient.createAlert({
                title: "Test Alert",
                message: "This is a test alert from the Global Notification Center.",
                severity: "info",
                status: "active",
                service: "system",
                source: "notification-center",
                timestamp: new Date().toISOString()
            });

            toast({
                title: "Test Alert Sent",
                description: "A test alert has been created. Check your webhook endpoint."
            });
        } catch (e) {
            console.error("Failed to send test alert", e);
            toast({
                title: "Test Failed",
                description: "Could not send test alert.",
                variant: "destructive"
            });
        } finally {
            setTesting(false);
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
                    <p className="text-muted-foreground">Configure system-wide alert delivery.</p>
                </div>
            </div>

            <Card className="max-w-2xl backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <div className="p-2 bg-primary/10 rounded-md">
                            <Webhook className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <CardTitle>Global Webhook</CardTitle>
                            <CardDescription>
                                Receive notifications for all system alerts and incidents.
                            </CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="url">Payload URL</Label>
                        <Input
                            id="url"
                            placeholder="https://hooks.slack.com/services/..."
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                            disabled={loading}
                        />
                        <p className="text-[0.8rem] text-muted-foreground">
                            We will send a POST request with the alert JSON payload to this URL.
                        </p>
                    </div>
                </CardContent>
                <CardFooter className="flex justify-between border-t p-6 bg-muted/5">
                    <Button variant="outline" onClick={handleTest} disabled={loading || testing || !url}>
                        {testing ? <Zap className="mr-2 h-4 w-4 animate-spin" /> : <Zap className="mr-2 h-4 w-4" />}
                        Test Delivery
                    </Button>
                    <Button onClick={handleSave} disabled={loading || saving}>
                        {saving ? <Save className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        Save Changes
                    </Button>
                </CardFooter>
            </Card>
        </div>
    );
}
