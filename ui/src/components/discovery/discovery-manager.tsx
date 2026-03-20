/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { RefreshCw, PlayCircle, ShieldCheck, Download, Plus, CheckCircle2, AlertTriangle, Clock } from "lucide-react";
import { UpstreamServiceConfig } from "@/lib/types";

export function DiscoveryManager() {
  const { toast } = useToast();
  const [statuses, setStatuses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [triggering, setTriggering] = useState(false);
  const [discoveredServices, setDiscoveredServices] = useState<UpstreamServiceConfig[]>([]);

  const fetchStatus = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getDiscoveryStatus();
      setStatuses(data);

      // In a real scenario, discovered services might be returned separately or we query catalog
      const catalog = await apiClient.listCatalog();
      setDiscoveredServices(catalog);
    } catch (err: any) {
      console.error(err);
      toast({ title: "Failed to load discovery status", variant: "destructive", description: err.message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleTrigger = async () => {
    setTriggering(true);
    try {
      await apiClient.triggerDiscovery();
      toast({ title: "Discovery Triggered", description: "Auto-discovery is now running in the background." });
      setTimeout(fetchStatus, 1000);
    } catch (err: any) {
      toast({ title: "Trigger Failed", variant: "destructive", description: err.message });
    } finally {
      setTriggering(false);
    }
  };

  const handleImport = async (service: UpstreamServiceConfig) => {
     try {
         await apiClient.registerService(service);
         toast({ title: "Service Imported", description: `${service.name} has been added to your Active Services.` });
     } catch (err: any) {
         toast({ title: "Import Failed", variant: "destructive", description: err.message });
     }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Unified Discovery Manager</h1>
          <p className="text-muted-foreground mt-1">
            Manage auto-discovery providers and review newly found MCP servers.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={fetchStatus} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} /> Refresh
          </Button>
          <Button onClick={handleTrigger} disabled={triggering}>
            {triggering ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <PlayCircle className="mr-2 h-4 w-4" />}
            Trigger Scan
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Discovery Providers</CardTitle>
            <CardDescription>Status of configured discovery sources.</CardDescription>
          </CardHeader>
          <CardContent>
            {statuses.length === 0 ? (
              <div className="text-center py-6 text-muted-foreground">No providers configured.</div>
            ) : (
              <div className="space-y-4">
                {statuses.map(s => (
                  <div key={s.Name} className="flex items-center justify-between border-b pb-4 last:border-0 last:pb-0">
                    <div className="flex flex-col">
                      <span className="font-semibold flex items-center gap-2">
                        {s.Name}
                        <Badge variant={s.Status === "OK" ? "default" : s.Status === "ERROR" ? "destructive" : "secondary"}>
                          {s.Status}
                        </Badge>
                      </span>
                      <span className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
                         <Clock className="h-3 w-3" />
                         Last run: {new Date(s.LastRunAt).toLocaleString()}
                      </span>
                      {s.LastError && <span className="text-xs text-red-500 mt-1">{s.LastError}</span>}
                    </div>
                    <div className="text-right">
                      <span className="text-2xl font-bold">{s.DiscoveredCount}</span>
                      <span className="block text-xs text-muted-foreground">Discovered</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div>
        <h2 className="text-xl font-bold mb-4">Discovered Servers</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
           {discoveredServices.length === 0 ? (
               <div className="col-span-full py-12 text-center border rounded-lg border-dashed text-muted-foreground">
                   No new servers discovered. Try triggering a scan.
               </div>
           ) : (
               discoveredServices.map(svc => (
                   <Card key={svc.id || svc.name} className="flex flex-col hover:border-primary/50 transition-colors">
                       <CardHeader>
                           <CardTitle className="flex justify-between items-start">
                               <span>{svc.name}</span>
                               <ShieldCheck className="h-4 w-4 text-green-500" />
                           </CardTitle>
                           <CardDescription className="line-clamp-2 min-h-[40px]">
                               {svc.description || "A discovered MCP service."}
                           </CardDescription>
                       </CardHeader>
                       <CardContent className="flex-1">
                           <div className="flex gap-2 flex-wrap">
                               {svc.tags?.map((t: string) => <Badge key={t} variant="secondary" className="text-[10px]">{t}</Badge>)}
                               <Badge variant="outline" className="text-[10px]">
                                   {svc.httpService ? "HTTP" : svc.commandLineService ? "CLI" : "Other"}
                               </Badge>
                           </div>
                       </CardContent>
                       <CardFooter>
                           <Button className="w-full" onClick={() => handleImport(svc)}>
                               <Download className="mr-2 h-4 w-4" /> Import to Active
                           </Button>
                       </CardFooter>
                   </Card>
               ))
           )}
        </div>
      </div>

    </div>
  );
}
