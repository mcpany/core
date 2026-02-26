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
          // Seed the database with a heavy mock service using the backend seeder.
          // This avoids dependency on local 'go' binary or source code.
          await apiClient.seedData({
              upstream_services: [
                  {
                      id: "context-heavy-mock",
                      name: "Context Heavy Mock",
                      version: "1.0.0",
                      disable: false,
                      priority: 0,
                      // We use a simple echo command which is generally available
                      command_line_service: {
                          command: "echo",
                          args: ["-n", "{\"jsonrpc\": \"2.0\", \"result\": {\"content\": [{\"type\": \"text\", \"text\": \"Mock Data\"}]}, \"id\": 1}"],
                          working_directory: ".",
                          env: {}
                      },
                      tools: [
                          { name: "heavy_tool", description: "Returns a large payload", call_id: "heavy_call" }
                      ],
                      calls: {
                          heavy_call: {
                              // Mapping args to echo logic is tricky without a real script.
                              // But for visualization of *availability*, this is enough.
                              // If we need execution, we might need a better mock or 'cat'.
                          }
                      }
                  }
              ],
              users: [
                  {
                      id: "admin",
                      authentication: {
                          basic_auth: {
                              username: "admin",
                              password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // password
                          }
                      },
                      roles: ["admin"],
                      profile_ids: ["default"]
                  }
              ],
              profiles: [
                  {
                      name: "default",
                      description: "Default Profile"
                  }
              ]
          });

           toast({ title: "Seeded Data", description: "Database seeded with Context Heavy Mock." });

           // Trigger reload to refresh context data
           window.location.reload();

      } catch (e) {
          console.error("Seeding failed", e);
          toast({
              title: "Seeding Failed",
              description: "Could not seed database. Check server logs.",
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
