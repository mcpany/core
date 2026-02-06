/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState, memo, useCallback, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import Link from "next/link";
import { Settings, Trash2, CheckCircle, XCircle, AlertTriangle, MoreHorizontal, Copy, Download, Filter, PlayCircle, PauseCircle, Activity, RefreshCw, Terminal } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { UpstreamServiceConfig } from "@/lib/client";
import { ConnectionDiagnosticDialog } from "@/components/diagnostics/connection-diagnostic";
import { useServiceHealth } from "@/contexts/service-health-context";
import { Sparkline } from "@/components/charts/sparkline";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";


interface ServiceListProps {
  services: UpstreamServiceConfig[];
  isLoading?: boolean;
  onToggle?: (name: string, enabled: boolean) => void;
  onEdit?: (service: UpstreamServiceConfig) => void;
  onDelete?: (name: string) => void;
  onDuplicate?: (service: UpstreamServiceConfig) => void;
  onExport?: (service: UpstreamServiceConfig) => void;
  onBulkToggle?: (names: string[], enabled: boolean) => void;
  onBulkDelete?: (names: string[]) => void;
  onLogin?: (service: UpstreamServiceConfig) => void;
  onRestart?: (name: string) => void;
  onBulkEdit?: (names: string[], updates: { tags?: string[], resilience?: { timeout?: string, maxRetries?: number } }) => void;
}

/**
 * ServiceList.
 *
 * @param onExport - The onExport.
 */
