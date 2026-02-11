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
import { useToast } from "@/hooks/use-toast";
import { Webhook } from "lucide-react";

/**
 * WebhooksPage component.
 * Manages the global webhook configuration for system events.
 */
export default function WebhooksPage() {
    const [url, setUrl] = useState("");
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        loadWebhook();
    }, []);

    const loadWebhook = async () => {
        setIsLoading(true);
        try {
            const data = await apiClient.getWebhookURL();
            setUrl(data.url || "");
        } catch (error) {
            console.error("Failed to load webhook", error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load webhook configuration."
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleSave = async () => {
        setIsSaving(true);
        try {
            await apiClient.saveWebhookURL(url);
            toast({
                title: "Success",
                description: "Webhook URL updated successfully."
            });
        } catch (error) {
            console.error("Failed to save webhook", error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save webhook configuration."
            });
        } finally {
            setIsSaving(false);
        }
    };

    if (isLoading) {
        return <div className="flex-1 p-8 pt-6">Loading configuration...</div>;
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
                    <p className="text-muted-foreground">Configure the global outbound webhook for system events.</p>
                </div>
            </div>

            <Card className="backdrop-blur-sm bg-background/50 max-w-2xl">
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <Webhook className="h-5 w-5 text-primary" />
                        <CardTitle>Global Webhook Configuration</CardTitle>
                    </div>
                    <CardDescription>
                        Specify a URL to receive POST requests for all system alerts and events.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="webhook-url">Payload URL</Label>
                        <Input
                            id="webhook-url"
                            placeholder="https://example.com/webhook"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">
                            Events will be sent to this URL as JSON payloads.
                        </p>
                    </div>
                    <div className="flex justify-end">
                        <Button onClick={handleSave} disabled={isSaving}>
                            {isSaving ? "Saving..." : "Save Configuration"}
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
