/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ResilienceConfig } from "@proto/config/v1/upstream_service";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { formatDuration, parseDuration } from "@/lib/duration-utils";

interface ResilienceEditorProps {
    config: ResilienceConfig | undefined;
    onChange: (config: ResilienceConfig) => void;
}

/**
 * Component for editing resilience configuration settings (Circuit Breaker, Retry Policy, etc.).
 * @param props The component props.
 * @returns The rendered component.
 */
export function ResilienceEditor({ config, onChange }: ResilienceEditorProps) {
    const updateConfig = (updates: Partial<ResilienceConfig>) => {
        onChange({ ...config, ...updates });
    };

    // Helper for nested updates
    const updateCircuitBreaker = (updates: any) => {
        updateConfig({
            circuitBreaker: {
                ...(config?.circuitBreaker || { failureRateThreshold: 0, consecutiveFailures: 0, halfOpenRequests: 0 }),
                ...updates
            }
        });
    };

    const updateRetryPolicy = (updates: any) => {
        updateConfig({
            retryPolicy: {
                ...(config?.retryPolicy || { numberOfRetries: 0 }),
                ...updates
            }
        });
    };

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle>Global Settings</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-2">
                        <Label htmlFor="timeout">Request Timeout</Label>
                        <Input
                            id="timeout"
                            placeholder="30s"
                            value={formatDuration(config?.timeout)}
                            onChange={(e) => updateConfig({ timeout: parseDuration(e.target.value) })}
                        />
                        <p className="text-xs text-muted-foreground">Maximum time to wait for a response from the upstream service.</p>
                    </div>
                </CardContent>
            </Card>

            <Tabs defaultValue="circuit-breaker" className="w-full">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="circuit-breaker">Circuit Breaker</TabsTrigger>
                    <TabsTrigger value="retry-policy">Retry Policy</TabsTrigger>
                </TabsList>

                <TabsContent value="circuit-breaker">
                    <Card>
                        <CardHeader>
                            <CardTitle>Circuit Breaker</CardTitle>
                            <CardDescription>Protect your system from cascading failures by stopping requests to failing services.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="failure-threshold">Failure Rate Threshold</Label>
                                    <Input
                                        id="failure-threshold"
                                        type="number"
                                        step="0.1"
                                        min="0"
                                        max="1"
                                        placeholder="0.5"
                                        value={config?.circuitBreaker?.failureRateThreshold ?? ""}
                                        onChange={(e) => updateCircuitBreaker({ failureRateThreshold: parseFloat(e.target.value) })}
                                    />
                                    <p className="text-[10px] text-muted-foreground">Ratio of failures (0.0 to 1.0) to trigger open state.</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="consecutive-failures">Consecutive Failures</Label>
                                    <Input
                                        id="consecutive-failures"
                                        type="number"
                                        min="0"
                                        placeholder="5"
                                        value={config?.circuitBreaker?.consecutiveFailures ?? ""}
                                        onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) })}
                                    />
                                    <p className="text-[10px] text-muted-foreground">Minimum failures before checking threshold.</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="open-duration">Open Duration</Label>
                                    <Input
                                        id="open-duration"
                                        placeholder="60s"
                                        value={formatDuration(config?.circuitBreaker?.openDuration)}
                                        onChange={(e) => updateCircuitBreaker({ openDuration: parseDuration(e.target.value) })}
                                    />
                                    <p className="text-[10px] text-muted-foreground">How long the circuit remains open.</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="half-open-requests">Half-Open Requests</Label>
                                    <Input
                                        id="half-open-requests"
                                        type="number"
                                        min="0"
                                        placeholder="3"
                                        value={config?.circuitBreaker?.halfOpenRequests ?? ""}
                                        onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) })}
                                    />
                                    <p className="text-[10px] text-muted-foreground">Requests allowed to test recovery.</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="retry-policy">
                    <Card>
                        <CardHeader>
                            <CardTitle>Retry Policy</CardTitle>
                            <CardDescription>Automatically retry failed requests with exponential backoff.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="retries">Number of Retries</Label>
                                    <Input
                                        id="retries"
                                        type="number"
                                        min="0"
                                        placeholder="3"
                                        value={config?.retryPolicy?.numberOfRetries ?? ""}
                                        onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="max-elapsed">Max Elapsed Time</Label>
                                    <Input
                                        id="max-elapsed"
                                        placeholder="10s"
                                        value={formatDuration(config?.retryPolicy?.maxElapsedTime)}
                                        onChange={(e) => updateRetryPolicy({ maxElapsedTime: parseDuration(e.target.value) })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="base-backoff">Base Backoff</Label>
                                    <Input
                                        id="base-backoff"
                                        placeholder="100ms"
                                        value={formatDuration(config?.retryPolicy?.baseBackoff)}
                                        onChange={(e) => updateRetryPolicy({ baseBackoff: parseDuration(e.target.value) })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="max-backoff">Max Backoff</Label>
                                    <Input
                                        id="max-backoff"
                                        placeholder="1s"
                                        value={formatDuration(config?.retryPolicy?.maxBackoff)}
                                        onChange={(e) => updateRetryPolicy({ maxBackoff: parseDuration(e.target.value) })}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
