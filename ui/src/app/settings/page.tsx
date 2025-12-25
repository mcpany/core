/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

export default function SettingsPage() {
    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                 <Card>
                    <CardHeader>
                        <CardTitle>Profiles</CardTitle>
                        <CardDescription>Manage execution profiles</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-muted-foreground mb-4">
                            Define different configurations for Development, Production, etc.
                        </p>
                        <Button variant="outline" className="w-full">Manage Profiles</Button>
                    </CardContent>
                </Card>
                 <Card>
                    <CardHeader>
                        <CardTitle>Webhooks</CardTitle>
                        <CardDescription>Configure and test webhooks</CardDescription>
                    </CardHeader>
                    <CardContent>
                         <p className="text-sm text-muted-foreground mb-4">
                            Set up event listeners and notification endpoints.
                        </p>
                        <Button variant="outline" className="w-full" asChild>
                             <a href="/settings/webhooks">Manage Webhooks</a>
                        </Button>
                    </CardContent>
                </Card>
                 <Card>
                    <CardHeader>
                        <CardTitle>Middleware</CardTitle>
                        <CardDescription>Visual management pipeline</CardDescription>
                    </CardHeader>
                    <CardContent>
                         <p className="text-sm text-muted-foreground mb-4">
                            Configure middleware for request processing.
                        </p>
                        <Button variant="outline" className="w-full" asChild>
                             <a href="/settings/middleware">Manage Middleware</a>
                        </Button>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}
