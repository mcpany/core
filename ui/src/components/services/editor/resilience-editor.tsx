/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { ResilienceConfig } from "@/lib/client"; // Note: Types are from proto, but we handle durations as strings for REST
import { Separator } from "@/components/ui/separator";

interface ResilienceEditorProps {
    resilience: ResilienceConfig | undefined;
    onChange: (resilience: ResilienceConfig | undefined) => void;
}

// Helper type to treat Durations as strings for UI editing
type UIResilienceConfig = {
    timeout?: string;
    circuitBreaker?: {
        failureRateThreshold: number;
        consecutiveFailures: number;
        openDuration?: string;
        halfOpenRequests: number;
    };
    retryPolicy?: {
        numberOfRetries: number;
        baseBackoff?: string;
        maxBackoff?: string;
        maxElapsedTime?: string;
    };
};

export function ResilienceEditor({ resilience, onChange }: ResilienceEditorProps) {
    // Cast to UI type for easier handling
    const config = (resilience || {}) as unknown as UIResilienceConfig;

    const updateConfig = (updates: Partial<UIResilienceConfig>) => {
        onChange({ ...config, ...updates } as unknown as ResilienceConfig);
    };

    const updateCircuitBreaker = (updates: Partial<NonNullable<UIResilienceConfig['circuitBreaker']>>) => {
        updateConfig({
            circuitBreaker: {
                failureRateThreshold: 0.5,
                consecutiveFailures: 5,
                halfOpenRequests: 1,
                openDuration: "10s",
                ...config.circuitBreaker,
                ...updates
            }
        });
    };

    const updateRetryPolicy = (updates: Partial<NonNullable<UIResilienceConfig['retryPolicy']>>) => {
        updateConfig({
            retryPolicy: {
                numberOfRetries: 3,
                baseBackoff: "1s",
                maxBackoff: "10s",
                maxElapsedTime: "30s",
                ...config.retryPolicy,
                ...updates
            }
        });
    };

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle>Timeouts</CardTitle>
                    <CardDescription>
                        Control how long to wait for a response from the upstream service.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="space-y-2">
                        <Label htmlFor="timeout">Request Timeout</Label>
                        <Input
                            id="timeout"
                            placeholder="30s"
                            value={config.timeout || ""}
                            onChange={(e) => updateConfig({ timeout: e.target.value })}
                        />
                        <p className="text-[10px] text-muted-foreground">
                            Format: number + unit (e.g., 5s, 100ms, 1m).
                        </p>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <div className="space-y-1">
                        <CardTitle>Retry Policy</CardTitle>
                        <CardDescription>
                            Automatically retry failed requests.
                        </CardDescription>
                    </div>
                    <Switch
                        checked={!!config.retryPolicy}
                        onCheckedChange={(checked) => {
                            if (checked) {
                                updateRetryPolicy({}); // Set defaults
                            } else {
                                const newConfig = { ...config };
                                delete newConfig.retryPolicy;
                                updateConfig(newConfig); // Remove policy
                            }
                        }}
                    />
                </CardHeader>
                {config.retryPolicy && (
                    <CardContent className="space-y-4 pt-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="retries">Max Retries</Label>
                                <Input
                                    id="retries"
                                    type="number"
                                    value={config.retryPolicy.numberOfRetries}
                                    onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="base-backoff">Base Backoff</Label>
                                <Input
                                    id="base-backoff"
                                    placeholder="1s"
                                    value={config.retryPolicy.baseBackoff || ""}
                                    onChange={(e) => updateRetryPolicy({ baseBackoff: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="max-backoff">Max Backoff</Label>
                                <Input
                                    id="max-backoff"
                                    placeholder="10s"
                                    value={config.retryPolicy.maxBackoff || ""}
                                    onChange={(e) => updateRetryPolicy({ maxBackoff: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="max-elapsed">Max Elapsed Time</Label>
                                <Input
                                    id="max-elapsed"
                                    placeholder="30s"
                                    value={config.retryPolicy.maxElapsedTime || ""}
                                    onChange={(e) => updateRetryPolicy({ maxElapsedTime: e.target.value })}
                                />
                            </div>
                        </div>
                    </CardContent>
                )}
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <div className="space-y-1">
                        <CardTitle>Circuit Breaker</CardTitle>
                        <CardDescription>
                            Prevent cascading failures by stopping requests to failing services.
                        </CardDescription>
                    </div>
                    <Switch
                        checked={!!config.circuitBreaker}
                        onCheckedChange={(checked) => {
                            if (checked) {
                                updateCircuitBreaker({}); // Set defaults
                            } else {
                                const newConfig = { ...config };
                                delete newConfig.circuitBreaker;
                                updateConfig(newConfig);
                            }
                        }}
                    />
                </CardHeader>
                {config.circuitBreaker && (
                    <CardContent className="space-y-4 pt-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="failure-threshold">Failure Rate Threshold (0.0 - 1.0)</Label>
                                <Input
                                    id="failure-threshold"
                                    type="number"
                                    step="0.1"
                                    min="0"
                                    max="1"
                                    value={config.circuitBreaker.failureRateThreshold}
                                    onChange={(e) => updateCircuitBreaker({ failureRateThreshold: parseFloat(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="consecutive-failures">Consecutive Failures</Label>
                                <Input
                                    id="consecutive-failures"
                                    type="number"
                                    value={config.circuitBreaker.consecutiveFailures}
                                    onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="open-duration">Open Duration</Label>
                                <Input
                                    id="open-duration"
                                    placeholder="10s"
                                    value={config.circuitBreaker.openDuration || ""}
                                    onChange={(e) => updateCircuitBreaker({ openDuration: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="half-open">Half-Open Requests</Label>
                                <Input
                                    id="half-open"
                                    type="number"
                                    value={config.circuitBreaker.halfOpenRequests}
                                    onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                        </div>
                    </CardContent>
                )}
            </Card>
        </div>
    );
}
