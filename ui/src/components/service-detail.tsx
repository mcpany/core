/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
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
import { RegisterServiceDialog } from "./register-service-dialog";
import { Button } from "@/components/ui/button";

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
                   <Link href={`/service/${encodeURIComponent(serviceId)}/${linkPath}/${encodeURIComponent(item.name)}`} className="hover:underline text-primary/90">
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

  const fetchService = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await apiClient.getService(serviceId);
        setService(response.service || null);
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

  useEffect(() => {
    if (serviceId) {
      fetchService();
    }
  }, [serviceId]);

  const handleStatusChange = async (enabled: boolean) => {
    if (!service) return;

    const originalStatus = !service.disable;
    setService({ ...service, disable: !enabled });

    try {
      await apiClient.setServiceStatus(service.name, !enabled);
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

  const handleConnect = async () => {
    if (!service) return;
    try {
        // Use current window location as redirect URL basis, but append /auth/callback or similar if UI handles it?
        // Actually the backend expects a redirect URL that *it* will redirect the user to after upstream auth.
        // If we want the user to come back to the UI, we should pass a UI URL.
        // But the backend `handleOAuthCallback` *also* takes a `redirectURL` (to verify against what was sent mostly or to redirect?).
        // Wait, standard OAuth:
        // 1. Client (us) sends `redirect_uri` to Provider (via Backend).
        // 2. Provider redirects to `redirect_uri` with code.
        // 3. Backend exchanges code.
        // If `redirect_uri` points to the BACKEND directly (`/auth/oauth/callback`), the backend gets the code immediately.
        // But then the backend needs to know where to redirect the user *after* saving the token.
        // My implementation in `interactive.go`:
        // `InitiateOAuth` takes `redirectURL`. This is passed to `conf.AuthCodeURL`.
        // So this MUST be where the Provider redirects the browser.
        // If we want the Backend to handle it, `redirectURL` should be `https://backend/auth/oauth/callback`.
        // But `interactive.go` logic seems to imply the *caller* decides.
        // If we pass `window.location.origin + '/oauth/callback'`, the UI must have a route for this.
        // Let's assume we want the UI to handle the callback code and pass it to backend?
        // My backend API `handleOAuthCallback` takes `code` in the body.
        // This confirms the UI receives the code and posts it to backend.
        // So `redirectURL` should be a page in the UI. e.g. `/services/{id}/callback` or global `/oauth/callback`.
        // Let's use `/auth/callback` in UI.

        const redirectUrl = `${window.location.origin}/auth/callback`;
        const { authorization_url, state } = await apiClient.initiateOAuth(serviceId, redirectUrl);

        // Store state/serviceId in local storage to verify/resume?
        sessionStorage.setItem('oauth_state', state);
        sessionStorage.setItem('oauth_service_id', serviceId);
        sessionStorage.setItem('oauth_redirect_url', redirectUrl);

        // Redirect user
        window.location.href = authorization_url;
    } catch (e: any) {
         toast({
            variant: "destructive",
            title: "OAuth Initiation Failed",
            description: e.message,
        });
    }
  };

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

  const serviceType =
        service.grpcService ? 'gRPC' :
        service.httpService ? 'HTTP' :
        service.commandLineService ? 'CLI' :
        service.openapiService ? 'OpenAPI' :
        service.websocketService ? 'WebSocket' :
        service.webrtcService ? 'WebRTC' :
        service.graphqlService ? 'GraphQL' :
        service.mcpService ? 'MCP' :
        'Unknown';

  const isEnabled = !service.disable;

  const serviceData =
        service.grpcService ||
        service.httpService ||
        service.commandLineService ||
        service.openapiService ||
        service.websocketService ||
        service.webrtcService ||
        service.graphqlService ||
        service.mcpService;

  // Type guard or loose access
  const tools = (serviceData as any)?.tools;
  const prompts = (serviceData as any)?.prompts;
  const resources = (serviceData as any)?.resources;


  return (
    <Card className="w-full max-w-6xl shadow-2xl shadow-primary/5">
      <CardHeader>
        <div className="flex justify-between items-start">
          <div>
            <CardTitle className="text-3xl font-headline flex items-center gap-3">
              <Server className="text-primary size-8" /> {service.name}
            </CardTitle>
            <CardDescription className="mt-1">
              Service ID: <code className="bg-muted px-1 py-0.5 rounded-sm">{service.id || "N/A"}</code>
            </CardDescription>
             <Badge variant={isEnabled ? "default" : "secondary"} className="mt-2">
              {isEnabled ? "Enabled" : "Disabled"}
            </Badge>
          </div>
           <div className="flex items-center space-x-4">
             <RegisterServiceDialog
                serviceToEdit={service}
                onSuccess={fetchService}
                trigger={<Button variant="outline" size="sm"><Wrench className="mr-2 h-4 w-4"/> Edit Config</Button>}
             />
            <div className="flex items-center space-x-2">
                {/* Check if service has OAuth configured. We need to check upstreamAuth?.oauth2.
                    Since UpstreamServiceConfig type might not have it strictly typed in frontend depending on generation,
                    we use loose access or verify type.
                    Assuming 'upstreamAuth' field exists.
                 */}
                 { (service as any).upstreamAuth?.oauth2 && (
                     <Button variant="default" size="sm" onClick={handleConnect}>
                         Connect Account
                     </Button>
                 )}

                <Switch
                id="service-status"
                checked={isEnabled}
                onCheckedChange={handleStatusChange}
                aria-label="Service Status"
                />
                <Label htmlFor="service-status" className="flex flex-col">
                <span>{ isEnabled ? 'Enabled' : 'Disabled' }</span>
                </Label>
            </div>
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
                 {service.grpcService && (
                    <ServicePropertyCard title="gRPC Config" data={{
                        "Address": service.grpcService.address,
                        "Reflection Enabled": service.grpcService.useReflection ? "Yes" : "No",
                    }} />
                )}
                 {service.httpService && (
                    <ServicePropertyCard title="HTTP Config" data={{
                        "Address": service.httpService.address,
                    }} />
                )}
                 {service.commandLineService && (
                    <ServicePropertyCard title="CLI Config" data={{
                        "Command": service.commandLineService.command,
                    }} />
                )}
                {service.openapiService && (
                    <ServicePropertyCard title="OpenAPI Config" data={{
                        "Address": service.openapiService.address,
                        "Spec URL": service.openapiService.specUrl || "N/A",
                    }} />
                )}
                {service.websocketService && (
                    <ServicePropertyCard title="WebSocket Config" data={{
                        "Address": service.websocketService.address,
                    }} />
                )}
                {service.webrtcService && (
                    <ServicePropertyCard title="WebRTC Config" data={{
                        "Address": service.webrtcService.address,
                    }} />
                )}
                {service.graphqlService && (
                    <ServicePropertyCard title="GraphQL Config" data={{
                        "Address": service.graphqlService.address,
                    }} />
                )}
                {service.mcpService && (
                    <ServicePropertyCard title="MCP Upstream Config" data={{
                        "Type": service.mcpService.httpConnection ? "HTTP" : service.mcpService.stdioConnection ? "Stdio" : "Bundle",
                        "Address/Command": service.mcpService.httpConnection?.httpAddress || service.mcpService.stdioConnection?.command || service.mcpService.bundleConnection?.bundlePath || "N/A",
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
