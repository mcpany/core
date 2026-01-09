"use client";

import { useState, useMemo } from "react";
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
import { CheckCircle2, AlertCircle, AlertTriangle, Search, Filter, MoreHorizontal, Clock, RefreshCw, Activity } from "lucide-react";
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

const MOCK_ALERTS: Alert[] = [
  { id: "AL-1024", title: "High CPU Usage", message: "CPU usage > 90% for 5m", severity: "critical", status: "active", service: "weather-service", source: "System Monitor", timestamp: new Date(Date.now() - 1000 * 60 * 5).toISOString() },
  { id: "AL-1023", title: "API Latency Spike", message: "P99 Latency > 2000ms", severity: "warning", status: "active", service: "api-gateway", source: "Latency Watchdog", timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString() },
  { id: "AL-1022", title: "Disk Space Low", message: "Volume /data at 85%", severity: "warning", status: "acknowledged", service: "database-primary", source: "Disk Monitor", timestamp: new Date(Date.now() - 1000 * 60 * 45).toISOString() },
  { id: "AL-1021", title: "Connection Refused", message: "Upstream connection failed", severity: "critical", status: "resolved", service: "payment-provider", source: "Connectivity Check", timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString() },
  { id: "AL-1020", title: "New Service Deployed", message: "Service 'search-v2' detected", severity: "info", status: "resolved", service: "discovery", source: "Orchestrator", timestamp: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString() },
];

export function AlertList() {
  const [alerts, setAlerts] = useState<Alert[]>(MOCK_ALERTS);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterSeverity, setFilterSeverity] = useState<string>("all");
  const [filterStatus, setFilterStatus] = useState<string>("all");

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

  const handleStatusChange = (id: string, newStatus: AlertStatus) => {
    setAlerts(prev => prev.map(a => a.id === id ? { ...a, status: newStatus } : a));
  };

  const getSeverityBadge = (severity: Severity) => {
    switch (severity) {
      case "critical": return <Badge variant="destructive" className="uppercase text-[10px]">Critical</Badge>;
      case "warning": return <Badge variant="secondary" className="bg-yellow-500/15 text-yellow-600 dark:text-yellow-400 hover:bg-yellow-500/25 uppercase text-[10px]">Warning</Badge>;
      case "info": return <Badge variant="outline" className="text-blue-500 border-blue-200 dark:border-blue-800 uppercase text-[10px]">Info</Badge>;
    }
  };

  const getStatusIcon = (status: AlertStatus) => {
    switch (status) {
      case "active": return <AlertCircle className="h-4 w-4 text-red-500 animate-pulse" />;
      case "acknowledged": return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case "resolved": return <CheckCircle2 className="h-4 w-4 text-green-500" />;
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
          <Button variant="outline" size="icon" onClick={() => setAlerts(MOCK_ALERTS)}>
             <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px]">Severity</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
              <TableHead>Summary</TableHead>
              <TableHead className="hidden md:table-cell">Service</TableHead>
              <TableHead className="hidden md:table-cell">Time</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredAlerts.length === 0 ? (
                <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                        No alerts match your filters.
                    </TableCell>
                </TableRow>
            ) : (
                filteredAlerts.map((alert) => (
                <TableRow key={alert.id} className="group">
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
