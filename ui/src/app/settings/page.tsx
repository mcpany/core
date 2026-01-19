/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { GlobalSettingsForm } from "@/components/settings/global-settings-form";
import { SecretsManager } from "@/components/settings/secrets-manager";
import { AuthSettingsForm } from "@/components/settings/auth-settings";
import { SystemStatusWarning } from "@/components/settings/system-status-warning";

import Link from "next/link";

export default function SettingsPage() {
  const [activeProfile, setActiveProfile] = useState("development");

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
      </div>

      <SystemStatusWarning />

      <Tabs defaultValue="profiles" className="space-y-4 flex-1 flex flex-col">
        <TabsList>
          <TabsTrigger value="profiles">Profiles</TabsTrigger>
          <TabsTrigger value="webhooks" asChild>
            <Link href="/settings/webhooks">Webhooks</Link>
          </TabsTrigger>
          <TabsTrigger value="secrets">Secrets & Keys</TabsTrigger>
          <TabsTrigger value="auth">Authentication</TabsTrigger>
          <TabsTrigger value="general">General</TabsTrigger>
        </TabsList>
        <TabsContent value="profiles" className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card className="md:col-span-1">
                    <CardHeader>
                        <CardTitle>Execution Profiles</CardTitle>
                        <CardDescription>Select a profile to edit</CardDescription>
                    </CardHeader>
                    <CardContent className="grid gap-2">
                        {['development', 'production', 'debug', 'staging'].map(profile => (
                            <Button
                                key={profile}
                                variant={activeProfile === profile ? "default" : "outline"}
                                className="w-full justify-start capitalize"
                                onClick={() => setActiveProfile(profile)}
                            >
                                {profile}
                            </Button>
                        ))}
                         <Button variant="ghost" className="w-full justify-start text-muted-foreground border-dashed border">
                            + New Profile
                        </Button>
                    </CardContent>
                </Card>

                <Card className="md:col-span-2">
                    <CardHeader>
                        <CardTitle className="capitalize">{activeProfile} Configuration</CardTitle>
                        <CardDescription>Manage environment variables and settings for this profile.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                         <div className="grid gap-2">
                            <Label htmlFor="env-vars">Environment Variables</Label>
                            <Textarea id="env-vars" className="font-mono text-xs h-[200px]" defaultValue={`LOG_LEVEL=debug
MCP_PORT=8080
ENABLE_TRACING=true
ALLOWED_HOSTS=*`} />
                        </div>
                        <div className="flex justify-end">
                            <Button>Save Changes</Button>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </TabsContent>
        <TabsContent value="secrets" className="flex-1 h-full">
            <SecretsManager />
        </TabsContent>
        <TabsContent value="auth">
            <AuthSettingsForm />
        </TabsContent>
        <TabsContent value="general">
             <GlobalSettingsForm />
        </TabsContent>
      </Tabs>
    </div>
  );
}
