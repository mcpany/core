/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useEffect, useMemo } from "react";
import { apiClient, DoctorReport, SystemStatus } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Activity,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  RefreshCw,
  Server,
  Cpu,
  Globe,
  Loader2,
  Clock,
  ShieldCheck,
  ShieldAlert,
  Network
} from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * SystemHealth component.
 * Displays overall security posture, system status, and diagnostics.
 * @returns The rendered component.
 */
export function SystemHealth() {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [sysStatus, setSysStatus] = useState<SystemStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchHealth = async () => {
    setLoading(true);
    setError(null);
    try {
      const [doctorData, statusData] = await Promise.all([
         apiClient.getDoctorStatus(),
         apiClient.getSystemStatus()
      ]);
      setReport(doctorData);
      setSysStatus(statusData);
    } catch (err) {
      console.error("Failed to fetch system health", err);
      setError("Failed to retrieve diagnostics report. The backend might be unreachable.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
  }, []);

  const getStatusBadge = (status: string) => {
    switch (status.toLowerCase()) {
      case "ok":
      case "healthy":
        return <Badge variant="default" className="bg-green-600 hover:bg-green-700 uppercase tracking-wider text-[10px]">Healthy</Badge>;
      case "degraded":
      case "warning":
        return <Badge variant="secondary" className="bg-amber-500/10 text-amber-600 hover:bg-amber-500/20 uppercase tracking-wider text-[10px]">Degraded</Badge>;
      case "error":
      case "unhealthy":
      case "critical":
        return <Badge variant="destructive" className="uppercase tracking-wider text-[10px]">Critical</Badge>;
      default:
        return <Badge variant="outline" className="uppercase tracking-wider text-[10px]">Unknown</Badge>;
    }
  };

  const getIconForCheck = (name: string) => {
    const n = name.toLowerCase();
    if (n.includes("network") || n.includes("connectivity") || n.includes("internet")) return <Globe className="h-4 w-4" />;
    if (n.includes("database") || n.includes("storage")) return <Server className="h-4 w-4" />;
    if (n.includes("memory") || n.includes("cpu") || n.includes("runtime")) return <Cpu className="h-4 w-4" />;
    if (n.includes("security") || n.includes("auth")) return <ShieldCheck className="h-4 w-4" />;
    return <Activity className="h-4 w-4" />;
  };

  const securityScore = useMemo(() => {
      let score = 100;
      if (sysStatus?.security_warnings?.length) {
          score -= sysStatus.security_warnings.length * 20;
      }
      if (report?.status === 'degraded') score -= 10;
      if (report?.status === 'error') score -= 30;
      return Math.max(score, 0);
  }, [sysStatus, report]);

  if (loading && !report) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="text-muted-foreground text-sm">Evaluating system integrity...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <Alert variant="destructive" className="max-w-md shadow-lg border-red-500/30">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Diagnostics Failed</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button onClick={fetchHealth} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry Connection
        </Button>
      </div>
    );
  }

  const isSecure = securityScore >= 80;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
         {/* Security Posture Overview */}
         <Card className={cn(
             "col-span-1 md:col-span-2 overflow-hidden border shadow-sm backdrop-blur-xl bg-background/60",
             isSecure ? "border-green-500/20" : "border-amber-500/30"
         )}>
             <div className="flex flex-col md:flex-row h-full">
                 <div className={cn(
                     "p-6 flex flex-col items-center justify-center border-b md:border-b-0 md:border-r min-w-[200px]",
                     isSecure ? "bg-green-500/5" : "bg-amber-500/5"
                 )}>
                     <div className="relative">
                         {isSecure ? (
                             <ShieldCheck className="h-16 w-16 text-green-500" />
                         ) : (
                             <ShieldAlert className="h-16 w-16 text-amber-500" />
                         )}
                         {/* Animated pulse ring */}
                         <div className={cn(
                             "absolute inset-0 rounded-full animate-ping opacity-20",
                             isSecure ? "bg-green-500" : "bg-amber-500"
                         )} />
                     </div>
                     <h3 className="mt-4 text-3xl font-bold tracking-tight">{securityScore}/100</h3>
                     <p className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mt-1">Security Score</p>
                 </div>
                 <div className="p-6 flex-1 flex flex-col justify-between">
                     <div>
                         <h3 className="text-lg font-semibold flex items-center gap-2">
                             {isSecure ? "Network is Secure" : "Action Recommended"}
                             {isSecure && <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-500/20">Protected</Badge>}
                         </h3>
                         <p className="text-sm text-muted-foreground mt-1">
                             {sysStatus?.security_warnings?.length === 0
                               ? "Your MCP Any node is running with recommended security configurations. Connections are authenticated or isolated to local interfaces."
                               : "We detected some potential security risks in your environment."}
                         </p>

                         {sysStatus?.security_warnings && sysStatus.security_warnings.length > 0 && (
                             <div className="mt-4 space-y-2">
                                 {sysStatus.security_warnings.map((warn, i) => (
                                     <div key={i} className="flex items-start gap-2 text-sm bg-amber-500/10 text-amber-700 dark:text-amber-400 p-2 rounded-md border border-amber-500/20">
                                         <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
                                         <p>{warn}</p>
                                     </div>
                                 ))}
                             </div>
                         )}
                     </div>
                     <div className="flex items-center gap-4 mt-6">
                         <Button onClick={fetchHealth} disabled={loading} size="sm" variant="outline" className="h-8">
                            <RefreshCw className={cn("mr-2 h-3.5 w-3.5", loading && "animate-spin")} />
                            Re-evaluate
                         </Button>
                         <p className="text-xs text-muted-foreground flex items-center gap-1.5">
                            <Clock className="h-3 w-3" /> Last verified: {new Date().toLocaleTimeString()}
                         </p>
                     </div>
                 </div>
             </div>
         </Card>

         {/* Network Exposure Card */}
         <Card className="col-span-1 shadow-sm backdrop-blur-xl bg-background/60 border-muted/60">
             <CardHeader className="pb-3">
                 <CardTitle className="text-sm font-semibold flex items-center gap-2 text-muted-foreground uppercase tracking-wider">
                     <Network className="h-4 w-4" /> Active Interfaces
                 </CardTitle>
             </CardHeader>
             <CardContent className="space-y-4">
                 <div className="flex justify-between items-center pb-3 border-b">
                     <div className="flex flex-col">
                         <span className="text-sm font-medium">HTTP / JSON-RPC</span>
                         <span className="text-xs text-muted-foreground font-mono">Port {sysStatus?.bound_http_port || "Unknown"}</span>
                     </div>
                     <Badge variant="outline" className="bg-blue-500/10 text-blue-600 border-blue-500/20 px-2 flex items-center gap-1">
                         <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
                         </span>
                         Listening
                     </Badge>
                 </div>
                 <div className="flex justify-between items-center pb-3 border-b">
                     <div className="flex flex-col">
                         <span className="text-sm font-medium">gRPC Registration</span>
                         <span className="text-xs text-muted-foreground font-mono">Port {sysStatus?.bound_grpc_port || "Unknown"}</span>
                     </div>
                     <Badge variant="outline" className="bg-blue-500/10 text-blue-600 border-blue-500/20 px-2 flex items-center gap-1">
                         <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
                         </span>
                         Listening
                     </Badge>
                 </div>
                 <div className="flex justify-between items-center">
                     <div className="flex flex-col">
                         <span className="text-sm font-medium">Active Connections</span>
                         <span className="text-xs text-muted-foreground">Current web sessions</span>
                     </div>
                     <span className="text-lg font-bold font-mono">{sysStatus?.active_connections || 0}</span>
                 </div>
             </CardContent>
         </Card>
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-semibold tracking-tight">System Checks</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {Object.entries(report?.checks || {}).map(([name, result]) => (
                <Card key={name} className="flex flex-col overflow-hidden transition-all hover:border-primary/30 backdrop-blur-sm bg-background/50">
                    <CardHeader className="p-4 pb-2 flex flex-row items-center justify-between space-y-0">
                        <CardTitle className="text-sm font-semibold flex items-center gap-2">
                            {getIconForCheck(name)}
                            {name}
                        </CardTitle>
                        {getStatusBadge(result.status)}
                    </CardHeader>
                    <CardContent className="p-4 pt-2 flex-1 flex flex-col justify-between">
                        <div className="text-xs text-muted-foreground mb-4 line-clamp-3">
                            {result.message || "Integrity verified successfully."}
                        </div>
                        <div className="flex items-center justify-between pt-3 border-t border-border/50 text-[10px] text-muted-foreground uppercase tracking-wider font-semibold">
                            <span>Latency</span>
                            <span className="font-mono text-foreground">{result.latency || "< 1ms"}</span>
                        </div>
                        {result.diff && (
                            <div className="mt-3 p-2 bg-red-500/10 border border-red-500/20 rounded text-[10px] font-mono break-all text-red-600 dark:text-red-400 max-h-24 overflow-y-auto">
                                {result.diff}
                            </div>
                        )}
                    </CardContent>
                </Card>
            ))}
        </div>
      </div>
    </div>
  );
}
