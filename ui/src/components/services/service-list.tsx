/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState, memo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Settings, Trash2, CheckCircle, XCircle, AlertTriangle, MoreHorizontal, Copy, Download, Filter } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { UpstreamServiceConfig } from "@/lib/client";


interface ServiceListProps {
  services: UpstreamServiceConfig[];
  isLoading?: boolean;
  onToggle?: (name: string, enabled: boolean) => void;
  onEdit?: (service: UpstreamServiceConfig) => void;
  onDelete?: (name: string) => void;
  onDuplicate?: (service: UpstreamServiceConfig) => void;
  onExport?: (service: UpstreamServiceConfig) => void;
}

export function ServiceList({ services, isLoading, onToggle, onEdit, onDelete, onDuplicate, onExport }: ServiceListProps) {
  const [tagFilter, setTagFilter] = useState("");

  const filteredServices = useMemo(() => {
    if (!tagFilter) return services;
    return services.filter(s => s.tags?.some(tag => tag.toLowerCase().includes(tagFilter.toLowerCase())));
  }, [services, tagFilter]);

  if (isLoading) {
      return (
          <div className="space-y-4">
               {[...Array(3)].map((_, i) => (
                  <div key={i} className="w-full h-12 bg-muted animate-pulse rounded-md" />
               ))}
          </div>
      );
  }

  if (services.length === 0) {
      return <div className="text-center py-10 text-muted-foreground">No services registered.</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center space-x-2 w-full md:w-1/3">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Filter by tag..."
          value={tagFilter}
          onChange={(e) => setTagFilter(e.target.value)}
          className="h-8"
        />
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Tags</TableHead>
              <TableHead>Address / Command</TableHead>
              <TableHead>Tools</TableHead>
              <TableHead className="text-center">Secure</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredServices.map((service) => (
               <ServiceRow
                  key={service.name}
                  service={service}
                  onToggle={onToggle}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onDuplicate={onDuplicate}
                  onExport={onExport}
               />
            ))}
            {filteredServices.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} className="h-24 text-center">
                  No services match the tag filter.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

const ServiceRow = memo(function ServiceRow({ service, onToggle, onEdit, onDelete, onDuplicate, onExport }: {
    service: UpstreamServiceConfig,
    onToggle?: (name: string, enabled: boolean) => void,
    onEdit?: (service: UpstreamServiceConfig) => void,
    onDelete?: (name: string) => void,
    onDuplicate?: (service: UpstreamServiceConfig) => void,
    onExport?: (service: UpstreamServiceConfig) => void
}) {
    const type = useMemo(() => {
        if (service.httpService) return "HTTP";
        if (service.grpcService) return "gRPC";
        if (service.commandLineService) return "CLI";
        if (service.mcpService) return "MCP";
        return "Other";
    }, [service]);

    const address = useMemo(() => {
         return service.grpcService?.address ||
            service.httpService?.address ||
            service.commandLineService?.command ||
            service.mcpService?.httpConnection?.httpAddress ||
            service.mcpService?.stdioConnection?.command ||
            "-";
    }, [service]);

    const secure = useMemo(() => {
        return !!(service.grpcService?.tlsConfig || service.httpService?.tlsConfig || service.mcpService?.httpConnection?.tlsConfig);
    }, [service]);

    return (
        <TableRow className={service.disable ? "opacity-60 bg-muted/40" : ""}>
             <TableCell>
                 <div className="flex items-center gap-2">
                    {onToggle && (
                        <Switch
                            checked={!service.disable}
                            onCheckedChange={(checked) => onToggle(service.name, checked)}
                        />
                    )}
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger>
                                {service.disable ? (
                                    <Badge variant="secondary">Disabled</Badge>
                                ) : service.lastError ? (
                                    <Badge variant="destructive" className="gap-1">
                                        <AlertTriangle className="h-3 w-3" /> Error
                                    </Badge>
                                ) : (
                                    <Badge variant="default" className="bg-green-600 hover:bg-green-700 gap-1">
                                        <CheckCircle className="h-3 w-3" /> Healthy
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
             <TableCell className="font-medium">
                 <div className="flex items-center gap-2">
                     {service.name}
                 </div>
             </TableCell>
             <TableCell>
                 <Badge variant="outline">{type}</Badge>
             </TableCell>
             <TableCell>
                 <div className="flex flex-wrap gap-1">
                     {service.tags?.map((tag) => (
                         <Badge key={tag} variant="secondary" className="text-xs px-1 py-0 h-5">
                             {tag}
                         </Badge>
                     ))}
                 </div>
             </TableCell>
             <TableCell className="font-mono text-xs max-w-[200px] truncate" title={address}>
                 {address}
             </TableCell>
             <TableCell>
                 <div className="flex items-center gap-2">
                     <span className="font-medium">{service.toolCount || 0}</span>
                     <span className="text-xs text-muted-foreground">tools</span>
                 </div>
             </TableCell>
             <TableCell className="text-center">
                 {secure ? <CheckCircle className="h-4 w-4 text-green-500 mx-auto" /> : <XCircle className="h-4 w-4 text-muted-foreground mx-auto" />}
             </TableCell>
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
                        {onEdit && (
                            <DropdownMenuItem onClick={() => onEdit(service)}>
                                <Settings className="mr-2 h-4 w-4" />
                                Edit
                            </DropdownMenuItem>
                        )}
                        {onDuplicate && (
                             <DropdownMenuItem onClick={() => onDuplicate(service)}>
                                <Copy className="mr-2 h-4 w-4" />
                                Duplicate
                            </DropdownMenuItem>
                        )}
                        {onExport && (
                             <DropdownMenuItem onClick={() => onExport(service)}>
                                <Download className="mr-2 h-4 w-4" />
                                Export
                            </DropdownMenuItem>
                        )}
                        <DropdownMenuSeparator />
                        {onDelete && (
                            <DropdownMenuItem onClick={() => onDelete(service.name)} className="text-destructive focus:text-destructive">
                                <Trash2 className="mr-2 h-4 w-4" />
                                Delete
                            </DropdownMenuItem>
                        )}
                    </DropdownMenuContent>
                 </DropdownMenu>
             </TableCell>
        </TableRow>
    );
});
