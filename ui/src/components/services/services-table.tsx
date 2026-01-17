/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { MoreHorizontal, Settings, Trash, RefreshCw, AlertCircle, CheckCircle2 } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import Link from "next/link";
import { UpstreamServiceConfig } from "@/lib/client";

interface ServicesTableProps {
    services: UpstreamServiceConfig[];
    loading: boolean;
    onToggle: (service: UpstreamServiceConfig) => void;
    onDelete: (service: UpstreamServiceConfig) => void;
}

export function ServicesTable({ services, loading, onToggle, onDelete }: ServicesTableProps) {

  if (loading) {
      return <div className="p-4 text-center text-muted-foreground">Loading services...</div>;
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Tools</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {services.length === 0 && (
              <TableRow>
                  <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">
                      No services registered.
                  </TableCell>
              </TableRow>
          )}
          {services.map((service) => (
            <TableRow key={service.id || service.name}>
              <TableCell className="font-medium">
                <Link href={`/services/${service.name}`} className="hover:underline">
                    {service.name}
                </Link>
              </TableCell>
              <TableCell>
                <Badge variant="outline">
                    {service.grpcService ? "gRPC" :
                     service.httpService ? "HTTP" :
                     service.commandLineService ? "CMD" :
                     service.openapiService ? "OpenAPI" : "Unknown"}
                </Badge>
              </TableCell>
              <TableCell>
                  <div className="flex items-center gap-2">
                      <span className="font-medium">{service.toolCount || 0}</span>
                  </div>
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-2">
                    <Switch
                        checked={!service.disable}
                        onCheckedChange={() => onToggle(service)}
                    />
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger>
                                {service.disable ? (
                                    <Badge variant="secondary">Disabled</Badge>
                                ) : service.lastError ? (
                                    <Badge variant="destructive" className="gap-1">
                                        <AlertCircle className="h-3 w-3" /> Error
                                    </Badge>
                                ) : (
                                    <Badge variant="default" className="bg-green-600 hover:bg-green-700 gap-1">
                                        <CheckCircle2 className="h-3 w-3" /> Healthy
                                    </Badge>
                                )}
                            </TooltipTrigger>
                            <TooltipContent>
                                {service.disable ? (
                                    <p>Service is disabled.</p>
                                ) : service.lastError ? (
                                    <p className="text-red-500 font-medium">Error: {service.lastError}</p>
                                ) : (
                                    <p>Service is active and healthy.</p>
                                )}
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
              </TableCell>
              <TableCell>{service.priority || 0}</TableCell>
              <TableCell className="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0">
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                    <DropdownMenuItem>
                      <Settings className="mr-2 h-4 w-4" /> Configure
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-red-600" onClick={() => onDelete(service)}>
                      <Trash className="mr-2 h-4 w-4" /> Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
