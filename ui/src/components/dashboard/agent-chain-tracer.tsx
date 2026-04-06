/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Activity, ShieldCheck, ChevronRight, CheckCircle2, Server, Lock, Fingerprint } from "lucide-react";
import { cn } from "@/lib/utils";

// Mock data for the A2A Agent Chain tracer
import { useTraces } from "@/hooks/use-traces";
import { format } from "date-fns";

/**
 * Summary: Provides a hardware-attested timeline interface for tracking multi-agent handoffs.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.JSX.Element: The rendered AgentChainTracer interactive component.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export function AgentChainTracer() {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const { traces } = useTraces();

  const toggleExpand = (id: string) => {
    setExpandedId(expandedId === id ? null : id);
  };

  const chainData = traces.slice(0, 5).map((trace) => {
    let status = "active";
    if (trace.status === "success") status = "attested";
    else if (trace.status === "error") status = "speculative";

    let details = "No details provided.";
    if (trace.rootSpan?.errorMessage) {
      details = trace.rootSpan.errorMessage;
    } else if (trace.rootSpan?.input) {
      try {
        details = typeof trace.rootSpan.input === "string" ? trace.rootSpan.input : JSON.stringify(trace.rootSpan.input);
      } catch (e) {}
    }

    return {
      id: trace.id,
      agent: trace.rootSpan?.serviceName || trace.rootSpan?.name || "Unknown-Agent",
      action: trace.rootSpan?.name || "Unknown Action",
      status,
      latency: `${trace.totalDuration || 0}ms`,
      hash: trace.id ? trace.id.substring(0, 12) : "0x000",
      details,
      timestamp: trace.timestamp ? format(new Date(trace.timestamp), "HH:mm:ss.SSS") : ""
    };
  });

  return (
    <Card className="col-span-full xl:col-span-2 overflow-hidden border-border/50 bg-background/50 backdrop-blur-xl shadow-lg transition-all duration-300">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <div className="space-y-1">
          <CardTitle className="text-xl font-semibold flex items-center gap-2">
            <Activity className="h-5 w-5 text-indigo-500" />
            Agent Chain Tracer (A2A)
          </CardTitle>
          <CardDescription>
            Hardware-attested visualization of multi-agent task handoffs and reasoning chains.
          </CardDescription>
        </div>
        <div className="flex gap-2">
          <Badge variant="outline" className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">
            <ShieldCheck className="w-3 h-3 mr-1" />
            TPM Signed
          </Badge>
          <Badge variant="outline" className="bg-indigo-500/10 text-indigo-500 border-indigo-500/20">
            <Activity className="w-3 h-3 mr-1" />
            Live Trace
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pt-4">
        <div className="relative pl-6 border-l-2 border-muted-foreground/20 space-y-6">
          {chainData.map((step, index) => (
            <div key={step.id} className="relative group">
              {/* Timeline dot */}
              <div
                className={cn(
                  "absolute -left-[31px] top-1.5 w-4 h-4 rounded-full border-2 bg-background z-10 transition-colors duration-300",
                  step.status === "attested" ? "border-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" :
                  step.status === "speculative" ? "border-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)]" :
                  "border-indigo-500 shadow-[0_0_8px_rgba(99,102,241,0.5)] animate-pulse"
                )}
              />

              <div
                className={cn(
                  "rounded-lg border bg-card/40 p-3 transition-all duration-200 hover:shadow-md cursor-pointer",
                  expandedId === step.id ? "bg-card/80 border-indigo-500/30" : "hover:border-border"
                )}
                onClick={() => toggleExpand(step.id)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Server className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium text-sm">{step.agent}</span>
                    <ChevronRight className="h-3 w-3 text-muted-foreground" />
                    <span className="text-sm text-foreground/80">{step.action}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-xs text-muted-foreground">{step.timestamp}</span>
                    <Badge variant="secondary" className="text-xs font-mono">{step.latency}</Badge>
                  </div>
                </div>

                {expandedId === step.id && (
                  <div className="mt-4 pt-4 border-t border-border/50 grid grid-cols-2 gap-4 animate-in slide-in-from-top-2 fade-in duration-200">
                    <div className="space-y-2">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider font-semibold">
                        <Fingerprint className="h-3 w-3" />
                        Cryptographic Provenance
                      </div>
                      <div className="font-mono text-xs bg-muted/50 p-2 rounded-md border text-foreground/70">
                        {step.hash}
                      </div>
                    </div>
                    <div className="space-y-2">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider font-semibold">
                        <Lock className="h-3 w-3" />
                        Attestation State
                      </div>
                      <div className="flex items-center gap-2 text-sm">
                        {step.status === "attested" && (
                          <><CheckCircle2 className="h-4 w-4 text-emerald-500" /> <span className="text-emerald-600 dark:text-emerald-400 font-medium">Verified by QBS Hub</span></>
                        )}
                        {step.status === "speculative" && (
                          <><Activity className="h-4 w-4 text-amber-500" /> <span className="text-amber-600 dark:text-amber-400 font-medium">Pending Consensus</span></>
                        )}
                        {step.status === "active" && (
                          <><Activity className="h-4 w-4 text-indigo-500 animate-spin-slow" /> <span className="text-indigo-600 dark:text-indigo-400 font-medium">Computing...</span></>
                        )}
                      </div>
                    </div>
                    <div className="col-span-2 mt-2">
                      <p className="text-sm text-muted-foreground leading-relaxed">{step.details}</p>
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