export function ServiceList({ services, isLoading, onToggle, onEdit, onDelete, onDuplicate, onExport, onBulkToggle, onBulkDelete, onLogin, onRestart, onBulkEdit }: ServiceListProps) {
  const [tagFilter, setTagFilter] = useState("");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [isBulkEditDialogOpen, setIsBulkEditDialogOpen] = useState(false);
  const [bulkTags, setBulkTags] = useState("");
  const [bulkTimeout, setBulkTimeout] = useState("");
  const [bulkRetries, setBulkRetries] = useState("");

  const filteredServices = useMemo(() => {
    if (!tagFilter) return services;
    return services.filter(s => s.tags?.some(tag => tag.toLowerCase().includes(tagFilter.toLowerCase())));
  }, [services, tagFilter]);

  // Reset selection when filtering changes or services change
  useEffect(() => {
      setSelected(new Set());
  }, [tagFilter]);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelected(new Set(filteredServices.map(s => s.name)));
    } else {
      setSelected(new Set());
    }
  }, [filteredServices]);

  const handleSelectOne = useCallback((name: string, checked: boolean) => {
    setSelected(prev => {
        const newSelected = new Set(prev);
        if (checked) {
          newSelected.add(name);
        } else {
          newSelected.delete(name);
        }
        return newSelected;
    });
  }, []);

  const isAllSelected = filteredServices.length > 0 && selected.size === filteredServices.length;

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
      <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2 w-full md:w-1/3">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Filter by tag..."
              value={tagFilter}
              onChange={(e) => setTagFilter(e.target.value)}
              className="h-8"
            />
          </div>

                   {selected.size > 0 && (
                       <div className="flex items-center gap-2 animate-in fade-in slide-in-from-right-4 duration-300">
                           <span className="text-sm text-muted-foreground mr-2">{selected.size} selected</span>
                           {onBulkToggle && (
                               <>
                                 <Button size="sm" variant="outline" onClick={() => {
                                     onBulkToggle(Array.from(selected), true);
                                     setSelected(new Set());
                                 }}>
                                     <PlayCircle className="mr-2 h-4 w-4 text-green-600" /> Enable
                                 </Button>
                                 <Button size="sm" variant="outline" onClick={() => {
                                     onBulkToggle(Array.from(selected), false);
                                     setSelected(new Set());
                                 }}>
                                     <PauseCircle className="mr-2 h-4 w-4 text-amber-600" /> Disable
                                 </Button>
                               </>
                           )}
                           <Button size="sm" variant="outline" onClick={() => setIsBulkEditDialogOpen(true)}>
                               <Settings className="mr-2 h-4 w-4" /> Bulk Edit
                           </Button>
                           {onBulkDelete && (
                               <Button size="sm" variant="destructive" onClick={() => {
                                   onBulkDelete(Array.from(selected));
                                   setSelected(new Set());
                               }}>
                                   <Trash2 className="mr-2 h-4 w-4" /> Delete
                               </Button>
                           )}
                       </div>
                   )}
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[50px]">
                  <Checkbox
                    checked={isAllSelected}
                    onCheckedChange={(checked) => handleSelectAll(!!checked)}
                    aria-label="Select all"
                  />
              </TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Activity</TableHead>
              <TableHead>Tags</TableHead>
              <TableHead>Address / Command</TableHead>
              <TableHead>Version</TableHead>
              <TableHead className="text-center">Secure</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredServices.map((service) => (
               <ServiceRow
                  key={service.name}
                  service={service}
                  isSelected={selected.has(service.name)}
                  onSelect={handleSelectOne}
                  onToggle={onToggle}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onDuplicate={onDuplicate}
                  onExport={onExport}
                  onLogin={onLogin}
                  onRestart={onRestart}
               />
            ))}
            {filteredServices.length === 0 && (
              <TableRow>
                <TableCell colSpan={10} className="h-24 text-center">
                  No services match the tag filter.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <Dialog open={isBulkEditDialogOpen} onOpenChange={setIsBulkEditDialogOpen}>
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Bulk Edit Services</DialogTitle>
                <DialogDescription>
                    Update {selected.size} selected services. Currently only supports updating tags.
                </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
                <div className="space-y-2">
                    <Label htmlFor="bulk-tags">Add Tags (comma separated)</Label>
                    <Input
                        id="bulk-tags"
                        placeholder="production, web, internal"
                        value={bulkTags}
                        onChange={(e) => setBulkTags(e.target.value)}
                    />
                </div>
                <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                        <Label htmlFor="bulk-timeout">Timeout (optional)</Label>
                        <Input
                            id="bulk-timeout"
                            placeholder="e.g. 30s"
                            value={bulkTimeout}
                            onChange={(e) => setBulkTimeout(e.target.value)}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="bulk-retries">Max Retries (optional)</Label>
                        <Input
                            id="bulk-retries"
                            type="number"
                            placeholder="e.g. 3"
                            value={bulkRetries}
                            onChange={(e) => setBulkRetries(e.target.value)}
                        />
                    </div>
                </div>
                <p className="text-xs text-muted-foreground">Empty fields will be left unchanged.</p>
            </div>
            <DialogFooter>
                <Button variant="outline" onClick={() => setIsBulkEditDialogOpen(false)}>Cancel</Button>
                <Button onClick={() => {
                    if (onBulkEdit) {
                        const resilienceUpdates: any = {};
                        if (bulkTimeout) resilienceUpdates.timeout = bulkTimeout;
                        if (bulkRetries) resilienceUpdates.maxRetries = parseInt(bulkRetries);

                        onBulkEdit(
                            Array.from(selected),
                            {
                                tags: bulkTags ? bulkTags.split(",").map(t => t.trim()).filter(Boolean) : undefined,
                                resilience: Object.keys(resilienceUpdates).length > 0 ? resilienceUpdates : undefined
                            }
                        );
                    }
                    setIsBulkEditDialogOpen(false);
                    setSelected(new Set());
                    setBulkTags("");
                    setBulkTimeout("");
                    setBulkRetries("");
                }}>Apply Changes</Button>
            </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

/**
 * ServiceRow component.
 * @param props - The component props.
 * @param props.service - The service property.
 * @param props.isSelected - The isSelected property.
 * @param props.onSelect - The onSelect property.
 * @param props.onToggle - The onToggle property.
 * @param props.onEdit - The onEdit property.
 * @param props.onDelete - The onDelete property.
 * @param props.onDuplicate - The onDuplicate property.
 * @param props.onExport - The onExport property.
 * @param props.onLogin - The onLogin property.
 * @returns The rendered component.
 */
const ServiceRow = memo(function ServiceRow({ service, isSelected, onSelect, onToggle, onEdit, onDelete, onDuplicate, onExport, onLogin, onRestart }: {
    service: UpstreamServiceConfig,
    isSelected: boolean,
    onSelect: (name: string, checked: boolean) => void,
    onToggle?: (name: string, enabled: boolean) => void,
    onEdit?: (service: UpstreamServiceConfig) => void,
    onDelete?: (name: string) => void,
    onDuplicate?: (service: UpstreamServiceConfig) => void,
    onExport?: (service: UpstreamServiceConfig) => void,
    onLogin?: (service: UpstreamServiceConfig) => void,
    onRestart?: (name: string) => void
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

    const { getServiceHistory } = useServiceHealth();
    const history = getServiceHistory(service.name);
    const latencies = useMemo(() => history.map(h => h.latencyMs), [history]);
    const maxLatency = useMemo(() => Math.max(...latencies, 50), [latencies]); // Minimum max of 50ms for scale

    // Determine color based on latest health
    const healthColor = useMemo(() => {
        if (!history.length) return "#94a3b8"; // slate-400
        const latest = history[history.length - 1];
        if (latest.status === 'NODE_STATUS_ERROR' || latest.errorRate > 0.1) return "#ef4444"; // red-500
        if (latest.latencyMs > 500) return "#eab308"; // yellow-500
        return "#22c55e"; // green-500
    }, [history]);

    return (
        <TableRow className={service.disable ? "opacity-60 bg-muted/40" : ""}>
             <TableCell>
                 <Checkbox
                    checked={isSelected}
                    onCheckedChange={(checked) => onSelect(service.name, !!checked)}
                    aria-label={`Select ${service.name}`}
                 />
             </TableCell>
             <TableCell>
                 <div className="flex items-center gap-2">
                    {onToggle && (
                        <Switch
                            checked={!service.disable}
                            onCheckedChange={(checked) => onToggle(service.name, checked)}
                        />
                    )}
                    {service.lastError && (
                        <ConnectionDiagnosticDialog
                            service={service}
                            trigger={
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-6 w-6 text-destructive hover:text-destructive hover:bg-destructive/10"
                                    title="View Error & Troubleshoot"
                                >
                                    <AlertTriangle className="h-4 w-4" />
                                </Button>
                            }
                        />
                    )}
                 </div>
             </TableCell>
             <TableCell className="font-medium">
                 <div className="flex items-center gap-2">
                     {service.name}
                     {service.lastError && (
                         <Tooltip>
                             <TooltipTrigger asChild>
                                 <Badge variant="destructive" className="ml-2 text-[10px] px-1 h-5 cursor-pointer">Error</Badge>
                             </TooltipTrigger>
                             <TooltipContent>
                                 <p className="max-w-xs break-words text-xs">{service.lastError}</p>
                             </TooltipContent>
                         </Tooltip>
                     )}
                 </div>
             </TableCell>
             <TableCell>
                 <Badge variant="outline">{type}</Badge>
             </TableCell>
             <TableCell>
                <div className="w-[80px] h-[24px]">
                    {!service.disable && (
                        <Sparkline
                            data={latencies}
                            width={80}
                            height={24}
                            color={healthColor}
                            max={maxLatency}
                        />
                    )}
                </div>
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
                 {service.version}
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
                        <ConnectionDiagnosticDialog
                            service={service}
                            trigger={
                                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                                    <Activity className="mr-2 h-4 w-4" />
                                    Diagnose
                                </DropdownMenuItem>
                            }
                        />
                        <DropdownMenuItem asChild>
                            <Link href={`/logs?source=${encodeURIComponent(service.name)}`}>
                                <Terminal className="mr-2 h-4 w-4" />
                                View Logs
                            </Link>
                        </DropdownMenuItem>
                        {onRestart && (
                            <DropdownMenuItem onClick={() => onRestart(service.name)}>
                                <RefreshCw className="mr-2 h-4 w-4" />
                                Restart
                            </DropdownMenuItem>
                        )}
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
                        {onLogin && service.upstreamAuth?.oauth2 && (
                             <DropdownMenuItem onClick={() => onLogin(service)}>
                                <CheckCircle className="mr-2 h-4 w-4" />
                                Log In
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
