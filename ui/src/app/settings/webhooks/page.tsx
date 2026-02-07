/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import Link from "next/link";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiClient } from "@/lib/client";

/**
 * WebhooksPage component.
 * @returns The rendered component.
 */
export default function WebhooksPage() {
    const [webhookUrl, setWebhookUrl] = useState("");
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        async function fetchWebhook() {
            try {
                const data = await apiClient.getWebhookURL();
                setWebhookUrl(data.url || "");
            } catch (error) {
                console.error("Failed to fetch webhook", error);
            }
        }
        fetchWebhook();
    }, []);

    const handleSave = async () => {
        setLoading(true);
        try {
            await apiClient.saveWebhookURL(webhookUrl);
            toast({
                title: "Webhook updated",
                description: "Global alert webhook URL has been saved.",
            });
        } catch (error) {
            toast({
                title: "Error",
                description: "Failed to save webhook URL.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

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
                            <h3 className="text-lg font-medium">Global Alert Webhook</h3>
                            <p className="text-sm text-muted-foreground">Configure a global webhook to receive system alerts.</p>
                         </div>
                    </div>
                    <Card className="backdrop-blur-sm bg-background/50">
                        <CardHeader>
                            <CardTitle>Webhook Configuration</CardTitle>
                            <CardDescription>
                                Alerts will be POSTed to this URL.
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center gap-4">
                                <Input
                                    placeholder="https://example.com/webhook"
                                    value={webhookUrl}
                                    onChange={(e) => setWebhookUrl(e.target.value)}
                                />
                                <Button onClick={handleSave} disabled={loading}>
                                    {loading ? "Saving..." : "Save"}
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
