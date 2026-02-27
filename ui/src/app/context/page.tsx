/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Loader2, Zap } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { ContextProvider } from "@/components/context/context-provider";
import { ContextTreemap } from "@/components/context/context-treemap";
import { ContextSimulator } from "@/components/context/context-simulator";

/**
 * ContextPage component.
 * Displays the Recursive Context Dashboard.
 * @returns The rendered component.
 */
export default function ContextPage() {
  const [seeding, setSeeding] = useState(false);
  const { toast } = useToast();

  const handleSeedData = async () => {
      setSeeding(true);
      try {
          // Use the server-side seeder endpoint instead of client-side logic
          await apiClient.seedTrafficData([{ serviceId: "context-heavy-mock", count: 100 }]); // Example payload

          // OR more generically, if we add a dedicated generic seed endpoint in client.ts
          // For now, let's assume seedTrafficData is what we have or we add a new one.
          // The previous code was registering a service directly.
          // The new strategy is to use the debug seeder.

          // Using a generic debug seed call if available, or just keeping the register call IF it's valid,
          // but the instructions say "Remove client-side mocks".
          // The previous code was NOT a mock in the sense of fake data, it was registering a REAL service (mock_mcp_server).
          // However, relying on "go run" in the client is fragile.
          // We should ask the backend to seed this.

          await fetch("/api/v1/debug/seed", { method: "POST" });

           toast({ title: "Seeded Data", description: "Triggered server-side seeding." });

           // Trigger reload to refresh context data
           window.location.reload();

      } catch (e) {
          console.error("Seeding failed", e);
          toast({
              title: "Seeding Failed",
              description: "Could not seed data.",
              variant: "destructive"
          });
      } finally {
          setSeeding(false);
      }
  };

  return (
    <ContextProvider>
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
        <div className="flex items-center justify-between">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Recursive Context Dashboard</h1>
                <p className="text-muted-foreground">Visualize and optimize your agent's context window.</p>
            </div>
            <div className="flex items-center gap-2">
                 <Button variant="outline" size="sm" onClick={handleSeedData} disabled={seeding}>
                    {seeding ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Zap className="mr-2 h-4 w-4" />}
                    Seed Data
                </Button>
            </div>
        </div>

        <div className="flex flex-1 gap-4 min-h-0">
             {/* Main Visualization Area */}
             <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
                <ContextTreemap />
             </Card>

             {/* Simulator Sidebar */}
             <div className="w-[350px] flex-none">
                 <ContextSimulator />
             </div>
        </div>
        </div>
    </ContextProvider>
  );
}
