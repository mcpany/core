/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useMemo, useEffect, useCallback } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import { CheckCircle2, AlertCircle, AlertTriangle, Search, Filter, MoreHorizontal, Clock, RefreshCw, Activity, Loader2, PlayCircle, PauseCircle, Trash2 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Alert, Severity, AlertStatus } from "./types";
import { formatDistanceToNow } from "date-fns";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

/**
 * AlertList component.
 * @returns The rendered component.
 */
export function AlertList() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterSeverity, setFilterSeverity] = useState<string>("all");
  const [filterStatus, setFilterStatus] = useState<string>("all");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const { toast } = useToast();

  // Reset selection when alerts list changes (e.g. filtering or reloading)
  useEffect(() => {
    setSelected(new Set());
  }, [alerts]);

  const handleSelectAll = useCallback((checked: boolean, filteredAlerts: Alert[]) => {
    if (checked) {
      setSelected(new Set(filteredAlerts.map(a => a.id)));
    } else {
      setSelected(new Set());
    }
  }, []);

  const handleSelectOne = useCallback((id: string, checked: boolean) => {
    setSelected(prev => {
        const newSelected = new Set(prev);
        if (checked) {
          newSelected.add(id);
        } else {
          newSelected.delete(id);
        }
        return newSelected;
    });
  }, []);

  const handleBulkStatusChange = async (newStatus: AlertStatus) => {
      const selectedIds = Array.from(selected);
      try {
          await Promise.all(selectedIds.map(id => apiClient.updateAlertStatus(id, newStatus)));

          setAlerts(prev => prev.map(a => selectedIds.includes(a.id) ? { ...a, status: newStatus } : a));

          toast({
              title: "Alerts Updated",
              description: `${selectedIds.length} alert(s) marked as ${newStatus}.`
          });
          setSelected(new Set());
      } catch (e) {
          console.error("Failed to bulk update alerts", e);
          fetchAlerts(); // Revert
          toast({
              variant: "destructive",
              title: "Error",
              description: "Failed to update some alerts."
          });
      }
  };

  const handleBulkDelete = async () => {
      const selectedIds = Array.from(selected);
      try {
          await Promise.all(selectedIds.map(id => apiClient.deleteAlert(id)));

          setAlerts(prev => prev.filter(a => !selectedIds.includes(a.id)));

          toast({
              title: "Alerts Deleted",
              description: `${selectedIds.length} alert(s) deleted.`
          });
          setSelected(new Set());
      } catch (e) {
          console.error("Failed to bulk delete alerts", e);
          fetchAlerts(); // Revert
          toast({
              variant: "destructive",
              title: "Error",
              description: "Failed to delete some alerts."
          });
      }
  };

  const handleDelete = async (id: string) => {
    try {
        await apiClient.deleteAlert(id);
        setAlerts(prev => prev.filter(a => a.id !== id));
        toast({
            title: "Alert Deleted",
            description: `Alert has been deleted.`,
        });
    } catch (error) {
        console.error(error);
        toast({
            title: "Error",
            description: "Failed to delete alert",
            variant: "destructive",
        });
    }
  };

  const fetchAlerts = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listAlerts();
      setAlerts(data);
    } catch (error) {
      console.error(error);
      toast({
        title: "Error",
        description: "Failed to load alerts",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAlerts();
  }, []);

  const filteredAlerts = useMemo(() => {
    return alerts.filter(alert => {
      const matchesSearch =
        alert.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        alert.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
        alert.service.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesSeverity = filterSeverity === "all" || alert.severity === filterSeverity;
      const matchesStatus = filterStatus === "all" || alert.status === filterStatus;

      return matchesSearch && matchesSeverity && matchesStatus;
    });
  }, [alerts, searchQuery, filterSeverity, filterStatus]);

  const handleStatusChange = async (id: string, newStatus: AlertStatus) => {
    try {
        const updated = await apiClient.updateAlertStatus(id, newStatus);
        setAlerts(prev => prev.map(a => a.id === id ? updated : a));
        toast({
            title: "Status Updated",
            description: `Alert marked as ${newStatus}`,
        });
    } catch (error) {
        console.error(error);
        toast({
            title: "Error",
            description: "Failed to update status",
            variant: "destructive",
        });
    }
  };

  const getSeverityBadge = (severity: Severity) => {
    switch (severity) {
      case "critical": return <Badge variant="destructive" className="uppercase text-[10px]">Critical</Badge>;
      case "warning": return <Badge variant="secondary" className="bg-yellow-500/15 text-yellow-600 dark:text-yellow-400 hover:bg-yellow-500/25 uppercase text-[10px]">Warning</Badge>;
      case "info": return <Badge variant="outline" className="text-blue-500 border-blue-200 dark:border-blue-800 uppercase text-[10px]">Info</Badge>;
      default: return <Badge variant="outline" className="uppercase text-[10px]">{severity}</Badge>;
    }
  };

  const getStatusIcon = (status: AlertStatus) => {
    switch (status) {
      case "active": return <AlertCircle className="h-4 w-4 text-red-500 animate-pulse" />;
      case "acknowledged": return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case "resolved": return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      default: return <Activity className="h-4 w-4 text-muted-foreground" />;
    }
  };

  const isAllSelected = filteredAlerts.length > 0 && selected.size === filteredAlerts.length;

  return (
    <div className="space-y-4">
      {selected.size > 0 && (
          <div className="flex items-center gap-2 p-2 bg-muted/40 rounded-md animate-in fade-in slide-in-from-top-1 duration-200 sticky top-0 z-10 backdrop-blur-md border">
              <span className="text-sm text-muted-foreground mr-2 font-medium px-2">{selected.size} selected</span>
              <div className="h-4 w-px bg-border mx-1" />
              <Button size="sm" variant="ghost" onClick={() => handleBulkStatusChange('acknowledged')} className="h-8 text-yellow-600 hover:text-yellow-700 hover:bg-yellow-100 dark:hover:bg-yellow-900/20">
                  <AlertTriangle className="mr-2 h-4 w-4" /> Acknowledge
              </Button>
              <Button size="sm" variant="ghost" onClick={() => handleBulkStatusChange('resolved')} className="h-8 text-green-600 hover:text-green-700 hover:bg-green-100 dark:hover:bg-green-900/20">
                  <CheckCircle2 className="mr-2 h-4 w-4" /> Resolve
              </Button>
              <Button size="sm" variant="ghost" onClick={() => handleBulkDelete()} className="h-8 text-red-600 hover:text-red-700 hover:bg-red-100 dark:hover:bg-red-900/20">
                  <Trash2 className="mr-2 h-4 w-4" /> Delete
              </Button>
          </div>
      )}

      <div className="flex flex-col sm:flex-row gap-4 justify-between items-center">
        <div className="relative w-full sm:w-96">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search alerts by title, message, service..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8"
          />
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto">
           <Select value={filterSeverity} onValueChange={setFilterSeverity}>
            <SelectTrigger className="w-[130px]">
              <div className="flex items-center gap-2">
                 <Filter className="h-3.5 w-3.5 text-muted-foreground" />
                 <SelectValue placeholder="Severity" />
              </div>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Severities</SelectItem>
              <SelectItem value="critical">Critical</SelectItem>
              <SelectItem value="warning">Warning</SelectItem>
              <SelectItem value="info">Info</SelectItem>
            </SelectContent>
          </Select>
          <Select value={filterStatus} onValueChange={setFilterStatus}>
             <SelectTrigger className="w-[130px]">
              <div className="flex items-center gap-2">
                 <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                 <SelectValue placeholder="Status" />
              </div>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="acknowledged">Acknowledged</SelectItem>
              <SelectItem value="resolved">Resolved</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="icon" onClick={fetchAlerts} disabled={loading}>
             <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[30px] pr-0">
                 <Checkbox
                    checked={isAllSelected}
                    onCheckedChange={(checked) => handleSelectAll(!!checked, filteredAlerts)}
                    aria-label="Select all"
                    className="translate-y-[2px]"
                  />
              </TableHead>
              <TableHead className="w-[100px]">Severity</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
              <TableHead>Summary</TableHead>
              <TableHead className="hidden md:table-cell">Service</TableHead>
              <TableHead className="hidden md:table-cell">Time</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && alerts.length === 0 ? (
                 <TableRow>
                    <TableCell colSpan={7} className="h-24 text-center text-muted-foreground">
                        <div className="flex items-center justify-center gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Loading alerts...
                        </div>
                    </TableCell>
                </TableRow>
            ) : filteredAlerts.length === 0 ? (
                <TableRow>
                    <TableCell colSpan={7} className="h-24 text-center text-muted-foreground">
                        No alerts match your filters.
                    </TableCell>
                </TableRow>
            ) : (
                filteredAlerts.map((alert) => (
                <TableRow key={alert.id} className={cn("group", selected.has(alert.id) ? "bg-muted/50" : "")}>
                    <TableCell className="pr-0">
                       <Checkbox
                          checked={selected.has(alert.id)}
                          onCheckedChange={(checked) => handleSelectOne(alert.id, !!checked)}
                          aria-label={`Select ${alert.id}`}
                          className="translate-y-[2px]"
                       />
                    </TableCell>
                    <TableCell>{getSeverityBadge(alert.severity)}</TableCell>
                    <TableCell>
                    <div className="flex items-center gap-2" title={alert.status}>
                        {getStatusIcon(alert.status)}
                        <span className="capitalize text-xs hidden sm:inline">{alert.status}</span>
                    </div>
                    </TableCell>
                    <TableCell>
                    <div className="flex flex-col">
                        <span className="font-medium text-sm">{alert.title}</span>
                        <span className="text-xs text-muted-foreground">{alert.message}</span>
                    </div>
                    </TableCell>
                    <TableCell className="hidden md:table-cell">
                        <Badge variant="outline" className="font-mono text-xs">{alert.service}</Badge>
                    </TableCell>
                    <TableCell className="hidden md:table-cell text-xs text-muted-foreground whitespace-nowrap">
                        <div className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {formatDistanceToNow(new Date(alert.timestamp), { addSuffix: true })}
                        </div>
                    </TableCell>
                    <TableCell>
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                            <span className="sr-only">Open menu</span>
                            <MoreHorizontal className="h-4 w-4" />
                        </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                        <DropdownMenuItem onClick={() => navigator.clipboard.writeText(alert.id)}>
                            Copy Alert ID
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={() => handleStatusChange(alert.id, 'acknowledged')} disabled={alert.status !== 'active'}>
                            Acknowledge
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleStatusChange(alert.id, 'resolved')} disabled={alert.status === 'resolved'}>
                            Resolve
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleDelete(alert.id)} className="text-red-600 focus:text-red-600">
                            Delete
                        </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                    </TableCell>
                </TableRow>
                ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
