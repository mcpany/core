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
          // Use the real backend seeding API which is reliable and consistent
          const res = await fetch('/api/v1/debug/seed', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                  upstream_services: [{
                      id: "context-heavy-mock",
                      name: "Context Heavy Mock",
                      version: "1.0.0",
                      http_service: {
                          address: "http://example.com/mock", // Placeholder, seed logic handles data
                          tools: [
                              { name: "heavy_tool", description: "Generates heavy context", call_id: "heavy" }
                          ]
                      }
                  }],
                  service_templates: [],
                  users: [],
                  credentials: [],
                  secrets: [],
                  profiles: []
              })
          });

          if (!res.ok) {
              throw new Error(`Failed to seed: ${res.statusText}`);
          }

           toast({ title: "Seeded Data", description: "Successfully seeded context data via API." });

           // Trigger reload to refresh context data
           window.location.reload();

      } catch (e) {
          console.error("Seeding failed", e);
          toast({
              title: "Seeding Failed",
              description: "Could not seed data. Check server logs.",
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
