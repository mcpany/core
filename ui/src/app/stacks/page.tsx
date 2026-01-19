/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import Link from "next/link";
import { Layers, Cuboid } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  // In a real Portainer, this would list multiple stacks.
  // For MCP Any, we assume one main "MCP Any Stack" for now, or maybe files as stacks?
  // Let's assume one "System" stack.

  const stacks = [
    {
      id: "system",
      name: "mcpany-system",
      status: "active",
      services: "Dynamic",
      type: "Compose"
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
        <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {stacks.map((stack) => (
          <Link key={stack.id} href={`/stacks/${stack.id}`}>
             <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50">
               <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Stack
                  </CardTitle>
                  <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
               </CardHeader>
               <CardContent>
                 <div className="flex items-center gap-3 mb-4">
                    <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                        <Layers className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                        <div className="text-2xl font-bold tracking-tight">{stack.name}</div>
                        <div className="text-xs text-muted-foreground font-mono">{stack.id}</div>
                    </div>
                 </div>

                 <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                    <div className="flex items-center gap-1.5">
                        <span className="relative flex h-2 w-2">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                        </span>
                        Online
                    </div>
                    <div>
                        {stack.services} Services
                    </div>
                 </div>
               </CardContent>
             </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
