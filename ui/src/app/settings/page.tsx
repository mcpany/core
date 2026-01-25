/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GlobalSettingsForm } from "@/components/settings/global-settings-form";
import { SecretsManager } from "@/components/settings/secrets-manager";
import { AuthSettingsForm } from "@/components/settings/auth-settings";
import { ProfileManager } from "@/components/settings/profile-manager";

import Link from "next/link";

/**
 * SettingsPage component.
 * @returns The rendered component.
 */
export default function SettingsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
      </div>

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
        <TabsContent value="profiles" className="flex-1 h-full">
            <ProfileManager />
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
