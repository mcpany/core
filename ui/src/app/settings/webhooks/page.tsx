
"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, Webhook, Activity, Trash2 } from "lucide-react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";

interface WebhookConfig {
    id: string;
    url: string;
    events: string[];
    active: boolean;
}

export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<WebhookConfig[]>([]);

    useEffect(() => {
        fetch("/api/webhooks")
            .then(res => res.json())
            .then(setWebhooks);
    }, []);

    const toggleWebhook = async (id: string) => {
        const webhook = webhooks.find(w => w.id === id);
        if (!webhook) return;

        setWebhooks(webhooks.map(w => w.id === id ? { ...w, active: !w.active } : w));

        try {
            await fetch("/api/webhooks", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ id, active: !webhook.active })
            });
        } catch (e) {
            console.error("Failed to toggle webhook", e);
            setWebhooks(webhooks); // Revert
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
                <Button><Plus className="mr-2 h-4 w-4" /> Add Webhook</Button>
            </div>

            <Card className="backdrop-blur-sm bg-background/50 border-muted/20">
                <CardHeader>
                    <CardTitle>Configured Webhooks</CardTitle>
                    <CardDescription>Manage outbound webhooks for system events.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>URL</TableHead>
                                <TableHead>Events</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {webhooks.map((hook) => (
                                <TableRow key={hook.id}>
                                    <TableCell className="font-mono text-xs">{hook.url}</TableCell>
                                    <TableCell>
                                        <div className="flex gap-1 flex-wrap">
                                            {hook.events.map(event => (
                                                <Badge key={event} variant="secondary" className="text-xs">{event}</Badge>
                                            ))}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center space-x-2">
                                            <Switch
                                                checked={hook.active}
                                                onCheckedChange={() => toggleWebhook(hook.id)}
                                            />
                                            <span className="text-sm text-muted-foreground min-w-[60px]">
                                                {hook.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">
                                         <Button variant="ghost" size="icon">
                                            <Trash2 className="h-4 w-4 text-red-500" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <Card className="mt-8 border-muted/20">
                <CardHeader>
                     <CardTitle className="flex items-center"><Activity className="mr-2 h-5 w-5" /> Recent Deliveries</CardTitle>
                     <CardDescription>Log of recent webhook attempts.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="text-sm text-muted-foreground text-center py-8">
                        No recent deliveries found.
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
