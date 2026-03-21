"use client";

import { useState, useEffect } from "react";
import { apiClient, DiscoveryData } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Loader2, RefreshCw, FileText, Info, AlertTriangle, AlertCircle, Eye, Activity, Search } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";

/**
 * DiscoveryStatus component.
 * Displays information about how servers were discovered and configured.
 * @returns The rendered component.
 */
export function DiscoveryStatus() {
  const [data, setData] = useState<DiscoveryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDiscoveryData = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.getDiscoveryStatus();
      setData(res);
    } catch (err) {
      console.error("Failed to fetch discovery status", err);
      setError("Failed to retrieve discovery information. The backend might be unreachable.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDiscoveryData();
  }, []);

  if (loading && !data) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <div className="relative">
           <Search className="h-10 w-10 text-primary" />
           <Activity className="absolute bottom-0 right-0 h-4 w-4 text-green-500 animate-pulse" />
           <div className="absolute inset-0 rounded-full bg-primary/20 animate-ping opacity-30" />
        </div>
        <p className="text-muted-foreground text-sm uppercase tracking-wider font-semibold">Tracing registry sources...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center pb-4 border-b">
        <div>
          <h3 className="text-lg font-semibold tracking-tight">Configuration Registry</h3>
          <p className="text-xs text-muted-foreground mt-1 uppercase tracking-wider">
            Active servers and their discovery pathways
          </p>
        </div>
        <Button onClick={fetchDiscoveryData} disabled={loading} variant="outline" size="sm">
          <RefreshCw className={cn("mr-2 h-3.5 w-3.5", loading && "animate-spin")} />
          Resync Sources
        </Button>
      </div>

      {/* Error handling */}
      {error && (
        <div className="bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20 px-4 py-3 rounded-md flex items-start gap-3 shadow-sm">
          <AlertTriangle className="h-5 w-5 mt-0.5 shrink-0" />
          <div className="text-sm font-medium">{error}</div>
        </div>
      )}

      {/* List of active servers */}
      <div className="space-y-4">
        {(!data?.servers || data.servers.length === 0) ? (
          <div className="text-center p-12 border border-dashed rounded-lg bg-background/50 text-muted-foreground">
            <div className="relative mx-auto w-12 h-12 mb-3">
                <Info className="absolute inset-0 h-12 w-12 opacity-50 text-muted-foreground" />
            </div>
            <p className="text-sm font-medium">No active server registries detected.</p>
            <p className="text-xs opacity-70 mt-1">Connect a server via configuration files or the API.</p>
          </div>
        ) : (
          <div className="grid gap-4">
            {data.servers.map((server: any) => {
                const hasError = !!server.error;
                const isError = hasError || !server.connected;

                return (
                  <Card key={server.id} className={cn("overflow-hidden transition-all backdrop-blur-xl bg-background/60 shadow-sm", isError ? "border-red-500/30" : "border-border/50 hover:border-primary/40")}>
                    <CardHeader className="py-3 px-4 flex flex-row items-center justify-between space-y-0 bg-muted/20 border-b">
                      <div className="flex items-center gap-3">
                        <CardTitle className="text-base font-medium flex items-center gap-2">
                          {server.name}
                          {hasError && (
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <AlertCircle className="h-4 w-4 text-red-500" />
                                </TooltipTrigger>
                                <TooltipContent className="bg-red-500 text-white">
                                  <p>{server.error}</p>
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          )}
                        </CardTitle>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant={server.connected ? "default" : "secondary"} className={cn("px-2 py-0.5 text-[10px] uppercase tracking-wider font-semibold shadow-sm border",
                          server.connected ? "bg-green-500/10 text-green-600 border-green-500/20 hover:bg-green-500/20" : "bg-muted text-muted-foreground"
                        )}>
                          {server.connected ? (
                              <span className="flex items-center gap-1.5">
                                  <span className="relative flex h-1.5 w-1.5"><span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span><span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-green-500"></span></span>
                                  Active
                              </span>
                          ) : "Disconnected"}
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent className="p-0">
                      <div className="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x">
                        <div className="p-5 space-y-4">
                          <div className="flex items-center gap-2 text-[10px] font-semibold text-muted-foreground uppercase tracking-wider border-b pb-2">
                              <FileText className="h-3 w-3" /> Origin Profile
                          </div>

                          {/* Source Type / Path */}
                          <div className="flex flex-col gap-1">
                             <div className="flex items-center justify-between">
                               <span className="text-xs text-muted-foreground">Provider:</span>
                               <Badge variant="outline" className="text-[10px] uppercase tracking-wider font-mono bg-background shadow-sm">{server.sourceType}</Badge>
                             </div>
                             {server.sourcePath && (
                               <div className="text-xs text-muted-foreground bg-muted/30 p-2 rounded-md font-mono break-all mt-2 border border-border/50">
                                 {server.sourcePath}
                               </div>
                             )}
                          </div>

                          {/* Raw Config View */}
                          {server.rawConfig && (
                             <Dialog key={`dialog-${server.id}`}>
                              <DialogTrigger asChild>
                                <Button variant="secondary" size="sm" className="w-full mt-2 h-7 text-xs shadow-sm bg-background border border-border/50 hover:bg-muted/50">
                                  <Eye className="mr-2 h-3 w-3 text-primary" /> Inspect Payload
                                </Button>
                              </DialogTrigger>
                              <DialogContent className="max-w-2xl">
                                <DialogHeader>
                                  <DialogTitle className="flex items-center gap-2 text-base">
                                    <FileText className="h-4 w-4" /> Raw Configuration
                                  </DialogTitle>
                                  <DialogDescription>
                                    The original configuration object as it was parsed from the source.
                                  </DialogDescription>
                                </DialogHeader>
                                <ScrollArea className="h-[400px] w-full rounded-md border p-4 bg-muted/30">
                                  <pre className="text-xs font-mono text-primary/80">{JSON.stringify(server.rawConfig, null, 2)}</pre>
                                </ScrollArea>
                              </DialogContent>
                            </Dialog>
                          )}
                        </div>

                        <div className="p-5 bg-gradient-to-br from-muted/10 to-transparent space-y-4">
                          <div className="flex items-center gap-2 text-[10px] font-semibold text-muted-foreground uppercase tracking-wider border-b pb-2">
                              <Activity className="h-3 w-3" /> Execution Context
                          </div>

                          {/* Command */}
                          <div className="flex flex-col gap-2">
                            <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Binary & Args</span>
                            <div className="bg-background/80 border border-border/60 rounded px-3 py-2 text-xs font-mono overflow-x-auto whitespace-nowrap shadow-sm flex items-center gap-2">
                              <span className="text-green-600 dark:text-green-400 font-bold">$</span>
                              <span className="text-foreground">{server.command}</span>
                              {server.args && server.args.length > 0 && (
                                <span className="text-muted-foreground ml-2">{server.args.join(" ")}</span>
                              )}
                            </div>
                          </div>

                          {/* Environment Variables */}
                          {server.env && Object.keys(server.env).length > 0 && (
                            <div className="flex flex-col gap-1.5 mt-3">
                              <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Environment ({Object.keys(server.env).length})</span>
                              <ScrollArea className="h-[80px] w-full rounded border bg-background/50">
                                <div className="p-2 space-y-1">
                                  {Object.entries(server.env).map(([k, v]) => (
                                    <div key={k} className="text-xs font-mono flex">
                                      <span className="text-blue-600 dark:text-blue-400 font-semibold w-1/3 truncate pr-2" title={k}>{k}</span>
                                      <span className="text-muted-foreground w-2/3 truncate text-[10px]" title={v as string}>={String(v)}</span>
                                    </div>
                                  ))}
                                </div>
                              </ScrollArea>
                            </div>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
