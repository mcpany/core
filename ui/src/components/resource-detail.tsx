/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { UpstreamServiceConfig, ResourceDefinition } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Database, AlertTriangle, TrendingUp } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { ServicePropertyCard } from "./service-property-card";

/**
 * ResourceDetail.
 *
 * @param resourceName - The resourceName.
 * @param resourceName - The resourceName.
 */
export function ResourceDetail({ serviceId, resourceName }: { serviceId: string, resourceName: string }) {
  const [resource, setResource] = useState<ResourceDefinition | null>(null);
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
        const { service: serviceDetails } = await apiClient.getService(serviceId);
        setService(serviceDetails || null);

        if (!serviceDetails) {
            setError("Service not found");
            setIsLoading(false);
            return;
        }
        const serviceData = serviceDetails.grpcService || serviceDetails.httpService || serviceDetails.commandLineService ||
            serviceDetails.openapiService ||
            serviceDetails.websocketService ||
            serviceDetails.webrtcService ||
            serviceDetails.graphqlService ||
            serviceDetails.mcpService;

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const foundResource = (serviceData as any)?.resources?.find((r: any) => r.name === resourceName);

        if (foundResource) {
          setResource(foundResource);
          const statusRes = await apiClient.getServiceStatus(serviceId);
          setMetrics(statusRes.metrics);
        } else {
          throw new Error(`Resource "${resourceName}" not found in service "${serviceDetails.name}".`);
        }
      } catch (e: any) {
        setError(e.message || "An unknown error occurred.");
        toast({
          variant: "destructive",
          title: "Failed to fetch resource details",
          description: e.message,
        });
      } finally {
        setIsLoading(false);
      }
    };

    if (serviceId && resourceName) {
      fetchDetails();
    }
  }, [serviceId, resourceName, toast]);

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

  if (!resource || !service) {
    return null;
  }

  const usageCount = metrics?.[`resource_access:${resource.name}`] ?? 'N/A';

  return (
    <Card className="w-full max-w-4xl shadow-2xl shadow-primary/5">
      <CardHeader>
        <CardTitle className="text-3xl font-headline flex items-center gap-3">
          <Database className="text-primary size-8" /> {resource.name}
        </CardTitle>
        <CardDescription className="mt-1">
          Part of the <code className="bg-muted px-1 py-0.5 rounded-sm">{service.name}</code> service.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6">
        <ServicePropertyCard title="Resource Definition" data={{
            "Name": resource.name,
            "Mime Type": resource.mimeType || "N/A",
        }} />
        <Card>
            <CardHeader>
                <CardTitle className="text-xl flex items-center gap-2"><TrendingUp /> Usage Metrics</CardTitle>
            </CardHeader>
            <CardContent>
                 <dl className="space-y-2">
                    <div className="flex justify-between items-start">
                        <dt className="text-muted-foreground">Total Accesses</dt>
                        <dd className="text-right font-mono text-sm">{usageCount.toLocaleString()}</dd>
                    </div>
                 </dl>
            </CardContent>
        </Card>
      </CardContent>
    </Card>
  );
}
