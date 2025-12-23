/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";

export default function SettingsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
      </div>
      <Tabs defaultValue="profiles" className="space-y-4">
        <TabsList>
          <TabsTrigger value="profiles">Profiles</TabsTrigger>
          <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
          <TabsTrigger value="middleware">Middleware</TabsTrigger>
        </TabsList>
        <TabsContent value="profiles" className="space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle>Execution Profiles</CardTitle>
                    <CardDescription>
                        Manage environments like Development, Production, and Debug.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between border p-4 rounded-lg">
                        <div>
                            <p className="font-medium">Development</p>
                            <p className="text-sm text-muted-foreground">Verbose logging, auto-reload.</p>
                        </div>
                        <Button variant="outline">Edit</Button>
                    </div>
                    <div className="flex items-center justify-between border p-4 rounded-lg">
                        <div>
                            <p className="font-medium">Production</p>
                            <p className="text-sm text-muted-foreground">Optimized for performance, strict security.</p>
                        </div>
                        <Button variant="outline">Edit</Button>
                    </div>
                </CardContent>
            </Card>
        </TabsContent>
        <TabsContent value="webhooks" className="space-y-4">
             <Card>
                <CardHeader>
                    <CardTitle>Webhooks</CardTitle>
                    <CardDescription>
                        Configure and test webhooks for event notifications.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid w-full max-w-sm items-center gap-1.5">
                        <Label htmlFor="webhook-url">Endpoint URL</Label>
                        <Input type="url" id="webhook-url" placeholder="https://api.example.com/webhook" />
                    </div>
                    <div className="flex items-center space-x-2">
                         <Switch id="webhook-active" />
                         <Label htmlFor="webhook-active">Active</Label>
                    </div>
                </CardContent>
                <CardFooter>
                    <Button>Save Webhook</Button>
                </CardFooter>
            </Card>
        </TabsContent>
        <TabsContent value="middleware" className="space-y-4">
             <Card>
                <CardHeader>
                    <CardTitle>Middleware Pipeline</CardTitle>
                    <CardDescription>
                        Visual management of the request processing pipeline.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center space-x-2 p-4 border rounded-lg bg-muted/50">
                        <div className="p-2 bg-background border rounded shadow-sm">Auth</div>
                        <div className="h-px w-8 bg-muted-foreground"></div>
                        <div className="p-2 bg-background border rounded shadow-sm">Rate Limit</div>
                        <div className="h-px w-8 bg-muted-foreground"></div>
                        <div className="p-2 bg-background border rounded shadow-sm">Cache</div>
                        <div className="h-px w-8 bg-muted-foreground"></div>
                        <div className="p-2 bg-background border rounded shadow-sm">Router</div>
                    </div>
                </CardContent>
            </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
