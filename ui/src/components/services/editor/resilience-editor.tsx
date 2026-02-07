/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResilienceConfig, CircuitBreakerConfig, RetryConfig } from "@proto/config/v1/upstream_service";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface ResilienceEditorProps {
    config?: ResilienceConfig;
    onChange: (config: ResilienceConfig) => void;
}

/**
 * ResilienceEditor component for configuring timeout, retry policy, and circuit breaker settings.
 *
 * @param props - The component props.
 * @param props.config - The current resilience configuration.
 * @param props.onChange - Callback to update the configuration.
 * @returns The rendered ResilienceEditor component.
 */
export function ResilienceEditor({ config, onChange }: ResilienceEditorProps) {
    // Helper to update root config
    const updateConfig = (updates: Partial<ResilienceConfig>) => {
        onChange({ ...config, ...updates });
    };

    // Helper to update circuit breaker
    const updateCircuitBreaker = (updates: Partial<CircuitBreakerConfig> | undefined) => {
        if (updates === undefined) {
            updateConfig({ circuitBreaker: undefined });
            return;
        }
        const current = config?.circuitBreaker || {
            failureRateThreshold: 0.5,
            consecutiveFailures: 5,
            openDuration: "60s" as any, // Default string for duration
            halfOpenRequests: 3
        };
        updateConfig({ circuitBreaker: { ...current, ...updates } });
    };

    // Helper to update retry policy
    const updateRetryPolicy = (updates: Partial<RetryConfig> | undefined) => {
        if (updates === undefined) {
            updateConfig({ retryPolicy: undefined });
            return;
        }
        const current = config?.retryPolicy || {
            numberOfRetries: 3,
            baseBackoff: "1s" as any,
            maxBackoff: "10s" as any,
            maxElapsedTime: "30s" as any
        };
        updateConfig({ retryPolicy: { ...current, ...updates } });
    };

    return (
        <div className="space-y-6">
            {/* Global Timeout */}
            <div className="grid grid-cols-1 gap-4">
                <Card>
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base font-medium flex items-center gap-2">
                            Timeout
                            <Tooltip>
                                <TooltipTrigger>
                                    <Info className="h-4 w-4 text-muted-foreground" />
                                </TooltipTrigger>
                                <TooltipContent>
                                    The maximum duration for a request before it is cancelled (e.g. "30s", "1m").
                                </TooltipContent>
                            </Tooltip>
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center gap-4">
                            <Input
                                placeholder="30s"
                                value={(config?.timeout as unknown as string) || ""}
                                onChange={(e) => updateConfig({ timeout: e.target.value as any })}
                                className="max-w-[200px]"
                            />
                            <span className="text-xs text-muted-foreground">
                                Supported units: ns, us, ms, s, m, h.
                            </span>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Retry Policy */}
                <Card className={!config?.retryPolicy ? "opacity-75" : ""}>
                    <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                            <CardTitle className="text-base font-medium">Retry Policy</CardTitle>
                            <Switch
                                checked={!!config?.retryPolicy}
                                onCheckedChange={(checked) => updateRetryPolicy(checked ? {} : undefined)}
                            />
                        </div>
                        <CardDescription>
                            Automatically retry failed requests with exponential backoff.
                        </CardDescription>
                    </CardHeader>
                    {config?.retryPolicy && (
                        <CardContent className="space-y-4 animate-in slide-in-from-top-2 duration-200">
                            <div className="space-y-2">
                                <Label htmlFor="retry-count">Number of Retries</Label>
                                <Input
                                    id="retry-count"
                                    type="number"
                                    value={config.retryPolicy.numberOfRetries}
                                    onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="retry-base">Base Backoff</Label>
                                <Input
                                    id="retry-base"
                                    placeholder="1s"
                                    value={(config.retryPolicy.baseBackoff as unknown as string) || ""}
                                    onChange={(e) => updateRetryPolicy({ baseBackoff: e.target.value as any })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="retry-max">Max Backoff</Label>
                                <Input
                                    id="retry-max"
                                    placeholder="10s"
                                    value={(config.retryPolicy.maxBackoff as unknown as string) || ""}
                                    onChange={(e) => updateRetryPolicy({ maxBackoff: e.target.value as any })}
                                />
                            </div>
                        </CardContent>
                    )}
                </Card>

                {/* Circuit Breaker */}
                <Card className={!config?.circuitBreaker ? "opacity-75" : ""}>
                    <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                            <CardTitle className="text-base font-medium">Circuit Breaker</CardTitle>
                            <Switch
                                checked={!!config?.circuitBreaker}
                                onCheckedChange={(checked) => updateCircuitBreaker(checked ? {} : undefined)}
                            />
                        </div>
                        <CardDescription>
                            Stop sending requests to a failing service to allow it to recover.
                        </CardDescription>
                    </CardHeader>
                    {config?.circuitBreaker && (
                        <CardContent className="space-y-4 animate-in slide-in-from-top-2 duration-200">
                            <div className="space-y-2">
                                <Label htmlFor="cb-threshold">Failure Threshold (0.0 - 1.0)</Label>
                                <Input
                                    id="cb-threshold"
                                    type="number"
                                    step="0.1"
                                    min="0"
                                    max="1"
                                    value={config.circuitBreaker.failureRateThreshold}
                                    onChange={(e) => updateCircuitBreaker({ failureRateThreshold: parseFloat(e.target.value) || 0 })}
                                />
                                <p className="text-[10px] text-muted-foreground">e.g. 0.5 means 50% failure rate opens circuit.</p>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cb-failures">Consecutive Failures</Label>
                                <Input
                                    id="cb-failures"
                                    type="number"
                                    value={config.circuitBreaker.consecutiveFailures}
                                    onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cb-duration">Open Duration</Label>
                                <Input
                                    id="cb-duration"
                                    placeholder="60s"
                                    value={(config.circuitBreaker.openDuration as unknown as string) || ""}
                                    onChange={(e) => updateCircuitBreaker({ openDuration: e.target.value as any })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cb-half-open">Half-Open Requests</Label>
                                <Input
                                    id="cb-half-open"
                                    type="number"
                                    value={config.circuitBreaker.halfOpenRequests}
                                    onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                        </CardContent>
                    )}
                </Card>
            </div>
        </div>
    );
}
