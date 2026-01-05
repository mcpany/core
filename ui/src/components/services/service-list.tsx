/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Settings, Trash2, CheckCircle, XCircle } from "lucide-react";
import { UpstreamServiceConfig } from "@/lib/client";


interface ServiceListProps {
  services: UpstreamServiceConfig[];
  isLoading?: boolean;
  onToggle?: (name: string, enabled: boolean) => void;
  onEdit?: (service: UpstreamServiceConfig) => void;
  onDelete?: (name: string) => void;
}

export function ServiceList({ services, isLoading, onToggle, onEdit, onDelete }: ServiceListProps) {

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
    <>
      {/* Desktop View */}
      <div className="hidden md:block rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Address / Command</TableHead>
              <TableHead>Version</TableHead>
              <TableHead className="text-center">Secure</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {services.map((service) => (
              <ServiceRow
                  key={service.name}
                  service={service}
                  onToggle={onToggle}
                  onEdit={onEdit}
                  onDelete={onDelete}
              />
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Mobile View */}
      <div className="grid grid-cols-1 gap-4 md:hidden">
        {services.map((service) => (
            <ServiceCard
              key={service.name}
              service={service}
              onToggle={onToggle}
              onEdit={onEdit}
              onDelete={onDelete}
            />
        ))}
      </div>
    </>
  );
}

function useServiceDetails(service: UpstreamServiceConfig) {
    const type = useMemo(() => {
        if (service.http_service) return "HTTP";
        if (service.grpc_service) return "gRPC";
        if (service.command_line_service) return "CLI";
        if (service.mcp_service) return "MCP";
        return "Other";
    }, [service]);

    const address = useMemo(() => {
         return service.grpc_service?.address ||
            service.http_service?.address ||
            service.command_line_service?.command ||
            service.mcp_service?.http_connection?.http_address ||
            service.mcp_service?.stdio_connection?.command ||
            "-";
    }, [service]);

    const secure = useMemo(() => {
        return !!(service.grpc_service?.tls_config || service.http_service?.tls_config || service.mcp_service?.http_connection?.tls_config);
    }, [service]);

    return { type, address, secure };
}

function ServiceRow({ service, onToggle, onEdit, onDelete }: {
    service: UpstreamServiceConfig,
    onToggle?: (name: string, enabled: boolean) => void,
    onEdit?: (service: UpstreamServiceConfig) => void,
    onDelete?: (name: string) => void
}) {
    const { type, address, secure } = useServiceDetails(service);

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
                 </div>
             </TableCell>
             <TableCell className="font-medium">
                 {service.name}
             </TableCell>
             <TableCell>
                 <Badge variant="outline">{type}</Badge>
             </TableCell>
             <TableCell className="font-mono text-xs max-w-[200px] truncate" title={address}>
                 {address}
             </TableCell>
             <TableCell>
                 {service.version}
             </TableCell>
             <TableCell className="text-center">
                 {secure ? <CheckCircle className="h-4 w-4 text-green-500 mx-auto" /> : <XCircle className="h-4 w-4 text-muted-foreground mx-auto" />}
             </TableCell>
             <TableCell className="text-right">
                 <div className="flex justify-end gap-2">
                    {onEdit && (
                        <Button variant="ghost" size="icon" onClick={() => onEdit(service)} aria-label="Edit">
                            <Settings className="h-4 w-4" />
                        </Button>
                    )}
                    {onDelete && (
                        <Button variant="ghost" size="icon" onClick={() => onDelete(service.name)} className="text-destructive hover:text-destructive" aria-label="Delete">
                             <Trash2 className="h-4 w-4" />
                        </Button>
                    )}
                 </div>
             </TableCell>
        </TableRow>
    );
}

function ServiceCard({ service, onToggle, onEdit, onDelete }: {
    service: UpstreamServiceConfig,
    onToggle?: (name: string, enabled: boolean) => void,
    onEdit?: (service: UpstreamServiceConfig) => void,
    onDelete?: (name: string) => void
}) {
    const { type, address, secure } = useServiceDetails(service);

    return (
      <div className={`p-4 rounded-lg border bg-card text-card-foreground shadow-sm ${service.disable ? "opacity-75 bg-muted/20" : ""}`}>
          <div className="flex items-start justify-between mb-3">
              <div>
                  <h3 className="font-semibold">{service.name}</h3>
                  <div className="text-xs text-muted-foreground mt-0.5">v{service.version}</div>
              </div>
              <Badge variant="outline">{type}</Badge>
          </div>

          <div className="space-y-2 mb-4">
              <div className="text-xs font-mono bg-muted p-2 rounded truncate" title={address}>
                  {address}
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  {secure ? <CheckCircle className="h-3 w-3 text-green-500" /> : <XCircle className="h-3 w-3 text-slate-400" />}
                  {secure ? "Secure Connection" : "Insecure Connection"}
              </div>
          </div>

          <div className="flex items-center justify-between pt-2 border-t">
              <div className="flex items-center gap-2">
                 {onToggle && (
                    <div className="flex items-center gap-2">
                        <Switch
                            id={`mobile-toggle-${service.name}`}
                            checked={!service.disable}
                            onCheckedChange={(checked) => onToggle(service.name, checked)}
                        />
                        <label htmlFor={`mobile-toggle-${service.name}`} className="text-xs text-muted-foreground">
                            {service.disable ? "Disabled" : "Enabled"}
                        </label>
                    </div>
                )}
              </div>
              <div className="flex gap-1">
                {onEdit && (
                    <Button variant="ghost" size="sm" onClick={() => onEdit(service)} className="h-8 w-8 p-0">
                        <Settings className="h-4 w-4" />
                    </Button>
                )}
                {onDelete && (
                    <Button variant="ghost" size="sm" onClick={() => onDelete(service.name)} className="h-8 w-8 p-0 text-destructive hover:text-destructive hover:bg-destructive/10">
                            <Trash2 className="h-4 w-4" />
                    </Button>
                )}
              </div>
          </div>
      </div>
    );
}
