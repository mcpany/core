/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { NetworkGraph } from "@/components/network-graph/network-graph";
import { Service } from "@/types/service";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Share2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function GraphPage() {
  const [services, setServices] = useState<Service[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchServices = async () => {
    setIsLoading(true);
    try {
        const res = await fetch("/api/services");
        if (res.ok) {
            setServices(await res.json());
        }
    } catch (error) {
        console.error("Failed to fetch services", error);
    } finally {
        setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchServices();
  }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                <Share2 className="size-8" /> Network Graph
            </h2>
            <p className="text-muted-foreground mt-1">
                Visualize the topology of your MCP Gateway and connected services.
            </p>
        </div>
        <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-base py-1">
                {services.length} Services
            </Badge>
            <Button variant="outline" size="icon" onClick={fetchServices} disabled={isLoading}>
                <RefreshCw className={`size-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
        </div>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm bg-background/50 backdrop-blur-sm">
        <CardContent className="flex-1 p-0 relative">
             <NetworkGraph services={services} />
        </CardContent>
      </Card>
    </div>
  );
}
