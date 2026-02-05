/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ResilienceConfig } from "@proto/config/v1/upstream_service";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface ResilienceEditorProps {
    config?: ResilienceConfig;
    onChange: (config: ResilienceConfig) => void;
}

export function ResilienceEditor({ config, onChange }: ResilienceEditorProps) {
    const updateConfig = (updates: Partial<ResilienceConfig>) => {
        onChange({ ...config, ...updates });
    };

    const updateRetryPolicy = (updates: any) => {
        updateConfig({
            retryPolicy: {
                ...config?.retryPolicy,
                ...updates
            }
        } as any);
    };

    const updateCircuitBreaker = (updates: any) => {
        updateConfig({
            circuitBreaker: {
                ...config?.circuitBreaker,
                ...updates
            }
        } as any);
    };

    // Helper to handle duration inputs (currently simplified as string, but proto expects object or string in JSON)
    // The backend handles string "30s" for duration.
    // However, the typed object expects `Duration` which is `{ seconds: Long, nanos: number }`.
    // But sending a string in JSON usually works for Duration fields if the marshaler supports it.
    // Given we are using `fetch` with JSON body, passing a string for duration might work if we cast it,
    // or we might need to parse "30s" into { seconds: 30, nanos: 0 }.
    // Let's assume the UI Input holds string, and we cast it to any to satisfy the type checker for now,
    // as the backend JSON unmarshaler (protojson) supports string durations.
    // BUT `client.ts` uses `ResilienceConfig` type which expects `Duration` object.
    // If we pass string, TS will complain.
    // Let's use `any` cast for now to allow string input, assuming backend handles "10s".

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        Global Timeout
                        <Tooltip>
                            <TooltipTrigger><Info className="h-4 w-4 text-muted-foreground" /></TooltipTrigger>
                            <TooltipContent>The maximum duration for a request before it is cancelled.</TooltipContent>
                        </Tooltip>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-2">
                        <Label htmlFor="timeout">Timeout</Label>
                        <Input
                            id="timeout"
                            placeholder="30s"
                            value={(config?.timeout as any) || ""}
                            onChange={(e) => updateConfig({ timeout: e.target.value as any })}
                        />
                        <p className="text-xs text-muted-foreground">Example: 30s, 1m, 500ms</p>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        Retry Policy
                        <Tooltip>
                            <TooltipTrigger><Info className="h-4 w-4 text-muted-foreground" /></TooltipTrigger>
                            <TooltipContent>Configure how failed requests should be retried.</TooltipContent>
                        </Tooltip>
                    </CardTitle>
                    <CardDescription>Exponential backoff strategy for transient failures.</CardDescription>
                </CardHeader>
                <CardContent className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                        <Label htmlFor="retries">Max Retries</Label>
                        <Input
                            id="retries"
                            type="number"
                            min="0"
                            value={config?.retryPolicy?.numberOfRetries || 0}
                            onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) || 0 })}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="max-elapsed">Max Elapsed Time</Label>
                        <Input
                            id="max-elapsed"
                            placeholder="30s"
                            value={(config?.retryPolicy?.maxElapsedTime as any) || ""}
                            onChange={(e) => updateRetryPolicy({ maxElapsedTime: e.target.value })}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="base-backoff">Base Backoff</Label>
                        <Input
                            id="base-backoff"
                            placeholder="100ms"
                            value={(config?.retryPolicy?.baseBackoff as any) || ""}
                            onChange={(e) => updateRetryPolicy({ baseBackoff: e.target.value })}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="max-backoff">Max Backoff</Label>
                        <Input
                            id="max-backoff"
                            placeholder="1s"
                            value={(config?.retryPolicy?.maxBackoff as any) || ""}
                            onChange={(e) => updateRetryPolicy({ maxBackoff: e.target.value })}
                        />
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        Circuit Breaker
                        <Tooltip>
                            <TooltipTrigger><Info className="h-4 w-4 text-muted-foreground" /></TooltipTrigger>
                            <TooltipContent>Prevent cascading failures by stopping requests to failing services.</TooltipContent>
                        </Tooltip>
                    </CardTitle>
                    <CardDescription>Opens the circuit when failure threshold is reached.</CardDescription>
                </CardHeader>
                <CardContent className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                        <Label htmlFor="failure-threshold">Failure Rate Threshold (0.0 - 1.0)</Label>
                        <Input
                            id="failure-threshold"
                            type="number"
                            step="0.1"
                            min="0"
                            max="1"
                            value={config?.circuitBreaker?.failureRateThreshold || 0}
                            onChange={(e) => updateCircuitBreaker({ failureRateThreshold: parseFloat(e.target.value) || 0 })}
                        />
                        <p className="text-xs text-muted-foreground">0.5 = 50% failure rate</p>
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="consecutive-failures">Consecutive Failures</Label>
                        <Input
                            id="consecutive-failures"
                            type="number"
                            min="0"
                            value={config?.circuitBreaker?.consecutiveFailures || 0}
                            onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) || 0 })}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="open-duration">Open Duration</Label>
                        <Input
                            id="open-duration"
                            placeholder="30s"
                            value={(config?.circuitBreaker?.openDuration as any) || ""}
                            onChange={(e) => updateCircuitBreaker({ openDuration: e.target.value })}
                        />
                        <p className="text-xs text-muted-foreground">Time to wait before retrying (Half-Open).</p>
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="half-open-requests">Half-Open Requests</Label>
                        <Input
                            id="half-open-requests"
                            type="number"
                            min="0"
                            value={config?.circuitBreaker?.halfOpenRequests || 0}
                            onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) || 0 })}
                        />
                        <p className="text-xs text-muted-foreground">Requests allowed to test recovery.</p>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
