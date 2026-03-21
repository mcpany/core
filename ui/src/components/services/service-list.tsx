/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useMemo, useState, memo, useCallback, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Link } from 'react-router-dom';
import { Settings, Trash2, CheckCircle, XCircle, AlertTriangle, MoreHorizontal, Copy, Download, Filter, PlayCircle, PauseCircle, Activity, RefreshCw, Terminal, ShieldCheck, ShieldAlert } from "lucide-react";
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
import { ServiceHealthSparkline } from "@/components/services/service-health-sparkline";
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
  onBulkEdit?: (names: string[], updates: { tags?: string[] }) => void;
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

  const filteredServices = useMemo(() => {
    if (!tagFilter) return services;
    return services.filter(s => s.tags?.some((tag: string) => tag.toLowerCase().includes(tagFilter.toLowerCase())));
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

  const handleBulkActionComplete = useCallback(() => {
    setSelected(new Set());
    setIsBulkEditDialogOpen(false);
    setBulkTags("");
  }, []);

  const isAllSelected = filteredServices.length > 0 && selected.size === filteredServices.length;

  if (isLoading) {
      return (
          <div className="space-y-4">
              <div className="flex justify-between">
                  <div className="h-10 w-64 bg-muted animate-pulse rounded"></div>
              </div>
              <div className="border rounded-md">
                 <div className="h-12 border-b bg-muted/50 animate-pulse"></div>
                 <div className="h-16 border-b animate-pulse"></div>
                 <div className="h-16 border-b animate-pulse"></div>
                 <div className="h-16 animate-pulse"></div>
              </div>
          </div>
      );
  }

  if (!services || services.length === 0) {
      return (
          <div className="text-center p-8 border border-dashed rounded-lg bg-muted/10">
              <p className="text-muted-foreground">No upstream services registered yet.</p>
              <p className="text-sm text-muted-foreground mt-2">
                 Register a service to connect MCP Any to your backend.
              </p>
          </div>
      );
  }

  return (
    <div className="space-y-4 relative pb-16">
        {/* Bulk Edit Dialog */}
        <Dialog open={isBulkEditDialogOpen} onOpenChange={setIsBulkEditDialogOpen}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Edit Multiple Services</DialogTitle>
                    <DialogDescription>
                        Update settings for {selected.size} selected services.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid gap-2">
                        <Label htmlFor="bulkTags">Add/Replace Tags (comma separated)</Label>
                        <Input
                            id="bulkTags"
                            placeholder="production, experimental"
                            value={bulkTags}
                            onChange={(e) => setBulkTags(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">This will replace existing tags for all selected services.</p>
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => setIsBulkEditDialogOpen(false)}>Cancel</Button>
                    <Button onClick={() => {
                        if (onBulkEdit) {
                            const newTags = bulkTags.split(',').map(t => t.trim()).filter(Boolean);
                            onBulkEdit(Array.from(selected), { tags: newTags });
                        }
                        handleBulkActionComplete();
                    }}>Apply Changes</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>

        {/* Floating Bulk Actions Bar */}
        {selected.size > 0 && (
            <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-4 px-6 py-3 bg-background/80 backdrop-blur-lg border shadow-lg rounded-full animate-in slide-in-from-bottom-4 fade-in duration-200">
                <span className="text-sm font-medium whitespace-nowrap">
                    {selected.size} selected
                </span>
                <div className="w-px h-6 bg-border mx-2"></div>

                {onBulkToggle && (
                    <>
                        <Button
                            variant="ghost"
                            size="sm"
                            className="text-green-600 hover:text-green-700 hover:bg-green-50"
                            onClick={() => {
                                onBulkToggle(Array.from(selected), true);
                                handleBulkActionComplete();
                            }}
                        >
                            <PlayCircle className="w-4 h-4 mr-2" />
                            Enable
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            className="text-amber-600 hover:text-amber-700 hover:bg-amber-50"
                            onClick={() => {
                                onBulkToggle(Array.from(selected), false);
                                handleBulkActionComplete();
                            }}
                        >
                            <PauseCircle className="w-4 h-4 mr-2" />
                            Disable
                        </Button>
                    </>
                )}

                {onBulkEdit && (
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setIsBulkEditDialogOpen(true)}
                    >
                        <Settings className="w-4 h-4 mr-2" />
                        Edit Tags
                    </Button>
                )}

                {onBulkDelete && (
                    <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive hover:bg-destructive/10"
                        onClick={() => {
                            if (window.confirm(`Are you sure you want to delete ${selected.size} services? This action cannot be undone.`)) {
                                onBulkDelete(Array.from(selected));
                                handleBulkActionComplete();
                            }
                        }}
                    >
                        <Trash2 className="w-4 h-4 mr-2" />
                        Delete
                    </Button>
                )}

                <Button
                    variant="ghost"
                    size="icon"
                    className="ml-2 rounded-full h-8 w-8 hover:bg-muted"
                    onClick={() => setSelected(new Set())}
                    title="Clear selection"
                >
                    <XCircle className="w-4 h-4" />
                </Button>
            </div>
        )}

        <div className="flex gap-2">
           <div className="relative max-w-sm flex-1">
                <Filter className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Filter by tag..."
                    value={tagFilter}
                    onChange={(e) => setTagFilter(e.target.value)}
                    className="pl-8"
                />
            </div>
            {tagFilter && (
                <Button variant="ghost" onClick={() => setTagFilter("")}>Clear Filter</Button>
            )}
        </div>

      <div className="rounded-md border bg-card">
          <Table>
              <TableHeader>
                  <TableRow>
                      <TableHead className="w-[40px]">
                          <Checkbox
                              checked={isAllSelected}
                              onCheckedChange={handleSelectAll}
                              aria-label="Select all"
                          />
                      </TableHead>
                      <TableHead className="w-[100px]">Status</TableHead>
                      <TableHead className="w-[200px]">Name</TableHead>
                      <TableHead className="w-[120px]">Security</TableHead>
                      <TableHead className="w-[100px]">Type</TableHead>
                      <TableHead className="w-[120px]">Health (5m)</TableHead>
                      <TableHead className="w-[150px]">Tags</TableHead>
                      <TableHead>Endpoint</TableHead>
                      <TableHead className="w-[100px]">Version</TableHead>
                      <TableHead className="w-[80px] text-center">TLS</TableHead>
                      <TableHead className="text-right w-[80px]">Actions</TableHead>
                  </TableRow>
              </TableHeader>
              <TableBody>
                  {filteredServices.length === 0 ? (
                      <TableRow>
                          <TableCell colSpan={11} className="text-center py-6 text-muted-foreground">
                              {tagFilter ? "No services match the filter." : "No services found."}
                          </TableCell>
                      </TableRow>
                  ) : (
                      filteredServices.map((service) => (
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
                      ))
                  )}
              </TableBody>
          </Table>
      </div>
    </div>
  );
}

/**
 * ServiceRow
 *
 * @param props
 * @param props.service
 * @param props.isSelected
 * @param props.onSelect
 * @param props.onToggle
 * @param props.onEdit
 * @param props.onDelete
 * @param props.onDuplicate
 * @param props.onExport
 * @param props.onLogin
 * @param props.onRestart
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
                     <Link to={`/upstream-services/${service.name}`} className="hover:underline font-semibold text-primary">
                        {service.name}
                     </Link>
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
                  {service.provenance?.verified ? (
                      <Tooltip>
                          <TooltipTrigger>
                              <div className="flex items-center gap-1 text-green-600">
                                  <ShieldCheck className="h-4 w-4" />
                                  <span className="text-xs font-medium">Verified</span>
                              </div>
                          </TooltipTrigger>
                          <TooltipContent>
                              <p className="font-semibold">Verified Source</p>
                              <p className="text-xs text-muted-foreground">Signer: {service.provenance.signerIdentity}</p>
                          </TooltipContent>
                      </Tooltip>
                  ) : (
                      <Tooltip>
                          <TooltipTrigger>
                              <div className="flex items-center gap-1 text-muted-foreground opacity-50">
                                  <ShieldAlert className="h-4 w-4" />
                                  <span className="text-xs">Unverified</span>
                              </div>
                          </TooltipTrigger>
                          <TooltipContent>
                              <p>Unverified Source</p>
                          </TooltipContent>
                      </Tooltip>
                  )}
             </TableCell>
             <TableCell>
                 <Badge variant="outline">{type}</Badge>
             </TableCell>
             <TableCell>
                <div className="w-[80px] h-[24px]">
                    <ServiceHealthSparkline serviceName={service.name} disabled={service.disable} />
                </div>
             </TableCell>
             <TableCell>
                 <div className="flex flex-wrap gap-1">
                     {service.tags?.map((tag: string) => (
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
                            <Link to={`/logs?source=${encodeURIComponent(service.name)}`}>
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
