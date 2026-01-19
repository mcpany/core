/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";

interface Webhook {
    id: string;
    url: string;
    events: string[];
}

import Link from "next/link";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState<Webhook[]>([]);

    useEffect(() => {
        async function fetchWebhooks() {
            const res = await fetch("/api/settings/webhooks");
            if (res.ok) {
                setWebhooks(await res.json());
            }
        }
        fetchWebhooks();
    }, []);

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
                            <h3 className="text-lg font-medium">Webhooks</h3>
                            <p className="text-sm text-muted-foreground">Manage your webhook subscriptions.</p>
                         </div>
                        <Button>Add Webhook</Button>
                    </div>
                    <Card className="backdrop-blur-sm bg-background/50">
                        <CardContent className="p-0">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>ID</TableHead>
                                        <TableHead>URL</TableHead>
                                        <TableHead>Events</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {webhooks.map((hook) => (
                                        <TableRow key={hook.id}>
                                            <TableCell>{hook.id}</TableCell>
                                            <TableCell>{hook.url}</TableCell>
                                            <TableCell>{hook.events.join(", ")}</TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
