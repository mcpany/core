/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { AlertList } from "@/components/alerts/alert-list";
import { AlertStats } from "@/components/alerts/alert-stats";
import { CreateRuleDialog } from "@/components/alerts/create-rule-dialog";

/**
 * AlertsPage component.
 * @returns The rendered component.
 */
export default function AlertsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Alerts & Incidents</h2>
            <p className="text-muted-foreground">Monitor system health and manage incident response.</p>
        </div>
        <CreateRuleDialog />
      </div>

      <div className="space-y-4 flex-1 flex flex-col min-h-0">
        <AlertStats />
        <div className="flex-1 overflow-auto rounded-md border bg-muted/10 p-4">
             <AlertList />
        </div>
      </div>
    </div>
  );
}
