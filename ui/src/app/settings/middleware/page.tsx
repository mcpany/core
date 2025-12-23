/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { GripVertical, Plus } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface Middleware {
  id: string;
  name: string;
  type: "Auth" | "Logging" | "RateLimit" | "Transform";
  enabled: boolean;
}

const initialMiddleware: Middleware[] = [
  { id: "1", name: "Global Rate Limiter", type: "RateLimit", enabled: true },
  { id: "2", name: "Request Logger", type: "Logging", enabled: true },
  { id: "3", name: "JWT Authentication", type: "Auth", enabled: true },
];

export default function MiddlewarePage() {
  const [middlewares, setMiddlewares] = useState(initialMiddleware);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h2>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Add Middleware
        </Button>
      </div>

      <div className="grid gap-6">
        <Card className="backdrop-blur-sm bg-background/50">
          <CardHeader>
            <CardTitle>Request Processing Pipeline</CardTitle>
            <CardDescription>
                Configure the sequence of middleware executed for each request. Drag and drop to reorder.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
                {middlewares.map((mw, index) => (
                    <div key={mw.id} className="flex items-center p-4 border rounded-md bg-card hover:bg-accent/50 transition-colors group">
                        <GripVertical className="h-5 w-5 text-muted-foreground cursor-grab mr-4" />
                        <div className="flex-1">
                            <div className="flex items-center space-x-2">
                                <span className="font-semibold">{mw.name}</span>
                                <Badge variant="outline">{mw.type}</Badge>
                            </div>
                        </div>
                         <div className="flex items-center space-x-4">
                            <Badge variant={mw.enabled ? "default" : "secondary"}>
                                {mw.enabled ? "Active" : "Inactive"}
                            </Badge>
                             <div className="text-xs text-muted-foreground">Order: {index + 1}</div>
                        </div>
                    </div>
                ))}
            </div>
             <div className="mt-4 flex justify-center">
                <div className="h-8 w-0.5 bg-border"></div>
            </div>
             <div className="mt-0 text-center text-sm text-muted-foreground font-mono bg-muted p-2 rounded">
                Upstream Service
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
