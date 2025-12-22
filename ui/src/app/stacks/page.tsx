/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import Link from "next/link";
import { Layers, Cuboid } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

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

      <div className="grid gap-4">
        {stacks.map((stack) => (
          <Link key={stack.id} href={`/stacks/${stack.id}`}>
             <Card className="hover:bg-muted/50 transition-colors cursor-pointer border-l-4 border-l-blue-500">
               <CardHeader className="flex flex-row items-center gap-4 py-4">
                  <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                    <Layers className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                  </div>
                  <div className="flex-1">
                    <CardTitle className="text-lg">{stack.name}</CardTitle>
                    <CardDescription>Total Services: {stack.services}</CardDescription>
                  </div>
                   <Badge variant="secondary" className="bg-green-100 text-green-800 hover:bg-green-100 dark:bg-green-900 dark:text-green-300">
                      {stack.status}
                   </Badge>
               </CardHeader>
             </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
