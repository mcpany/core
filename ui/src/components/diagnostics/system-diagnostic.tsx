/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { CheckCircle2, XCircle, Loader2, Play, Terminal, Server, ShieldCheck, Globe, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";

interface DiagnosticStep {
  id: string;
  name: string;
  status: "pending" | "running" | "success" | "failure" | "skipped";
  detail?: string;
  logs: string[];
  icon?: React.ReactNode;
}

interface DoctorReport {
    status: string;
    checks: Record<string, {
        status: string;
        message?: string;
        latency?: string;
    }>;
}

interface SystemStatus {
    version: string;
    uptime_seconds: number;
    active_connections: number;
    security_warnings: string[];
}

export function SystemDiagnostic() {
  const [isRunning, setIsRunning] = useState(false);
  const [steps, setSteps] = useState<DiagnosticStep[]>([]);

  // Auto-start on mount
  useEffect(() => {
      runDiagnostics();
  }, []);

  const resetSteps = () => {
    setSteps([
      { id: "system_status", name: "System Status", status: "pending", logs: [], icon: <Server className="h-4 w-4" /> },
      { id: "doctor_checks", name: "Health Checks (Doctor)", status: "pending", logs: [], icon: <ShieldCheck className="h-4 w-4" /> },
      { id: "services_health", name: "Services Connectivity", status: "pending", logs: [], icon: <Globe className="h-4 w-4" /> },
    ]);
  };

  const runDiagnostics = async () => {
    setIsRunning(true);
    resetSteps();

    const updateStep = (id: string, updates: Partial<DiagnosticStep>) => {
      setSteps((prev) => prev.map((s) => (s.id === id ? { ...s, ...updates } : s)));
    };

    const addLog = (id: string, message: string) => {
      setSteps((prev) => prev.map((s) => (s.id === id ? { ...s, logs: [...s.logs, `[${new Date().toLocaleTimeString()}] ${message}`] } : s)));
    };

    // --- Step 1: System Status ---
    updateStep("system_status", { status: "running" });
    try {
        const res = await fetch("/api/v1/system/status");
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const status: SystemStatus = await res.json();

        addLog("system_status", `Version: ${status.version}`);
        addLog("system_status", `Uptime: ${status.uptime_seconds}s`);
        addLog("system_status", `Active Connections: ${status.active_connections}`);

        if (status.security_warnings && status.security_warnings.length > 0) {
            status.security_warnings.forEach(w => addLog("system_status", `WARNING: ${w}`));
            updateStep("system_status", { status: "success", detail: "Running with Warnings" }); // Green but warns? Or Yellow?
        } else {
            updateStep("system_status", { status: "success", detail: "Running" });
        }
    } catch (e: any) {
        addLog("system_status", `Failed to fetch system status: ${e.message}`);
        updateStep("system_status", { status: "failure", detail: "Unreachable" });
    }

    // --- Step 2: Doctor Checks ---
    updateStep("doctor_checks", { status: "running" });
    try {
        const res = await fetch("/api/v1/doctor");
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const report: DoctorReport = await res.json();

        addLog("doctor_checks", `Overall Status: ${report.status}`);
        let allOk = true;
        for (const [checkName, result] of Object.entries(report.checks)) {
            if (result.status !== "ok") {
                addLog("doctor_checks", `[FAIL] ${checkName}: ${result.message}`);
                allOk = false;
            } else {
                addLog("doctor_checks", `[OK] ${checkName} (${result.latency || '0s'})`);
            }
        }

        if (report.status === "healthy" && allOk) {
            updateStep("doctor_checks", { status: "success", detail: "All Systems Go" });
        } else {
            updateStep("doctor_checks", { status: "failure", detail: "Issues Detected" });
        }
    } catch (e: any) {
        addLog("doctor_checks", `Doctor check failed: ${e.message}`);
        updateStep("doctor_checks", { status: "failure", detail: "Failed" });
    }

    // --- Step 3: Services Health ---
    updateStep("services_health", { status: "running" });
    try {
        // We use the dashboard health API (ui/src/app/api/dashboard/health/route.ts) which wraps /api/v1/services
        // Or call /api/v1/services directly?
        // Let's call the frontend API to leverage its logic, or backend directly.
        // Calling backend directly via proxy is better for raw diagnostics.
        const res = await fetch("/api/v1/services");
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const services: any[] = await res.json();

        const activeServices = services.filter((s: any) => !s.disable);
        const unhealthyServices = activeServices.filter((s: any) => s.last_error);

        addLog("services_health", `Total Configured Services: ${services.length}`);
        addLog("services_health", `Active: ${activeServices.length}`);

        if (unhealthyServices.length > 0) {
            addLog("services_health", `Found ${unhealthyServices.length} unhealthy services:`);
            unhealthyServices.forEach((s: any) => {
                addLog("services_health", ` - ${s.name}: ${s.last_error}`);
            });
            updateStep("services_health", { status: "failure", detail: `${unhealthyServices.length} Unhealthy` });
        } else {
             if (activeServices.length === 0) {
                 addLog("services_health", "No active services configured.");
                 updateStep("services_health", { status: "skipped", detail: "No Services" });
             } else {
                 addLog("services_health", "All active services are healthy.");
                 updateStep("services_health", { status: "success", detail: "All Healthy" });
             }
        }

    } catch (e: any) {
        addLog("services_health", `Failed to list services: ${e.message}`);
        updateStep("services_health", { status: "failure", detail: "Failed" });
    }

    setIsRunning(false);
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6 h-[600px]">
        {/* Steps List */}
        <Card className="col-span-1 border-r bg-muted/10 h-full flex flex-col">
            <CardHeader>
                <CardTitle>System Health</CardTitle>
                <CardDescription>Self-diagnostic checks</CardDescription>
            </CardHeader>
            <CardContent className="flex-1 overflow-y-auto">
            <div className="space-y-4">
                {steps.map((step) => (
                    <div key={step.id} className={cn(
                        "relative pl-6 pb-4 border-l-2 last:border-0 transition-colors",
                        step.status === 'success' ? "border-green-500" :
                        step.status === 'failure' ? "border-red-500" :
                        step.status === 'running' ? "border-blue-500" : "border-muted"
                    )}>
                        <div className={cn(
                            "absolute -left-[9px] top-0 w-5 h-5 rounded-full border-2 bg-background flex items-center justify-center",
                            step.status === 'success' ? "border-green-500 text-green-500" :
                            step.status === 'failure' ? "border-red-500 text-red-500" :
                            step.status === 'running' ? "border-blue-500 text-blue-500 animate-pulse" : "border-muted text-muted-foreground"
                        )}>
                            {step.status === 'success' && <CheckCircle2 className="h-3 w-3" />}
                            {step.status === 'failure' && <XCircle className="h-3 w-3" />}
                            {step.status === 'running' && <div className="h-2 w-2 rounded-full bg-blue-500" />}
                            {step.status === 'skipped' && <div className="h-2 w-2 rounded-full bg-muted-foreground" />}
                            {step.status === 'pending' && <div className="h-2 w-2 rounded-full bg-muted" />}
                        </div>

                        <div className="space-y-1">
                            <div className="flex items-center gap-2">
                                {step.icon}
                                <p className={cn("text-sm font-medium leading-none",
                                    step.status === 'running' && "text-blue-600 dark:text-blue-400",
                                    step.status === 'failure' && "text-red-600 dark:text-red-400"
                                )}>
                                    {step.name}
                                </p>
                            </div>
                            <p className="text-xs text-muted-foreground">
                                {step.status === 'pending' ? 'Pending...' :
                                 step.status === 'running' ? 'Checking...' :
                                 step.detail || step.status}
                            </p>
                        </div>
                    </div>
                ))}
            </div>
            </CardContent>
             <div className="p-4 border-t">
                <Button
                    onClick={runDiagnostics}
                    disabled={isRunning}
                    className="w-full"
                    variant={steps.some(s => s.status === 'failure') ? "destructive" : "default"}
                >
                    {isRunning ? (
                        <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Running...
                        </>
                    ) : (
                        <>
                            <Play className="mr-2 h-4 w-4" />
                            Rerun Diagnostics
                        </>
                    )}
                </Button>
            </div>
        </Card>

        {/* Logs View */}
        <Card className="col-span-1 md:col-span-2 bg-black text-green-400 font-mono text-xs h-full flex flex-col border-none shadow-inner">
            <CardHeader className="bg-muted/10 border-b border-white/10 py-3">
                <div className="flex items-center gap-2 text-muted-foreground">
                    <Terminal className="h-4 w-4" />
                    <span className="font-sans font-medium">Diagnostic Logs</span>
                </div>
            </CardHeader>
            <CardContent className="p-4 flex-1 overflow-hidden">
                <ScrollArea className="h-full pr-4">
                    <div className="space-y-1">
                        {steps.flatMap(s => s.logs).length === 0 ? (
                            <span className="text-muted-foreground/50 italic">Waiting to start diagnostics...</span>
                        ) : (
                            steps.flatMap(s => s.logs).map((log, i) => (
                                <div key={i} className="break-all whitespace-pre-wrap font-mono">{log}</div>
                            ))
                        )}
                        {isRunning && (
                            <div className="animate-pulse">_</div>
                        )}
                    </div>
                </ScrollArea>
            </CardContent>
        </Card>
    </div>
  );
}
