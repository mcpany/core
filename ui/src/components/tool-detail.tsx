/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Wrench, AlertTriangle, TrendingUp, Braces } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { ServicePropertyCard } from "./service-property-card";
import { SchemaVisualizer } from "./schema-visualizer";

/**
 * Displays details of a specific tool within a service.
 *
 * @param props - The component props.
 * @param props.serviceId - The ID of the service containing the tool.
 * @param props.toolName - The name of the tool to display.
 * @returns The rendered tool detail card, or null/error state.
 */
export function ToolDetail({ serviceId, toolName }: { serviceId: string, toolName: string }) {
  const [tool, setTool] = useState<ToolDefinition | null>(null);
  const [service, setService] = useState<UpstreamServiceConfig | null>(null);
  const [metrics, setMetrics] = useState<Record<string, number> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    const fetchDetails = async () => {
      setIsLoading(true);
      setError(null);
      try {
        // ⚡ Bolt Optimization: Fetch service details and status (metrics) in parallel
        // This removes the waterfall where we waited for service details before fetching metrics.
        const [serviceRes, statusRes] = await Promise.allSettled([
          apiClient.getService(serviceId),
          apiClient.getServiceStatus(serviceId)
        ]);

        if (serviceRes.status === 'rejected') {
          throw serviceRes.reason;
        }

        const { service: serviceDetails } = serviceRes.value;
        setService(serviceDetails || null);

        if (!serviceDetails) {
            setError("Service not found");
            setIsLoading(false);
            return;
        }
        const serviceData = serviceDetails.grpcService || serviceDetails.httpService || serviceDetails.commandLineService || serviceDetails.openapiService || serviceDetails.websocketService || serviceDetails.webrtcService || serviceDetails.graphqlService || serviceDetails.mcpService;

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const foundTool = (serviceData as any)?.tools?.find((t: ToolDefinition) => t.name === toolName);

        if (foundTool) {
          setTool(foundTool);

          // Only set metrics if the status fetch succeeded
          if (statusRes.status === 'fulfilled') {
             setMetrics(statusRes.value.metrics);
          } else {
             console.warn("Failed to fetch service metrics", statusRes.reason);
          }
        } else {
          throw new Error(`Tool "${toolName}" not found in service "${serviceDetails.name}".`);
        }
      } catch (e) {
        const errorMessage = e instanceof Error ? e.message : String(e);
        setError(errorMessage || "An unknown error occurred.");
        toast({
          variant: "destructive",
          title: "Failed to fetch tool details",
          description: errorMessage,
        });
      } finally {
        setIsLoading(false);
      }
    };

    if (serviceId && toolName) {
      fetchDetails();
    }
  }, [serviceId, toolName, toast]);

  if (isLoading) {
    return (
      <Card className="w-full max-w-4xl">
        <CardHeader>
           <Skeleton className="h-8 w-3/4" />
           <Skeleton className="h-4 w-1/2" />
        </CardHeader>
        <CardContent className="grid gap-6">
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-24 w-full" />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <div className="w-full max-w-4xl">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!tool || !service) {
    return null;
  }

  const usageCount = metrics?.[`tool_usage:${tool.name}`] ?? 'N/A';

  return (
    <Card className="w-full max-w-4xl shadow-2xl shadow-primary/5">
      <CardHeader>
        <CardTitle className="text-3xl font-headline flex items-center gap-3">
          <Wrench className="text-primary size-8" /> {tool.name}
        </CardTitle>
        <CardDescription className="mt-1">
          Part of the <code className="bg-muted px-1 py-0.5 rounded-sm">{service.name}</code> service.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6">
        <ServicePropertyCard title="Tool Definition" data={{
            "Name": tool.name,
            "Description": tool.description || "N/A",
            //"Source": tool.source || "N/A"
        }} />

        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center gap-2">
              <Braces className="h-5 w-5" /> Input Schema
            </CardTitle>
          </CardHeader>
          <CardContent>
             {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
             <SchemaVisualizer schema={tool.inputSchema || (tool as any).input_schema} />
          </CardContent>
        </Card>

        <Card>
            <CardHeader>
                <CardTitle className="text-xl flex items-center gap-2"><TrendingUp /> Usage Metrics</CardTitle>
            </CardHeader>
            <CardContent>
                 <dl className="space-y-2">
                    <div className="flex justify-between items-start">
                        <dt className="text-muted-foreground">Total Calls</dt>
                        <dd className="text-right font-mono text-sm">{usageCount.toLocaleString()}</dd>
                    </div>
                 </dl>
            </CardContent>
        </Card>
      </CardContent>
    </Card>
  );
}
