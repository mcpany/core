/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

export function GlobalSettingsForm() {
    const { toast } = useToast();
    const [loading, setLoading] = useState(true);
    const [settings, setSettings] = useState<any>({});

    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        try {
            const data = await apiClient.getGlobalSettings();
            setSettings(data);
        } catch (error) {
            toast({
                title: "Error",
                description: "Failed to load global settings",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    const saveSettings = async () => {
        try {
            await apiClient.saveGlobalSettings(settings);
            toast({
                title: "Success",
                description: "Global settings saved successfully",
            });
        } catch (error) {
            toast({
                title: "Error",
                description: "Failed to save global settings",
                variant: "destructive",
            });
        }
    };

    if (loading) {
        return <div className="p-4 text-center text-muted-foreground">Loading configuration...</div>;
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>General Settings</CardTitle>
                <CardDescription>Configure global server settings.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="grid gap-2">
                    <Label htmlFor="mcp-address">MCP Listen Address</Label>
                    <Input
                        id="mcp-address"
                        value={settings.mcp_listen_address || ""}
                        onChange={e => setSettings({...settings, mcp_listen_address: e.target.value})}
                        placeholder=":50051"
                    />
                    <p className="text-[0.8rem] text-muted-foreground">The address where the MCP server listens for connections.</p>
                </div>

                <div className="grid gap-2">
                    <Label htmlFor="log-level">Log Level</Label>
                    <Select
                        value={settings.log_level || "INFO"}
                        onValueChange={v => setSettings({...settings, log_level: v})}
                    >
                        <SelectTrigger>
                            <SelectValue placeholder="Select log level" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="DEBUG">DEBUG</SelectItem>
                            <SelectItem value="INFO">INFO</SelectItem>
                            <SelectItem value="WARN">WARN</SelectItem>
                            <SelectItem value="ERROR">ERROR</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <div className="flex items-center justify-between space-x-2 border rounded-lg p-4">
                    <div className="space-y-0.5">
                        <Label htmlFor="audit-enabled">Enable Audit Logging</Label>
                        <p className="text-[0.8rem] text-muted-foreground">Log all access and configuration changes.</p>
                    </div>
                    <Switch
                        id="audit-enabled"
                        checked={settings.audit?.enabled || false}
                        onCheckedChange={checked => setSettings({
                            ...settings,
                            audit: { ...settings.audit, enabled: checked }
                        })}
                    />
                </div>

                 <div className="flex items-center justify-between space-x-2 border rounded-lg p-4">
                    <div className="space-y-0.5">
                        <Label htmlFor="dlp-enabled">Enable DLP (Data Loss Prevention)</Label>
                        <p className="text-[0.8rem] text-muted-foreground">Scan and redact sensitive information in logs and messages.</p>
                    </div>
                    <Switch
                        id="dlp-enabled"
                        checked={settings.dlp?.enabled || false}
                        onCheckedChange={checked => setSettings({
                            ...settings,
                            dlp: { ...settings.dlp, enabled: checked }
                        })}
                    />
                </div>

                <div className="flex justify-end pt-4">
                    <Button onClick={saveSettings}>Save Changes</Button>
                </div>
            </CardContent>
        </Card>
    );
}
