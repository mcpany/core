/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { CheckCircle2, XCircle, Loader2, Play, Activity, Terminal, AlertTriangle, Lightbulb, Copy, Check, Settings } from "lucide-react";
import { cn } from "@/lib/utils";
import { UpstreamServiceConfig } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { analyzeConnectionError, DiagnosticResult } from "@/lib/diagnostics-utils";

/* eslint-disable @typescript-eslint/no-explicit-any */

interface DiagnosticStep {
  id: string;
  name: string;
  status: "pending" | "running" | "success" | "failure" | "skipped";
  detail?: string;
  logs: string[];
}

interface ServiceHealth {
  id: string;
  name: string;
  status: "healthy" | "degraded" | "unhealthy" | "inactive";
  message?: string;
}

interface ConnectionDiagnosticDialogProps {
  service: UpstreamServiceConfig;
  trigger?: React.ReactNode;
  onEdit?: (service: UpstreamServiceConfig, tab?: string) => void;
}

/**
 * ConnectionDiagnosticDialog.
 *
 * @param trigger - The trigger.
 */
export function ConnectionDiagnosticDialog({ service, trigger, onEdit }: ConnectionDiagnosticDialogProps) {
  const [open, setOpen] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [steps, setSteps] = useState<DiagnosticStep[]>([]);
  const [diagnosticResult, setDiagnosticResult] = useState<DiagnosticResult | null>(null);
  const [copied, setCopied] = useState(false);

  const resetSteps = () => {
    const initialSteps: DiagnosticStep[] = [
      { id: "config", name: "Client-Side Configuration Check", status: "pending", logs: [] },
      { id: "backend_health", name: "Backend Status Check", status: "pending", logs: [] },
      { id: "operational", name: "Operational Verification", status: "pending", logs: [] },
    ];

    if (service.websocketService || service.httpService) {
        initialSteps.splice(1, 0, { id: "browser_connectivity", name: "Browser Connectivity Check", status: "pending", logs: [] });
    }

    setSteps(initialSteps);
    setDiagnosticResult(null);
  };

  const runDiagnostics = async () => {
    setIsRunning(true);
    resetSteps();

    // Helper to update a step
    const updateStep = (id: string, updates: Partial<DiagnosticStep>) => {
      setSteps((prev) => prev.map((s) => (s.id === id ? { ...s, ...updates } : s)));
    };

    const addLog = (id: string, message: string) => {
      setSteps((prev) => prev.map((s) => (s.id === id ? { ...s, logs: [...s.logs, `[${new Date().toLocaleTimeString()}] ${message}`] } : s)));
    };

    // --- Step 1: Config Validation ---
    updateStep("config", { status: "running" });
    await new Promise(r => setTimeout(r, 600)); // Simulate UI delay

    let isValid = true;
    let url = "";

    if (service.httpService) {
        url = service.httpService.address;
        if (!url.startsWith("http")) {
            addLog("config", "Error: HTTP address must start with http:// or https://");
            isValid = false;
        } else {
            addLog("config", `Validating HTTP address: ${url}`);
        }
    } else if (service.grpcService) {
        url = service.grpcService.address;
        addLog("config", `Validating gRPC address: ${url}`);
    } else if (service.commandLineService) {
        addLog("config", `Validating Command: ${service.commandLineService.command}`);
        if (!service.commandLineService.command) {
             addLog("config", "Error: Command is empty");
             isValid = false;
        }
    } else if (service.websocketService) {
        url = service.websocketService.address;
        addLog("config", `Validating WebSocket address: ${url}`);
        if (!url.startsWith("ws://") && !url.startsWith("wss://")) {
            addLog("config", "Error: WebSocket address must start with ws:// or wss://");
            isValid = false;
        }
    }

    if (!isValid) {
        updateStep("config", { status: "failure", detail: "Invalid Configuration" });
        setIsRunning(false);
        return;
    }
    updateStep("config", { status: "success", detail: "Configuration valid" });

    // --- Step 1.5: Browser Connectivity (WebSocket & HTTP) ---
    if (service.websocketService) {
        updateStep("browser_connectivity", { status: "running" });
        addLog("browser_connectivity", `Attempting to connect to ${url} from browser...`);

        try {
            await new Promise<string>((resolve, reject) => {
                const ws = new WebSocket(url);
                const timer = setTimeout(() => {
                    ws.close();
                    reject(new Error("Connection timed out (5s)"));
                }, 5000);

                ws.onopen = () => {
                    clearTimeout(timer);
                    ws.close();
                    resolve("Success");
                };

                ws.onerror = () => {
                    clearTimeout(timer);
                    reject(new Error("Connection failed (Network error or blocked)"));
                };
            });

            addLog("browser_connectivity", "Successfully connected to WebSocket server from browser.");
            updateStep("browser_connectivity", { status: "success", detail: "Accessible" });
        } catch (error: unknown) {
            const msg = error instanceof Error ? error.message : "Unknown error";
            addLog("browser_connectivity", `Failed to connect from browser: ${msg}`);
            addLog("browser_connectivity", "Note: This is expected if the server is internal or behind a firewall not accessible from your browser.");
            updateStep("browser_connectivity", { status: "failure", detail: "Not Accessible" });
            // Don't stop diagnostics, backend might still see it
        }
    } else if (service.httpService) {
        updateStep("browser_connectivity", { status: "running" });
        const httpUrl = service.httpService.address;
        addLog("browser_connectivity", `Attempting to connect to ${httpUrl} from browser...`);

        try {
            // mode: 'no-cors' allows us to send a request to another origin.
            // We won't see the response, but if it doesn't throw, it means the server is reachable (DNS + TCP + TLS).
            await fetch(httpUrl, { mode: 'no-cors', cache: 'no-store' });

            addLog("browser_connectivity", "Successfully connected to HTTP server from browser.");
            addLog("browser_connectivity", "Note: 'no-cors' mode used. We can reach the server, but cannot read the response due to CORS policy. This confirms network connectivity.");
            updateStep("browser_connectivity", { status: "success", detail: "Accessible" });
        } catch (error: unknown) {
            const msg = error instanceof Error ? error.message : "Unknown error";
            addLog("browser_connectivity", `Failed to connect from browser: ${msg}`);

            // Localhost Warning Heuristic
            if (httpUrl.includes("localhost") || httpUrl.includes("127.0.0.1") || httpUrl.includes("0.0.0.0")) {
                addLog("browser_connectivity", "⚠️ WARNING: You are using 'localhost' or loopback address.");
                addLog("browser_connectivity", "If MCP Any is running in Docker, it cannot access 'localhost' of your host machine directly.");
                addLog("browser_connectivity", "Try using 'host.docker.internal' or your LAN IP address.");
            }

            addLog("browser_connectivity", "Possible causes: Server down, blocked by firewall, invalid SSL cert, Mixed Content blocking, or CSP (Content Security Policy) restrictions.");
            updateStep("browser_connectivity", { status: "failure", detail: "Not Accessible" });
            // Don't stop diagnostics
        }
    }


    // --- Step 2: Backend Health Check ---
    updateStep("backend_health", { status: "running" });
    addLog("backend_health", "Querying backend service status...");

    try {
        const res = await fetch("/api/dashboard/health", { cache: 'no-store' });
        if (!res.ok) {
             throw new Error(`API Error: ${res.status} ${res.statusText}`);
        }
        const data: ServiceHealth[] = await res.json();

        // Find our service
        const serviceStatus = data.find(s => s.id === service.id || s.name === service.name);

        if (!serviceStatus) {
             addLog("backend_health", "Warning: Service not found in backend registry.");
             addLog("backend_health", "This might happen if the service was just added or backend is restarting.");
             updateStep("backend_health", { status: "failure", detail: "Service Not Found" });
        } else {
             addLog("backend_health", `Backend reports status: ${serviceStatus.status.toUpperCase()}`);

             if (serviceStatus.status === 'healthy') {
                 addLog("backend_health", "Service is connected and responding.");
                 updateStep("backend_health", { status: "success", detail: "Connected" });
             } else if (serviceStatus.status === 'inactive') {
                 addLog("backend_health", "Service is explicitly disabled.");
                 updateStep("backend_health", { status: "skipped", detail: "Disabled" });
             } else {
                 // Unhealthy or Degraded
                 addLog("backend_health", `Error: ${serviceStatus.message || "Unknown error"}`);

                 const diagnosis = analyzeConnectionError(serviceStatus.message || "");
                 if (diagnosis.category !== 'unknown') {
                     addLog("backend_health", `Analysis: ${diagnosis.title} - ${diagnosis.description}`);
                     addLog("backend_health", `Suggestion: ${diagnosis.suggestion}`);
                     setDiagnosticResult(diagnosis);
                 }

                 updateStep("backend_health", { status: "failure", detail: serviceStatus.status });
             }
        }

    } catch (error: unknown) {
        const msg = error instanceof Error ? error.message : "Unknown error";
        addLog("backend_health", `Failed to contact backend API: ${msg}`);
        updateStep("backend_health", { status: "failure", detail: "API Failure" });
    }

    // --- Step 3: Operational Verification ---
    // Only proceed if backend check was successful or skipped (disabled)
    // We access current state via a fresh look or logic flow, but here 'steps' state is stale closure.
    // However, we can infer success if we reached here without returning early?
    // Actually, I didn't return early in backend_health block. I should probably check if previous steps failed.

    // Let's perform operational check regardless, as it fetches specific service details which might reveal more.
    updateStep("operational", { status: "running" });
    addLog("operational", "Verifying service operations...");

    try {
        const fullServiceRes = await apiClient.getService(service.name);

        // API might return { service: ... } or just the service object depending on impl
        const s = fullServiceRes.service || fullServiceRes;

        if (s) {
            addLog("operational", `Tools Discovered: ${s.toolCount ?? 0}`);

            if (s.lastError) {
                addLog("operational", `Service reported error: ${s.lastError}`);
                const diagnosis = analyzeConnectionError(s.lastError);
                // Only override diagnosis if it's more specific/severe or if we didn't have one
                setDiagnosticResult(prev => {
                    if (!prev || diagnosis.category !== 'unknown') return diagnosis;
                    return prev;
                });
                updateStep("operational", { status: "failure", detail: "Operational Error" });
            } else if ((s.toolCount ?? 0) === 0) {
                addLog("operational", "Warning: Service is healthy but exposed 0 tools.");
                addLog("operational", "Check if the service requires specific configuration to expose tools (e.g., allowed paths, exposed schemas).");

                // Only set warning if we don't already have a critical/failure diagnosis
                setDiagnosticResult(prev => {
                    if (prev && prev.severity === 'critical') return prev;

                    return {
                        category: "configuration",
                        title: "No Tools Discovered",
                        description: "The service connected successfully but did not register any tools.",
                        suggestion: "1. Check upstream service configuration (e.g. allowed paths for filesystem).\n2. Verify the upstream service actually exposes tools (some might only expose resources).\n3. Check logs for silent failures.",
                        severity: "warning"
                    };
                });
                updateStep("operational", { status: "success", detail: "No Tools (Warning)" }); // Green but warned in logs
            } else {
                updateStep("operational", { status: "success", detail: "Fully Operational" });
            }
        } else {
            throw new Error("Failed to retrieve service details");
        }

    } catch (e: any) {
        addLog("operational", `Failed to verify operations: ${e.message}`);
        updateStep("operational", { status: "failure", detail: "Verification Failed" });
    }

    setIsRunning(false);
  };

  const handleOpenChange = (open: boolean) => {
      setOpen(open);
      if (open && steps.length === 0) {
          resetSteps();
      }
      setCopied(false);
  };

  const copyLogs = () => {
      const allLogs = steps.flatMap(s => s.logs).join('\n');
      navigator.clipboard.writeText(allLogs).then(() => {
          setCopied(true);
          setTimeout(() => setCopied(false), 2000);
      });
  };

  const handleFixConfiguration = () => {
      if (!onEdit || !diagnosticResult) return;

      let targetTab = "connection";
      if (diagnosticResult.category === "auth") {
          targetTab = "auth";
      } else if (diagnosticResult.title.toLowerCase().includes("environment")) {
          // If it's a CLI service, env vars are in connection tab.
          // Ideally we could point specifically to env section, but connection tab is close enough.
          targetTab = "connection";
      }

      setOpen(false);
      onEdit(service, targetTab);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        {trigger || (
            <Button variant="outline" size="sm" className="gap-2">
                <Activity className="h-4 w-4" /> Troubleshoot
            </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] h-[600px] flex flex-col p-0 gap-0 overflow-hidden backdrop-blur-xl bg-background/95 border-primary/20 shadow-2xl">
        <DialogHeader className="p-6 border-b bg-muted/30">
          <DialogTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5 text-primary" />
              Connection Diagnostics
          </DialogTitle>
          <DialogDescription>
            Diagnose connection issues with <strong>{service.name}</strong>.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden grid grid-cols-3">
            {/* Steps List */}
            <div className="col-span-1 border-r bg-muted/10 overflow-y-auto p-4 space-y-4">
                {steps.map((step) => (
                    <div key={step.id} className={cn(
                        "relative pl-6 pb-4 border-l-2 last:border-0",
                        step.status === 'success' ? "border-green-500" :
                        step.status === 'failure' ? "border-red-500" :
                        step.status === 'running' ? "border-blue-500" : "border-muted"
                    )}>
                        <div className={cn(
                            "absolute -left-[9px] top-0 w-4 h-4 rounded-full border-2 bg-background flex items-center justify-center",
                            step.status === 'success' ? "border-green-500 text-green-500" :
                            step.status === 'failure' ? "border-red-500 text-red-500" :
                            step.status === 'running' ? "border-blue-500 text-blue-500 animate-pulse" : "border-muted text-muted-foreground"
                        )}>
                            {step.status === 'success' && <CheckCircle2 className="h-3 w-3" />}
                            {step.status === 'failure' && <XCircle className="h-3 w-3" />}
                            {step.status === 'running' && <div className="h-2 w-2 rounded-full bg-blue-500" />}
                            {step.status === 'skipped' && <div className="h-2 w-2 rounded-full bg-muted-foreground" />}
                        </div>

                        <div className="space-y-1">
                            <p className={cn("text-sm font-medium leading-none",
                                step.status === 'running' && "text-blue-600 dark:text-blue-400",
                                step.status === 'failure' && "text-red-600 dark:text-red-400"
                            )}>
                                {step.name}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                {step.status === 'pending' ? 'Pending...' :
                                 step.status === 'running' ? 'Checking...' :
                                 step.detail || step.status}
                            </p>
                        </div>
                    </div>
                ))}
            </div>

            {/* Logs View & Diagnosis Report */}
            <div className="col-span-2 flex flex-col h-full bg-zinc-950/90 dark:bg-zinc-950/50 backdrop-blur-sm">
                {/* Logs */}
                <div className="flex-1 overflow-hidden flex flex-col p-4 text-green-400 font-mono text-xs leading-relaxed">
                    <div className="flex items-center justify-between mb-2 pb-2 border-b border-white/10 text-muted-foreground">
                        <div className="flex items-center gap-2">
                            <Terminal className="h-3 w-3" />
                            <span className="font-semibold tracking-wide uppercase text-[10px]">Diagnostic Console</span>
                        </div>
                        {steps.flatMap(s => s.logs).length > 0 && (
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 text-muted-foreground hover:text-white hover:bg-white/10"
                                onClick={copyLogs}
                            >
                                {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                            </Button>
                        )}
                    </div>
                    <ScrollArea className="flex-1">
                        <div className="space-y-1">
                            {steps.flatMap(s => s.logs).length === 0 ? (
                                <span className="text-muted-foreground/50 italic">Waiting to start diagnostics...</span>
                            ) : (
                                steps.flatMap(s => s.logs).map((log, i) => (
                                    <div key={i} className="break-all whitespace-pre-wrap">{log}</div>
                                ))
                            )}
                            {isRunning && (
                                <div className="animate-pulse">_</div>
                            )}
                        </div>
                    </ScrollArea>
                </div>

                {/* Diagnosis Suggestion Card */}
                {diagnosticResult && (
                    <div className="p-4 bg-muted/30 border-t border-white/10">
                        <div className={cn(
                            "rounded-lg border p-4 flex gap-4 shadow-sm",
                            diagnosticResult.severity === 'critical' ? "bg-red-500/10 border-red-500/20" :
                            diagnosticResult.severity === 'warning' ? "bg-amber-500/10 border-amber-500/20" :
                            "bg-blue-500/10 border-blue-500/20"
                        )}>
                            <div className={cn(
                                "shrink-0 mt-0.5 p-1 rounded-full bg-background/50 h-8 w-8 flex items-center justify-center shadow-sm",
                                diagnosticResult.severity === 'critical' ? "text-red-500" :
                                diagnosticResult.severity === 'warning' ? "text-amber-500" :
                                "text-blue-500"
                            )}>
                                <AlertTriangle className="h-4 w-4" />
                            </div>
                            <div className="space-y-1 text-sm flex-1">
                                <p className="font-semibold text-foreground tracking-tight">{diagnosticResult.title}</p>
                                <p className="text-muted-foreground text-xs">{diagnosticResult.description}</p>
                                <div className="mt-3 flex gap-2 text-foreground/90 bg-background/50 p-3 rounded-md text-xs items-start border border-white/5 shadow-sm">
                                    <Lightbulb className="h-3.5 w-3.5 shrink-0 mt-0.5 text-yellow-500" />
                                    <span className="whitespace-pre-wrap font-medium">{diagnosticResult.suggestion}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>

        <DialogFooter className="p-4 border-t bg-muted/30 backdrop-blur-md flex justify-between items-center">
            <div className="flex-1">
                {diagnosticResult && diagnosticResult.action === 'edit_config' && onEdit && (
                    <Button
                        variant="default"
                        size="sm"
                        onClick={handleFixConfiguration}
                        className="gap-2 bg-amber-600 hover:bg-amber-700 text-white"
                    >
                        <Settings className="h-4 w-4" />
                        Fix Configuration
                    </Button>
                )}
            </div>
            <Button
                onClick={runDiagnostics}
                disabled={isRunning}
                className={cn("w-full sm:w-auto min-w-[150px]", isRunning && "opacity-80")}
            >
                {isRunning ? (
                    <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Running...
                    </>
                ) : (
                    <>
                        <Play className="mr-2 h-4 w-4" />
                        {steps.some(s => s.status !== 'pending') ? "Rerun Diagnostics" : "Start Diagnostics"}
                    </>
                )}
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
