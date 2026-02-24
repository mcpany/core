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
          // Attempt to register a mock service using the in-tree mock server.
          // This requires the server to be running in an environment where 'go' is available
          // or the binary is built.
          await apiClient.registerService({
              id: "context-heavy-mock",
              name: "Context Heavy Mock",
              version: "1.0.0",
              disable: false,
              priority: 0,
              commandLineService: {
                  command: "go run server/cmd/mock_mcp_server/main.go",
                  workingDirectory: ".",
                  env: {}
              }
           } as any);

           toast({ title: "Seeded Mock Service", description: "Registered 'Context Heavy Mock'." });

           // Trigger reload to refresh context data
           window.location.reload();

      } catch (e) {
          console.error("Seeding failed", e);
          toast({
              title: "Seeding Failed",
              description: "Could not register mock service. Ensure 'go' is in the server path or use an existing service.",
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
