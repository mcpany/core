/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { ArrowDown } from "lucide-react";

interface Middleware {
  name: string;
  priority: number;
  disabled: boolean;
  description: string;
}

export default function MiddlewarePage() {
  const [middlewares, setMiddlewares] = useState<Middleware[]>([]);

  useEffect(() => {
    async function fetchMiddlewares() {
      const res = await fetch("/api/middleware");
      if (res.ok) {
        setMiddlewares(await res.json());
      }
    }
    fetchMiddlewares();
  }, []);

  // Sort by priority
  const sortedMiddlewares = [...middlewares].sort((a, b) => a.priority - b.priority);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h2>
      </div>

      <div className="flex flex-col items-center space-y-4 max-w-2xl mx-auto">
         <div className="text-sm text-muted-foreground uppercase tracking-widest font-semibold">Incoming Request</div>
         <ArrowDown className="text-muted-foreground" />

         {sortedMiddlewares.map((mw, index) => (
             <Card key={mw.name} className={`w-full relative ${mw.disabled ? 'opacity-50 border-dashed' : ''} backdrop-blur-sm bg-background/50`}>
                 <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                     <div className="flex items-center space-x-2">
                         <Badge variant="outline" className="mr-2">{mw.priority}</Badge>
                         <CardTitle className="text-lg">{mw.name}</CardTitle>
                     </div>
                     <Switch checked={!mw.disabled} />
                 </CardHeader>
                 <CardContent>
                     <p className="text-sm text-muted-foreground">{mw.description}</p>
                 </CardContent>
             </Card>
         ))}

         <ArrowDown className="text-muted-foreground" />
         <div className="text-sm text-muted-foreground uppercase tracking-widest font-semibold">Service Execution</div>
      </div>
    </div>
  );
}
