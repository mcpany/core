/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Server, CheckCircle, XCircle, PowerOff, Power } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import Link from "next/link";
import { UpstreamServiceConfig, ListServicesResponse } from "@/lib/types";
import { apiClient } from "@/lib/client";


export function McpAnyManager() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  const fetchServices = async () => {
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
  };

  useEffect(() => {
    fetchServices();
  }, []);


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
           <Button onClick={fetchServices} disabled={isLoading}>
            {isLoading ? "Refreshing..." : "Refresh Services"}
          </Button>
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
                <TableHead className="text-right">TLS</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                 [...Array(4)].map((_, i) => (
                    <TableRow key={i}>
                        <TableCell colSpan={6} className="p-4">
                          <div className="w-full h-8 bg-muted animate-pulse rounded-md" />
                        </TableCell>
                    </TableRow>
                 ))
              )}
              {!isLoading && services.map((service) => (
                <TableRow key={service.id} className={service.disable ? "opacity-50" : ""}>
                   <TableCell>
                      <TooltipProvider>
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
                      </TooltipProvider>
                  </TableCell>
                  <TableCell className="font-medium">
                    <Link href={`/service/${service.id}`} className="hover:underline text-primary/90">
                      {service.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      { service.grpc_service ? 'gRPC' :
                        service.http_service ? 'HTTP' :
                        service.command_line_service ? 'CLI' :
                        'N/A'
                      }
                    </Badge>
                  </TableCell>
                  <TableCell className="font-code">
                    { service.grpc_service?.address ||
                      service.http_service?.address ||
                      service.command_line_service?.command
                    }
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{service.version || 'N/A'}</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                           {service.grpc_service?.tls_config || service.http_service?.tls_config ? <CheckCircle className="size-5 text-primary" /> : <XCircle className="size-5 text-muted-foreground"/>}
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>{service.grpc_service?.tls_config || service.http_service?.tls_config ? "Secure (TLS)" : "Insecure"}</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </TableCell>
                </TableRow>
              ))}
               {!isLoading && services.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center">
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
