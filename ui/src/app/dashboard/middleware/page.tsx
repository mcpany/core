/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { GripVertical } from "lucide-react";

interface Middleware {
  id: string;
  name: string;
  enabled: boolean;
  priority: number;
}

export default function MiddlewarePage() {
  const [middlewares, setMiddlewares] = useState<Middleware[]>([
      { id: "1", name: "Authentication", enabled: true, priority: 1 },
      { id: "2", name: "Rate Limiting", enabled: true, priority: 2 },
      { id: "3", name: "Logging", enabled: true, priority: 3 },
      { id: "4", name: "Trace Context", enabled: false, priority: 4 },
  ]);

  const toggleMiddleware = (id: string) => {
      setMiddlewares(middlewares.map(m => m.id === id ? { ...m, enabled: !m.enabled } : m));
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Middleware</h2>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Pipeline Configuration</CardTitle>
          <CardDescription>Drag and drop to reorder middleware execution flow.</CardDescription>
        </CardHeader>
        <CardContent>
           <div className="space-y-4">
               {middlewares.sort((a,b) => a.priority - b.priority).map((m) => (
                   <div key={m.id} className="flex items-center justify-between p-4 border rounded-lg bg-card shadow-sm">
                       <div className="flex items-center gap-4">
                           <GripVertical className="text-muted-foreground cursor-move" />
                           <div className="font-medium">{m.name}</div>
                           <Badge variant="outline">Priority: {m.priority}</Badge>
                       </div>
                       <div className="flex items-center gap-2">
                           <span className="text-sm text-muted-foreground">{m.enabled ? "Enabled" : "Disabled"}</span>
                           <Switch checked={m.enabled} onCheckedChange={() => toggleMiddleware(m.id)} />
                       </div>
                   </div>
               ))}
           </div>
        </CardContent>
      </Card>
    </div>
  );
}
