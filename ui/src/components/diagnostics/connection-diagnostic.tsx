/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { CheckCircle2, XCircle, Loader2, Play, Activity, Terminal } from "lucide-react";
import { cn } from "@/lib/utils";
import { UpstreamServiceConfig } from "@/lib/types";
import { analyzeConnectionError } from "@/lib/diagnostics-utils";

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
}

export function ConnectionDiagnosticDialog({ service, trigger }: ConnectionDiagnosticDialogProps) {
  const [open, setOpen] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [steps, setSteps] = useState<DiagnosticStep[]>([]);

  const resetSteps = () => {
    setSteps([
      { id: "config", name: "Client-Side Configuration Check", status: "pending", logs: [] },
      { id: "backend_health", name: "Backend Status Check", status: "pending", logs: [] },
    ]);
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
    }

    if (!isValid) {
        updateStep("config", { status: "failure", detail: "Invalid Configuration" });
        setIsRunning(false);
        return;
    }
    updateStep("config", { status: "success", detail: "Configuration valid" });


    // --- Step 2: Active Diagnostics ---
    updateStep("backend_health", { status: "running" });
    addLog("backend_health", "Running active diagnostics...");

    try {
        const res = await fetch("/api/v1/diagnose", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ service_name: service.name }),
        });

        if (!res.ok) {
             throw new Error(`API Error: ${res.status} ${res.statusText}`);
        }

        const report = await res.json();

        // Add dynamic steps from backend report
        const newSteps: DiagnosticStep[] = report.steps.map((s: any) => ({
            id: s.name.toLowerCase().replace(/\s+/g, "_"),
            name: s.name,
            status: s.status,
            logs: s.message ? [`${s.message} ${s.latency ? `(${s.latency})` : ''}`] : []
        }));

        setSteps(prev => {
            // Keep config check, replace backend_health with new steps
            const configStep = prev.find(s => s.id === "config");
            return configStep ? [configStep, ...newSteps] : newSteps;
        });

        const allPassed = newSteps.every(s => s.status === 'success' || s.status === 'skipped');
        if (allPassed) {
             addLog("backend_health", "All active checks passed.");
        } else {
             const failedStep = newSteps.find(s => s.status === 'failure');
             if (failedStep) {
                 const diagnosis = analyzeConnectionError(failedStep.logs.join("\n"));
                 if (diagnosis.category !== 'unknown') {
                     // Add a summary step
                     const analysisStep: DiagnosticStep = {
                         id: "analysis",
                         name: "Failure Analysis",
                         status: "failure",
                         logs: [
                             `Analysis: ${diagnosis.title}`,
                             `Description: ${diagnosis.description}`,
                             `Suggestion: ${diagnosis.suggestion}`
                         ]
                     };
                     setSteps(prev => [...prev, analysisStep]);
                 }
             }
        }

    } catch (error: any) {
        addLog("backend_health", `Failed to contact diagnostics API: ${error.message}`);
        updateStep("backend_health", { status: "failure", detail: "API Failure" });
    }

    setIsRunning(false);
  };

  const handleOpenChange = (open: boolean) => {
      setOpen(open);
      if (open && steps.length === 0) {
          resetSteps();
      }
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
      <DialogContent className="sm:max-w-[700px] h-[600px] flex flex-col p-0 gap-0 overflow-hidden">
        <DialogHeader className="p-6 border-b bg-muted/20">
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
                {steps.map((step, index) => (
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

            {/* Logs View */}
            <div className="col-span-2 bg-black/90 text-green-400 font-mono text-xs p-4 overflow-hidden flex flex-col">
                <div className="flex items-center gap-2 mb-2 pb-2 border-b border-white/10 text-muted-foreground">
                    <Terminal className="h-3 w-3" />
                    <span>Diagnostic Logs</span>
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
        </div>

        <DialogFooter className="p-4 border-t bg-muted/20">
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
