/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
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
import { CheckCircle2, AlertCircle, AlertTriangle, Search, Filter, MoreHorizontal, Clock, RefreshCw, Activity, Loader2, CheckSquare } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
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
  const [selectedAlertIds, setSelectedAlertIds] = useState<Set<string>>(new Set());
  const [isBulkUpdating, setIsBulkUpdating] = useState(false);
  const { toast } = useToast();

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

  const handleBulkStatusChange = async (newStatus: AlertStatus) => {
    if (selectedAlertIds.size === 0) return;
    setIsBulkUpdating(true);
    try {
      const ids = Array.from(selectedAlertIds);
      const updatedAlerts = await Promise.all(
        ids.map(id => apiClient.updateAlertStatus(id, newStatus))
      );

      setAlerts(prev => prev.map(a => {
        const updated = updatedAlerts.find(u => u.id === a.id);
        return updated ? updated : a;
      }));

      setSelectedAlertIds(new Set());
      toast({
        title: "Bulk Update Successful",
        description: `${ids.length} alerts marked as ${newStatus}`,
      });
    } catch (error) {
      console.error("Bulk update failed:", error);
      toast({
        title: "Bulk Update Failed",
        description: "An error occurred while updating alerts.",
        variant: "destructive",
      });
      // Refresh to ensure consistent state if partial failure
      fetchAlerts();
    } finally {
      setIsBulkUpdating(false);
    }
  };

  const toggleSelectAll = () => {
    if (selectedAlertIds.size === filteredAlerts.length && filteredAlerts.length > 0) {
      setSelectedAlertIds(new Set());
    } else {
      setSelectedAlertIds(new Set(filteredAlerts.map(a => a.id)));
    }
  };

  const toggleSelectAlert = (id: string) => {
    const newSelected = new Set(selectedAlertIds);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedAlertIds(newSelected);
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

  return (
    <div className="space-y-4">
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

      {selectedAlertIds.size > 0 && (
        <div className="flex items-center justify-between p-3 bg-muted/30 border rounded-md shadow-sm mb-4 animate-in fade-in slide-in-from-top-2">
          <div className="text-sm font-medium flex items-center gap-2">
            <CheckSquare className="h-4 w-4 text-primary" />
            <span>{selectedAlertIds.size} alert{selectedAlertIds.size > 1 ? 's' : ''} selected</span>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleBulkStatusChange('acknowledged')}
              disabled={isBulkUpdating}
              className="text-yellow-600 hover:text-yellow-700 hover:bg-yellow-50 dark:hover:bg-yellow-950/20"
            >
              <AlertTriangle className="mr-2 h-4 w-4" />
              Acknowledge
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleBulkStatusChange('resolved')}
              disabled={isBulkUpdating}
              className="text-green-600 hover:text-green-700 hover:bg-green-50 dark:hover:bg-green-950/20"
            >
              <CheckCircle2 className="mr-2 h-4 w-4" />
              Resolve
            </Button>
          </div>
        </div>
      )}

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[40px] text-center">
                <Checkbox
                  checked={filteredAlerts.length > 0 && selectedAlertIds.size === filteredAlerts.length}
                  onCheckedChange={toggleSelectAll}
                  aria-label="Select all alerts"
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
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                        <div className="flex items-center justify-center gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Loading alerts...
                        </div>
                    </TableCell>
                </TableRow>
            ) : filteredAlerts.length === 0 ? (
                <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                        No alerts match your filters.
                    </TableCell>
                </TableRow>
            ) : (
                filteredAlerts.map((alert) => (
                <TableRow key={alert.id} className="group" data-state={selectedAlertIds.has(alert.id) ? "selected" : undefined}>
                    <TableCell className="text-center">
                      <Checkbox
                        checked={selectedAlertIds.has(alert.id)}
                        onCheckedChange={() => toggleSelectAlert(alert.id)}
                        aria-label={`Select alert ${alert.id}`}
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
