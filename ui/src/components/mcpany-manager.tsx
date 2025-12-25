/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback, memo } from "react";
import { Server, CheckCircle, XCircle, PowerOff, Power, Trash2 } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import Link from "next/link";
import { UpstreamServiceConfig } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { RegisterServiceDialog } from "./register-service-dialog";

// Helper functions for extracting service details
function getServiceType(service: UpstreamServiceConfig) {
  if (service.grpc_service) return 'gRPC';
  if (service.http_service) return 'HTTP';
  if (service.command_line_service) return 'CLI';
  if (service.openapi_service) return 'OpenAPI';
  if (service.websocket_service) return 'WebSocket';
  if (service.webrtc_service) return 'WebRTC';
  if (service.graphql_service) return 'GraphQL';
  if (service.mcp_service) return 'MCP';
  return 'Unknown';
}

function getServiceAddress(service: UpstreamServiceConfig) {
  return service.grpc_service?.address ||
    service.http_service?.address ||
    service.command_line_service?.command ||
    service.openapi_service?.address ||
    service.websocket_service?.address ||
    service.webrtc_service?.address ||
    service.graphql_service?.address ||
    service.mcp_service?.http_connection?.http_address ||
    service.mcp_service?.stdio_connection?.command;
}

function hasTLS(service: UpstreamServiceConfig) {
  return !!(
    service.grpc_service?.tls_config ||
    service.http_service?.tls_config ||
    service.openapi_service?.tls_config ||
    service.websocket_service?.tls_config ||
    service.webrtc_service?.tls_config ||
    service.mcp_service?.http_connection?.tls_config
  );
}

// ⚡ Bolt: Custom comparison function to prevent unnecessary re-renders.
// The API returns new object references on every fetch, even if data hasn't changed.
// We strictly compare the fields that are actually displayed in the row.
function areServiceRowsEqual(
  prev: { service: UpstreamServiceConfig, onDelete: (name: string) => void },
  next: { service: UpstreamServiceConfig, onDelete: (name: string) => void }
) {
  if (prev.onDelete !== next.onDelete) return false;
  if (prev.service === next.service) return true;

  const s1 = prev.service;
  const s2 = next.service;

  // Compare stable identifiers and basic fields
  if (s1.id !== s2.id) return false;
  if (s1.name !== s2.name) return false;
  if (s1.version !== s2.version) return false;
  if (s1.disable !== s2.disable) return false;

  // Compare derived fields used for rendering
  if (getServiceType(s1) !== getServiceType(s2)) return false;
  if (getServiceAddress(s1) !== getServiceAddress(s2)) return false;
  if (hasTLS(s1) !== hasTLS(s2)) return false;

  return true;
}

// Memoized row component
const ServiceRow = memo(({ service, onDelete }: { service: UpstreamServiceConfig, onDelete: (name: string) => void }) => {
  return (
    <TableRow className={service.disable ? "opacity-50" : ""}>
      <TableCell>
        <Tooltip>
          <TooltipTrigger>
            {service.disable ?
              <PowerOff className="size-5 text-muted-foreground" /> :
              <Power className="size-5 text-primary" />
            }
          </TooltipTrigger>
          <TooltipContent>
            <p>{service.disable ? "Disabled" : "Enabled"}</p>
          </TooltipContent>
        </Tooltip>
      </TableCell>
      <TableCell className="font-medium">
        <Link href={`/service/${encodeURIComponent(service.name)}`} className="hover:underline text-primary/90">
          {service.name}
        </Link>
      </TableCell>
      <TableCell>
        <Badge variant="outline">
          {getServiceType(service)}
        </Badge>
      </TableCell>
      <TableCell className="font-code text-xs truncate max-w-[200px]">
        {getServiceAddress(service)}
      </TableCell>
      <TableCell>
        <Badge variant="secondary">{service.version || 'N/A'}</Badge>
      </TableCell>
      <TableCell className="text-center">
          <Tooltip>
            <TooltipTrigger>
                {hasTLS(service) ? <CheckCircle className="size-5 text-primary mx-auto" /> : <XCircle className="size-5 text-muted-foreground mx-auto"/>}
            </TooltipTrigger>
            <TooltipContent>
              <p>Secure (TLS)</p>
            </TooltipContent>
          </Tooltip>
      </TableCell>
      <TableCell className="text-right">
          <Button variant="ghost" size="icon" onClick={() => onDelete(service.name)} title="Delete Service">
              <Trash2 className="size-4 text-destructive" />
          </Button>
      </TableCell>
    </TableRow>
  );
}, areServiceRowsEqual);
ServiceRow.displayName = "ServiceRow";

export function McpAnyManager() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  const fetchServices = useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await apiClient.listServices();
      setServices(response.services);
      toast({
        title: "Services Loaded",
        description: `Found ${response.services.length} registered services.`,
      });
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to fetch services",
        description: "Could not connect to the MCPAny control plane.",
      });
      console.error("Error fetching services:", error);
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  const handleDelete = useCallback(async (serviceName: string) => {
    if (confirm(`Are you sure you want to delete service "${serviceName}"?`)) {
      try {
        await apiClient.unregisterService(serviceName);
        toast({ title: "Service Deleted", description: `${serviceName} unregistered.` });

        // ⚡ Bolt: Local state update to avoid network roundtrip and loading flash.
        setServices((prev) => prev.filter((s) => s.name !== serviceName));
      } catch (error: any) {
        toast({
          variant: "destructive",
          title: "Deletion Failed",
          description: error.message,
        });
      }
    }
  }, [toast]);


  return (
    <Card className="w-full max-w-6xl h-[90vh] flex flex-col shadow-2xl shadow-primary/5">
      <CardHeader>
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div>
            <CardTitle className="text-2xl font-headline flex items-center gap-3">
              <Server className="text-primary" /> MCPAny Service Registry
            </CardTitle>
            <CardDescription className="mt-1">
              A list of all services registered with the MCPAny control plane.
            </CardDescription>
          </div>
           <div className="flex gap-2">
            <Button variant="outline" onClick={fetchServices} disabled={isLoading}>
                {isLoading ? "Refreshing..." : "Refresh"}
            </Button>
            <RegisterServiceDialog onSuccess={fetchServices} />
           </div>
        </div>
      </CardHeader>

      <CardContent className="flex-grow flex flex-col overflow-hidden">
         <div className="flex-grow overflow-y-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Status</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Address / Command</TableHead>
                <TableHead>Version</TableHead>
                <TableHead className="text-center">TLS</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                 [...Array(4)].map((_, i) => (
                    <TableRow key={i}>
                        <TableCell colSpan={7} className="p-4">
                          <div className="w-full h-8 bg-muted animate-pulse rounded-md" />
                        </TableCell>
                    </TableRow>
                 ))
              )}
              {!isLoading && services.map((service) => (
                <ServiceRow
                  key={service.id || service.name}
                  service={service}
                  onDelete={handleDelete}
                />
              ))}
               {!isLoading && services.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="h-24 text-center">
                    No services registered.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
