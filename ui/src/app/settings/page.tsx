/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GlobalSettingsForm } from "@/components/settings/global-settings-form";
import { AuthSettingsForm } from "@/components/settings/auth-settings";
import { LLMProviderSettings } from "@/components/settings/llm-provider-settings";

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

      <Tabs defaultValue="general" className="space-y-4 flex-1 flex flex-col">
        <TabsList>
          <TabsTrigger value="general">Global Config</TabsTrigger>
          <TabsTrigger value="auth">Authentication</TabsTrigger>
          <TabsTrigger value="ai">AI Providers</TabsTrigger>
        </TabsList>
        <TabsContent value="auth">
            <AuthSettingsForm />
        </TabsContent>
        <TabsContent value="general">
             <GlobalSettingsForm />
        </TabsContent>
        <TabsContent value="ai">
             <LLMProviderSettings />
        </TabsContent>
      </Tabs>
    </div>
  );
}
