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
import { Settings, Trash2, CheckCircle, XCircle, AlertTriangle, MoreHorizontal, Copy, Download, Filter, PlayCircle, PauseCircle, Activity, RefreshCw, Terminal, ShieldCheck, ShieldAlert, LayoutGrid, List } from "lucide-react";
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
import { Card, CardHeader, CardContent, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { cn } from "@/lib/utils";


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
 * ServiceList executes the ServiceList logic.
 *
 * Summary: Executes the ServiceList logic.
 *
 * @param { services - The { services parameter.
 * @param isLoading - The isLoading parameter.
 * @param onToggle - The onToggle parameter.
 * @param onEdit - The onEdit parameter.
 * @param onDelete - The onDelete parameter.
 * @param onDuplicate - The onDuplicate parameter.
 * @param onExport - The onExport parameter.
 * @param onBulkToggle - The onBulkToggle parameter.
 * @param onBulkDelete - The onBulkDelete parameter.
 * @param onLogin - The onLogin parameter.
 * @param onRestart - The onRestart parameter.
 * @param onBulkEdit } - The onBulkEdit } parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function ServiceList({ services, isLoading, onToggle, onEdit, onDelete, onDuplicate, onExport, onBulkToggle, onBulkDelete, onLogin, onRestart, onBulkEdit }: ServiceListProps) {
  const [tagFilter, setTagFilter] = useState("");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [viewMode, setViewMode] = useState<"table" | "grid">(() => {
    return (localStorage.getItem("service_list_view_mode") as "table" | "grid") || "table";
  });
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

                   <div className="flex items-center gap-2 mr-2 border rounded-md p-1">
                       <Button
                           variant={viewMode === "table" ? "secondary" : "ghost"}
                           size="sm"
                           className="h-7 px-2"
                           onClick={() => {
                               setViewMode("table");
                               localStorage.setItem("service_list_view_mode", "table");
                           }}
                       >
                           <List className="h-4 w-4" />
                       </Button>
                       <Button
                           variant={viewMode === "grid" ? "secondary" : "ghost"}
                           size="sm"
                           className="h-7 px-2"
                           onClick={() => {
                               setViewMode("grid");
                               localStorage.setItem("service_list_view_mode", "grid");
                           }}
                       >
                           <LayoutGrid className="h-4 w-4" />
                       </Button>
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

      {viewMode === "table" ? (
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
                <TableHead>Trust</TableHead>
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
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredServices.map((service) => (
                <ServiceCard
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
                <div className="col-span-full h-24 flex items-center justify-center border rounded-md border-dashed text-muted-foreground">
                    No services match the tag filter.
                </div>
            )}
        </div>
      )}
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
            </div>
            <DialogFooter>
                <Button variant="outline" onClick={() => setIsBulkEditDialogOpen(false)}>Cancel</Button>
                <Button onClick={() => {
                    if (onBulkEdit) {
                        onBulkEdit(Array.from(selected), { tags: bulkTags.split(",").map(t => t.trim()).filter(Boolean) });
                    }
                    setIsBulkEditDialogOpen(false);
                    setSelected(new Set());
                    setBulkTags("");
                }}>Apply Changes</Button>
            </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

/**
 * ServiceCard component for Grid View.
 * @param props - The component props.
 * @returns The rendered component.
 */
const ServiceCard = memo(function ServiceCard({ service, isSelected, onSelect, onToggle, onEdit, onDelete, onDuplicate, onExport, onLogin, onRestart }: {
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
        <Card className={cn("relative flex flex-col backdrop-blur-sm transition-all duration-200 border-border/50", service.disable ? "opacity-60 bg-muted/40" : "bg-card/50 shadow-sm hover:shadow-md", isSelected ? "border-primary shadow-md ring-1 ring-primary" : "")}>
            <div className="absolute top-3 left-3 z-10">
                 <Checkbox
                    checked={isSelected}
                    onCheckedChange={(checked) => onSelect(service.name, !!checked)}
                    aria-label={`Select ${service.name}`}
                 />
            </div>

            <div className="absolute top-3 right-3 flex items-center gap-1 z-10">
                {onToggle && (
                    <Switch
                        checked={!service.disable}
                        onCheckedChange={(checked) => onToggle(service.name, checked)}
                        className="scale-75 origin-right"
                    />
                )}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0 hover:bg-muted/80">
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
            </div>

            <CardHeader className="pt-6 pb-2 px-4 text-center">
                 <div className="flex justify-center mb-2">
                    <div className={cn("h-3 w-3 rounded-full shadow-sm", service.disable ? "bg-slate-400 border border-slate-600" : service.lastError ? "bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)] animate-pulse" : "bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]")} />
                 </div>
                 <CardTitle className="text-lg font-bold flex items-center justify-center gap-2">
                     <Link to={`/upstream-services/${service.name}`} className="hover:underline text-primary truncate max-w-[200px]" title={service.name}>
                        {service.name}
                     </Link>
                 </CardTitle>
                 <CardDescription className="text-xs">
                     {service.version || "1.0.0"}
                 </CardDescription>
            </CardHeader>
            <CardContent className="px-4 py-2 flex-1 flex flex-col gap-3">
                 <div className="h-[40px] w-full flex items-center justify-center border-y border-dashed py-2 border-border/50 bg-background/30 rounded-md">
                     <div className="w-full h-full max-w-[120px]">
                        <ServiceHealthSparkline serviceName={service.name} disabled={service.disable} />
                     </div>
                 </div>

                 <div className="grid grid-cols-2 gap-2 text-xs">
                     <div className="flex flex-col gap-1 p-2 bg-muted/30 rounded border border-border/50">
                         <span className="text-muted-foreground uppercase tracking-wider text-[10px] font-semibold">Type</span>
                         <span className="font-medium flex items-center gap-1">
                             {type}
                             {secure && <CheckCircle className="h-3 w-3 text-green-500 ml-auto" />}
                         </span>
                     </div>
                     <div className="flex flex-col gap-1 p-2 bg-muted/30 rounded border border-border/50">
                         <span className="text-muted-foreground uppercase tracking-wider text-[10px] font-semibold">Trust</span>
                         <span className="font-medium flex items-center gap-1">
                            {service.provenance?.verified ? (
                                <><ShieldCheck className="h-3 w-3 text-green-600" /> Verified</>
                            ) : (
                                <><ShieldAlert className="h-3 w-3 text-muted-foreground opacity-50" /> Unverified</>
                            )}
                         </span>
                     </div>
                     <div className="col-span-2 flex flex-col gap-1 p-2 bg-muted/30 rounded border border-border/50">
                         <span className="text-muted-foreground uppercase tracking-wider text-[10px] font-semibold">Address / Command</span>
                         <span className="font-mono text-[10px] truncate" title={address}>{address}</span>
                     </div>
                 </div>

                 {service.lastError && (
                     <div className="mt-auto p-2 bg-red-500/10 border border-red-500/20 rounded-md flex items-start gap-2">
                         <AlertTriangle className="h-3 w-3 text-red-500 shrink-0 mt-0.5" />
                         <div className="text-[10px] text-red-600 dark:text-red-400 line-clamp-2" title={service.lastError}>
                             {service.lastError}
                         </div>
                     </div>
                 )}
            </CardContent>
            {service.tags && service.tags.length > 0 && (
                <CardFooter className="px-4 py-3 bg-muted/20 border-t flex flex-wrap gap-1 mt-auto">
                    {service.tags.map((tag: string) => (
                        <Badge key={tag} variant="secondary" className="text-[10px] px-1.5 py-0">
                            {tag}
                        </Badge>
                    ))}
                </CardFooter>
            )}
        </Card>
    );
});

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
