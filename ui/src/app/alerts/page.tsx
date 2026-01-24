/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { AlertList } from "@/components/alerts/alert-list";
import { AlertStats } from "@/components/alerts/alert-stats";
import { CreateRuleDialog } from "@/components/alerts/create-rule-dialog";
import { AlertRulesList } from "@/components/alerts/alert-rules-list";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

/**
 * AlertsPage component.
 * @returns The rendered component.
 */
export default function AlertsPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleRuleCreated = () => {
    setRefreshTrigger(prev => prev + 1);
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Alerts & Incidents</h2>
            <p className="text-muted-foreground">Monitor system health and manage incident response.</p>
        </div>
        <CreateRuleDialog onRuleCreated={handleRuleCreated} />
      </div>

      <div className="space-y-4 flex-1 flex flex-col min-h-0">
        <AlertStats />

        <Tabs defaultValue="incidents" className="flex-1 flex flex-col min-h-0">
            <div className="flex items-center justify-between border-b pb-2">
                <TabsList>
                    <TabsTrigger value="incidents">Incidents</TabsTrigger>
                    <TabsTrigger value="rules">Alert Rules</TabsTrigger>
                </TabsList>
            </div>

            <TabsContent value="incidents" className="flex-1 overflow-auto mt-4 rounded-md border bg-muted/10 p-4">
                <AlertList />
            </TabsContent>

            <TabsContent value="rules" className="flex-1 overflow-auto mt-4 rounded-md border bg-muted/10 p-4">
                <AlertRulesList refreshTrigger={refreshTrigger} />
            </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
