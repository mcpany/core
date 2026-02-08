/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useMemo } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { Loader2, Play, Square, Trash2, Edit, ChevronLeft, Layers } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ServiceList } from "@/components/services/service-list";
import { useToast } from "@/hooks/use-toast";
import Link from "next/link";

/**
 * StackDetailsPage component.
 * @returns The rendered component.
 */
export default function StackDetailsPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const name = typeof params.name === 'string' ? decodeURIComponent(params.name) : "";

  const [stack, setStack] = useState<ServiceCollection | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [registeredServices, setRegisteredServices] = useState<UpstreamServiceConfig[]>([]);

  // Fetch stack and current services to determine status
  const loadData = async () => {
    try {
        const [stackData, servicesData] = await Promise.all([
            apiClient.getCollection(name),
            apiClient.listServices()
        ]);
        setStack(stackData);
        // Normalize
        const list = Array.isArray(servicesData) ? servicesData : (servicesData.services || []);
        setRegisteredServices(list);
    } catch (e) {
        console.error("Failed to load stack data", e);
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to load stack details."
        });
    } finally {
        setLoading(false);
    }
  };

  useEffect(() => {
    if (name) loadData();
  }, [name]);

  const handleDeploy = async () => {
      if (!stack) return;
      if (!confirm(`Deploy ${stack.services.length} services from stack "${stack.name}"? This will overwrite existing configurations.`)) return;

      setProcessing(true);
      try {
          // Sequential or Parallel? Parallel is faster but might hit rate limits or race conditions if deps exist.
          // Let's go parallel for now.
          await Promise.all(stack.services.map(s => apiClient.registerService(s)));
          toast({
              title: "Stack Deployed",
              description: `Successfully deployed ${stack.services.length} services.`
          });
          loadData();
      } catch (e: any) {
          console.error("Deploy failed", e);
          toast({
              variant: "destructive",
              title: "Deploy Failed",
              description: e.message || "Failed to deploy some services."
          });
      } finally {
          setProcessing(false);
      }
  };

  const handleStop = async () => {
      if (!stack) return;
      if (!confirm(`Stop (disable) all services in stack "${stack.name}"?`)) return;

      setProcessing(true);
      try {
          await Promise.all(stack.services.map(s => apiClient.setServiceStatus(s.name, true)));
          toast({
              title: "Stack Stopped",
              description: "Services have been disabled."
          });
          loadData();
      } catch (e: any) {
          console.error("Stop failed", e);
          toast({
              variant: "destructive",
              title: "Stop Failed",
              description: e.message || "Failed to stop some services."
          });
      } finally {
          setProcessing(false);
      }
  };

  const handleDeleteStack = async () => {
      if (!confirm(`Are you sure you want to delete the stack definition "${name}"? This will NOT delete the running services.`)) return;

      setProcessing(true);
      try {
          await apiClient.deleteCollection(name);
          toast({
              title: "Stack Deleted",
              description: "Stack definition removed."
          });
          router.push("/stacks");
      } catch (e: any) {
          console.error("Delete failed", e);
          toast({
              variant: "destructive",
              title: "Delete Failed",
              description: e.message || "Failed to delete stack."
          });
          setProcessing(false);
      }
  };

  // Merge stack definition with runtime status
  const mergedServices = useMemo(() => {
      if (!stack) return [];
      const registeredMap = new Map(registeredServices.map(s => [s.name, s]));

      return stack.services.map(s => {
          const registered = registeredMap.get(s.name);
          if (registered) {
              return { ...s, ...registered, _status: 'deployed' };
          }
          return { ...s, _status: 'missing' };
      });
  }, [stack, registeredServices]);

  if (loading) {
      return (
          <div className="flex items-center justify-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (!stack) {
      return (
          <div className="flex items-center justify-center h-full text-muted-foreground">
              Stack not found.
          </div>
      );
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={() => router.push("/stacks")}>
                <ChevronLeft className="h-4 w-4" />
            </Button>
            <div>
                <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                    {stack.name}
                    <Badge variant="outline" className="text-base font-normal px-2 py-0.5">v{stack.version}</Badge>
                </h2>
                <p className="text-muted-foreground">{stack.description}</p>
            </div>
        </div>
        <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => router.push(`/stacks/${encodeURIComponent(stack.name)}/edit`)} disabled={processing}>
                <Edit className="mr-2 h-4 w-4" /> Edit Stack
            </Button>
            <Button variant="outline" onClick={handleStop} disabled={processing} className="text-amber-600 hover:text-amber-700">
                <Square className="mr-2 h-4 w-4 fill-current" /> Stop
            </Button>
            <Button onClick={handleDeploy} disabled={processing} className="bg-green-600 hover:bg-green-700">
                <Play className="mr-2 h-4 w-4 fill-current" /> Deploy
            </Button>
             <Button variant="ghost" size="icon" onClick={handleDeleteStack} disabled={processing} className="text-destructive hover:text-destructive">
                <Trash2 className="h-4 w-4" />
            </Button>
        </div>
      </div>

      <Card className="flex-1 overflow-hidden flex flex-col">
        <CardHeader className="pb-3 border-b bg-muted/20">
            <CardTitle className="text-base font-medium flex items-center gap-2">
                <Layers className="h-4 w-4" /> Services Configuration
            </CardTitle>
            <CardDescription>
                This stack defines {stack.services.length} services.
            </CardDescription>
        </CardHeader>
        <CardContent className="p-0 flex-1 overflow-auto bg-background/50 backdrop-blur-sm">
             <div className="p-6">
                <ServiceList
                    services={mergedServices}
                    // Disable actions that modify the *stack definition* via the list,
                    // but allow inspecting runtime status.
                    // Actually, ServiceList actions usually call API.
                    // If I edit a service here, I am editing the RUNNING service, not the stack definition.
                    // This is acceptable behavior (drift).
                    // But maybe I should warn?
                    // For now, let's allow inspection.
                    onToggle={async (name, enabled) => {
                        await apiClient.setServiceStatus(name, !enabled);
                        loadData();
                    }}
                    onRestart={async (name) => {
                        await apiClient.restartService(name);
                        loadData();
                    }}
                    // Edit goes to service editor, which is fine.
                    onEdit={(s) => router.push(`/upstream-services?service=${s.name}`)} // Ideally open sheet
                />
             </div>
        </CardContent>
      </Card>
    </div>
  );
}
