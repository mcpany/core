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
import { MoreHorizontal, Settings, Trash, RefreshCw, AlertCircle, CheckCircle2, CircleOff } from "lucide-react";
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
            <TableHead>Version</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {services.length === 0 && (
              <TableRow>
                  <TableCell colSpan={7} className="text-center h-24 text-muted-foreground">
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
                  {service.toolCount !== undefined ? (
                      <Badge variant="secondary">{service.toolCount}</Badge>
                  ) : (
                      <span className="text-muted-foreground">-</span>
                  )}
              </TableCell>
              <TableCell>{service.version || '-'}</TableCell>
              <TableCell>
                <div className="flex items-center space-x-4">
                    <Switch
                        checked={!service.disable}
                        onCheckedChange={() => onToggle(service)}
                    />
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="flex items-center space-x-1 cursor-help">
                                    {service.disable ? (
                                        <>
                                            <CircleOff className="h-4 w-4 text-muted-foreground" />
                                            <span className="text-sm text-muted-foreground">Disabled</span>
                                        </>
                                    ) : service.lastError ? (
                                        <>
                                            <AlertCircle className="h-4 w-4 text-destructive" />
                                            <span className="text-sm text-destructive font-medium">Error</span>
                                        </>
                                    ) : (
                                        <>
                                            <CheckCircle2 className="h-4 w-4 text-green-600" />
                                            <span className="text-sm text-green-600 font-medium">Active</span>
                                        </>
                                    )}
                                </div>
                            </TooltipTrigger>
                            <TooltipContent>
                                {service.disable ? (
                                    <p>Service is explicitly disabled.</p>
                                ) : service.lastError ? (
                                    <div className="max-w-xs">
                                        <p className="font-semibold">Error:</p>
                                        <p className="text-sm break-words">{service.lastError}</p>
                                    </div>
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
