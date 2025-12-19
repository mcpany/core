/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, use } from "react";
import Link from "next/link";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { UpstreamServiceConfig, ToolDefinition, PromptDefinition, ResourceDefinition } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Server, AlertTriangle, Wrench, Book, Database, Settings, TrendingUp } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { ServicePropertyCard } from "./service-property-card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Badge } from "./ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { FileConfigCard } from "./file-config-card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

function DefinitionsTable<T extends { name: string; description?: string; type?: string; source?: string; }>({ title, data, icon, serviceId, linkPath }: { title: string; data?: T[], icon: React.ReactNode, serviceId: string, linkPath: string }) {
  if (!data || data.length === 0) {
    return (
       <Card>
        <CardHeader>
          <CardTitle className="text-xl flex items-center gap-2">{icon}{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">No {title.toLowerCase()} configured for this service.</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-xl flex items-center gap-2">{icon}{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              { 'source' in data[0] && <TableHead>Source</TableHead>}
              { 'type' in data[0] && <TableHead>Type</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map((item) => (
              <TableRow key={item.name}>
                <TableCell className="font-medium">
                   <Link href={`/service/${serviceId}/${linkPath}/${encodeURIComponent(item.name)}`} className="hover:underline text-primary/90">
                    {item.name}
                  </Link>
                </TableCell>
                <TableCell className="text-muted-foreground">{item.description}</TableCell>
                 { 'source' in item && item.source && <TableCell><Badge variant={item.source === 'configured' ? "outline" : "secondary"}>{item.source}</Badge></TableCell>}
                 { 'type' in item && item.type && <TableCell><Badge variant="outline">{item.type}</Badge></TableCell>}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}

function MetricsCard({ serviceId }: { serviceId: string }) {
    const [metrics, setMetrics] = useState<Record<string, number> | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        setIsLoading(true);
        apiClient.getServiceStatus(serviceId)
            .then(res => setMetrics(res.metrics))
            .catch(() => setMetrics({}))
            .finally(() => setIsLoading(false));
    }, [serviceId])

    if (isLoading) {
        return <Skeleton className="h-48 w-full" />
    }

    if (!metrics || Object.keys(metrics).length === 0) {
        return (
             <Card>
                <CardHeader>
                    <CardTitle className="text-xl flex items-center gap-2"><TrendingUp /> Usage Metrics</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground text-sm">No metrics available for this service.</p>
                </CardContent>
            </Card>
        )
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-xl flex items-center gap-2"><TrendingUp /> Usage Metrics</CardTitle>
            </CardHeader>
            <CardContent>
                 <dl className="space-y-2">
                    {Object.entries(metrics).map(([key, value]) => (
                        <div key={key} className="flex justify-between items-start">
                            <dt className="text-muted-foreground capitalize">{key.replace(/_/g, ' ')}</dt>
                            <dd className="text-right font-mono text-sm">{value.toLocaleString()}</dd>
                        </div>
                    ))}
                 </dl>
            </CardContent>
        </Card>
    )
}

export function ServiceDetail({ serviceId }: { serviceId: string }) {
  const [service, setService] = useState<UpstreamServiceConfig | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    const fetchService = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await apiClient.getService(serviceId);
        setService(response.service);
      } catch (e: any) {
        setError(e.message || "An unknown error occurred.");
        toast({
          variant: "destructive",
          title: "Failed to fetch service details",
          description: e.message || `Could not find service with ID ${serviceId}.`,
        });
      } finally {
        setIsLoading(false);
      }
    };

    if (serviceId) {
      fetchService();
    }
  }, [serviceId, toast]);

  const handleStatusChange = async (enabled: boolean) => {
    if (!service) return;

    const originalStatus = !service.disable;
    setService({ ...service, disable: !enabled });

    try {
      await apiClient.setServiceStatus(service.id, !enabled);
      toast({
        title: "Service status updated",
        description: `${service.name} has been ${enabled ? 'enabled' : 'disabled'}.`,
      });
    } catch(e: any) {
      setService({ ...service, disable: !originalStatus });
       toast({
        variant: "destructive",
        title: "Update Failed",
        description: `Could not update status for ${service.name}.`,
      });
    }
  }

  if (isLoading) {
    return (
      <Card className="w-full max-w-6xl">
        <CardHeader>
           <Skeleton className="h-8 w-3/4" />
           <Skeleton className="h-4 w-1/2" />
        </CardHeader>
        <CardContent className="grid gap-6">
            <Skeleton className="h-48 w-full" />
            <Skeleton className="h-48 w-full" />
            <Skeleton className="h-48 w-full" />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <div className="w-full max-w-6xl">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!service) {
    return null;
  }

  const serviceType = service.grpc_service ? 'gRPC' : service.http_service ? 'HTTP' : service.command_line_service ? 'CLI' : 'Unknown';
  const isEnabled = !service.disable;

  const serviceData = service.grpc_service || service.http_service || service.command_line_service;
  const tools = serviceData?.tools;
  const prompts = serviceData?.prompts;
  const resources = serviceData?.resources;


  return (
    <Card className="w-full max-w-6xl shadow-2xl shadow-primary/5">
      <CardHeader>
        <div className="flex justify-between items-start">
          <div>
            <CardTitle className="text-3xl font-headline flex items-center gap-3">
              <Server className="text-primary size-8" /> {service.name}
            </CardTitle>
            <CardDescription className="mt-1">
              Service ID: <code className="bg-muted px-1 py-0.5 rounded-sm">{service.id}</code>
            </CardDescription>
             <Badge variant={isEnabled ? "default" : "secondary"} className="mt-2">
              {isEnabled ? "Enabled" : "Disabled"}
            </Badge>
          </div>
           <div className="flex items-center space-x-2">
            <Switch
              id="service-status"
              checked={isEnabled}
              onCheckedChange={handleStatusChange}
              aria-label="Service Status"
            />
            <Label htmlFor="service-status" className="flex flex-col">
              <span>{ isEnabled ? 'Enabled' : 'Disabled' }</span>
              <span className="text-xs text-muted-foreground">Toggle service on/off</span>
            </Label>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="general">
            <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="general">General</TabsTrigger>
                <TabsTrigger value="configuration"><Settings className="mr-2" />Configuration</TabsTrigger>
                <TabsTrigger value="metrics"><TrendingUp className="mr-2"/>Metrics</TabsTrigger>
            </TabsList>
             <TabsContent value="general" className="mt-4 grid gap-6">
                <ServicePropertyCard title="General" data={{
                    "Version": service.version || "N/A",
                    "Service Type": serviceType,
                }} />
                <DefinitionsTable title="Tools" data={tools} icon={<Wrench />} serviceId={serviceId} linkPath="tool" />
                <DefinitionsTable title="Prompts" data={prompts} icon={<Book />} serviceId={serviceId} linkPath="prompt" />
                <DefinitionsTable title="Resources" data={resources} icon={<Database />} serviceId={serviceId} linkPath="resource" />
            </TabsContent>
            <TabsContent value="configuration" className="mt-4 grid gap-6">
                 {service.grpc_service && (
                    <ServicePropertyCard title="gRPC Config" data={{
                        "Address": service.grpc_service.address,
                        "Reflection Enabled": service.grpc_service.use_reflection ? "Yes" : "No",
                    }} />
                )}
                 {service.http_service && (
                    <ServicePropertyCard title="HTTP Config" data={{
                        "Address": service.http_service.address,
                    }} />
                )}
                 {service.command_line_service && (
                    <ServicePropertyCard title="CLI Config" data={{
                        "Command": service.command_line_service.command,
                    }} />
                )}

                {(service.grpc_service?.tls_config || service.http_service?.tls_config) && (
                     <ServicePropertyCard title="TLS Config" data={{
                        "Server Name": service.grpc_service?.tls_config?.server_name || service.http_service?.tls_config?.server_name,
                        "Skip Verify": (service.grpc_service?.tls_config?.insecure_skip_verify || service.http_service?.tls_config?.insecure_skip_verify) ? "Yes" : "No",
                     }} />
                )}
                <FileConfigCard service={service} />
            </TabsContent>
            <TabsContent value="metrics" className="mt-4 grid gap-6">
                <MetricsCard serviceId={serviceId} />
            </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
