/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, XCircle, RefreshCw, ChevronRight, Activity, FileDiff } from "lucide-react";
import { apiClient, DoctorReport } from "@/lib/client";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { UnifiedDiffViewer } from "@/components/diagnostics/unified-diff-viewer";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";

/**
 * SystemStatusBanner component.
 * Displays a non-intrusive status bar when the system is degraded or has configuration errors.
 * Allows viewing detailed diagnostics and diffs in a modal.
 * @returns The rendered component.
 */
export function SystemStatusBanner() {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const fetchHealth = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getDoctorStatus();
      setReport(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      setReport(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
    // Poll every 10 seconds, less aggressive than before but frequent enough
    const interval = setInterval(fetchHealth, 10000);
    return () => clearInterval(interval);
  }, []);

  if (!report && !error) return null;
  if (report?.status === "healthy" || report?.status === "ok") return null;

  // Determine critical issues and aggregate them
  const issues: Array<{ title: string; message: string; type: "critical" | "warning"; diff?: string }> = [];

  if (error) {
      issues.push({ title: "Connection Error", message: `Could not connect to server: ${error}`, type: "critical" });
  } else if (report) {
      Object.entries(report.checks).forEach(([name, check]: [string, any]) => {
          if (check.status !== "ok") {
              const isConfig = name === "configuration";
              issues.push({
                  title: name.charAt(0).toUpperCase() + name.slice(1),
                  message: check.message || "Unknown issue",
                  type: isConfig ? "critical" : "warning",
                  diff: check.diff
              });
          }
      });
  }

  if (issues.length === 0) return null;

  const primaryIssue = issues[0];
  const criticalCount = issues.filter(i => i.type === "critical").length;
  const isCritical = criticalCount > 0 || !!error;

  return (
    <>
      <div className={cn(
          "w-full px-4 py-2 flex items-center justify-between transition-colors border-b shadow-sm z-40 relative",
          isCritical ? "bg-red-500/10 border-red-500/20 text-red-700 dark:text-red-400" : "bg-amber-500/10 border-amber-500/20 text-amber-700 dark:text-amber-400"
      )}>
          <div className="flex items-center gap-3 text-sm font-medium overflow-hidden">
              {isCritical ? <XCircle className="h-4 w-4 shrink-0" /> : <AlertTriangle className="h-4 w-4 shrink-0" />}
              <span className="truncate">
                  System Status: {isCritical ? "Critical" : "Degraded"}
                  <span className="font-normal opacity-80 ml-2 hidden sm:inline truncate">
                      — {primaryIssue.title}: {primaryIssue.message}
                  </span>
              </span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsOpen(true)}
            className={cn("h-7 text-xs gap-1 hover:bg-white/20 shrink-0", isCritical ? "hover:text-red-800 dark:hover:text-red-300" : "hover:text-amber-800 dark:hover:text-amber-300")}
          >
              View Details <ChevronRight className="h-3 w-3" />
          </Button>
      </div>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-2xl max-h-[85vh] flex flex-col p-0 gap-0 overflow-hidden">
          <DialogHeader className="p-6 pb-2 border-b bg-muted/10">
            <DialogTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-muted-foreground" />
                System Diagnostics
            </DialogTitle>
            <DialogDescription>
                The system is experiencing issues that may affect functionality.
            </DialogDescription>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto p-6 space-y-6">
              {issues.map((issue, idx) => (
                  <div key={idx} className="space-y-3">
                      <div className="flex items-start gap-2">
                          <Badge variant={issue.type === "critical" ? "destructive" : "secondary"}>
                              {issue.title}
                          </Badge>
                          <p className="text-sm text-foreground/80 mt-0.5 leading-tight">{issue.message}</p>
                      </div>
                      {issue.diff && (
                          <div className="ml-1 border-l-2 border-muted pl-4">
                              <div className="flex items-center gap-2 mb-2 text-xs text-muted-foreground font-medium uppercase tracking-wider">
                                  <FileDiff className="h-3 w-3" /> Configuration Changes
                              </div>
                              <UnifiedDiffViewer diff={issue.diff} />
                          </div>
                      )}
                  </div>
              ))}
          </div>

          <DialogFooter className="p-4 border-t bg-muted/10 gap-2 sm:gap-0">
            <Button variant="ghost" onClick={() => setIsOpen(false)}>Dismiss</Button>
            <Button onClick={fetchHealth} disabled={loading} className="gap-2">
                <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
                Retry Health Check
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
