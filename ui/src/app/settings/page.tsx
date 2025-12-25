/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";

export default function SettingsPage() {
  const [settings, setSettings] = useState<any>(null);

  useEffect(() => {
    async function fetchSettings() {
      const res = await fetch("/api/settings");
      if (res.ok) {
        setSettings(await res.json());
      }
    }
    fetchSettings();
  }, []);

  if (!settings) return <div>Loading...</div>;

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
      </div>

      <div className="grid gap-4 grid-cols-1 md:grid-cols-2">
        <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader>
                <CardTitle>Global Configuration</CardTitle>
                <CardDescription>Server-wide operational parameters.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="space-y-2">
                    <Label htmlFor="listen_addr">MCP Listen Address</Label>
                    <Input id="listen_addr" defaultValue={settings.mcp_listen_address} />
                </div>
                <div className="space-y-2">
                    <Label htmlFor="log_level">Log Level</Label>
                    <Select defaultValue={settings.log_level}>
                        <SelectTrigger>
                            <SelectValue placeholder="Select log level" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="LOG_LEVEL_DEBUG">Debug</SelectItem>
                            <SelectItem value="LOG_LEVEL_INFO">Info</SelectItem>
                            <SelectItem value="LOG_LEVEL_WARN">Warn</SelectItem>
                            <SelectItem value="LOG_LEVEL_ERROR">Error</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div className="space-y-2">
                    <Label htmlFor="api_key">API Key</Label>
                    <Input id="api_key" type="password" defaultValue={settings.api_key} />
                </div>
                <Button>Save Global Settings</Button>
            </CardContent>
        </Card>

        <Card className="backdrop-blur-sm bg-background/50">
             <CardHeader>
                <CardTitle>Security & Access</CardTitle>
                <CardDescription>Control who can access the server.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="space-y-2">
                    <Label>Allowed IPs</Label>
                    <div className="text-sm text-muted-foreground bg-muted p-2 rounded">
                        {settings.allowed_ips.join("\n")}
                    </div>
                </div>
                 <div className="space-y-2">
                    <Label>Profiles</Label>
                    <div className="flex flex-wrap gap-2">
                         {settings.profiles.map((p: string) => (
                             <div key={p} className="bg-primary/10 text-primary px-2 py-1 rounded text-sm">{p}</div>
                         ))}
                    </div>
                </div>
                 <Button variant="secondary">Manage Profiles</Button>
            </CardContent>
        </Card>
      </div>
    </div>
  );
}
