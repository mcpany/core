/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { UpstreamServiceConfig, PromptDefinition } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Book, AlertTriangle, TrendingUp } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { ServicePropertyCard } from "./service-property-card";

/**
 * PromptDetail.
 *
 * @param promptName - The promptName.
 * @param promptName - The promptName.
 */
export function PromptDetail({ serviceId, promptName }: { serviceId: string, promptName: string }) {
  const [prompt, setPrompt] = useState<PromptDefinition | null>(null);
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
        const serviceData = serviceDetails.grpcService || serviceDetails.httpService || serviceDetails.commandLineService || serviceDetails.mcpService;
        const foundPrompt = serviceData?.prompts?.find(p => p.name === promptName);

        if (foundPrompt) {
          setPrompt(foundPrompt);
          const statusRes = await apiClient.getServiceStatus(serviceId);
          setMetrics(statusRes.metrics);
        } else {
          throw new Error(`Prompt "${promptName}" not found in service "${serviceDetails.name}".`);
        }
      } catch (e: any) {
        setError(e.message || "An unknown error occurred.");
        toast({
          variant: "destructive",
          title: "Failed to fetch prompt details",
          description: e.message,
        });
      } finally {
        setIsLoading(false);
      }
    };

    if (serviceId && promptName) {
      fetchDetails();
    }
  }, [serviceId, promptName, toast]);

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

  if (!prompt || !service) {
    return null;
  }

  const usageCount = metrics?.[`prompt_usage:${prompt.name}`] ?? 'N/A';

  return (
    <Card className="w-full max-w-4xl shadow-2xl shadow-primary/5">
      <CardHeader>
        <CardTitle className="text-3xl font-headline flex items-center gap-3">
          <Book className="text-primary size-8" /> {prompt.name}
        </CardTitle>
        <CardDescription className="mt-1">
          Part of the <code className="bg-muted px-1 py-0.5 rounded-sm">{service.name}</code> service.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6">
        <ServicePropertyCard title="Prompt Definition" data={{
            "Name": prompt.name,
            "Description": prompt.description || 'N/A',
        }} />
        <Card>
            <CardHeader>
                <CardTitle className="text-xl flex items-center gap-2"><TrendingUp /> Usage Metrics</CardTitle>
            </CardHeader>
            <CardContent>
                 <dl className="space-y-2">
                    <div className="flex justify-between items-start">
                        <dt className="text-muted-foreground">Total Runs</dt>
                        <dd className="text-right font-mono text-sm">{usageCount.toLocaleString()}</dd>
                    </div>
                 </dl>
            </CardContent>
        </Card>
      </CardContent>
    </Card>
  );
}
