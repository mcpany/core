/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { formatDuration, parseDuration } from "@/lib/duration-utils";
import { Separator } from "@/components/ui/separator";

interface TrafficPolicyEditorProps {
    service: UpstreamServiceConfig;
    onChange: (updates: Partial<UpstreamServiceConfig>) => void;
}

export function TrafficPolicyEditor({ service, onChange }: TrafficPolicyEditorProps) {

    // --- Rate Limit Helpers ---
    const updateRateLimit = (updates: any) => {
        onChange({
            rateLimit: {
                ...service.rateLimit,
                ...updates,
                // Ensure defaults if creating new
                isEnabled: updates.isEnabled !== undefined ? updates.isEnabled : service.rateLimit?.isEnabled ?? false,
                requestsPerSecond: updates.requestsPerSecond !== undefined ? updates.requestsPerSecond : service.rateLimit?.requestsPerSecond ?? 0,
                burst: updates.burst !== undefined ? updates.burst : service.rateLimit?.burst ?? 0,
                keyBy: service.rateLimit?.keyBy ?? 0,
                costMetric: service.rateLimit?.costMetric ?? 0,
                storage: service.rateLimit?.storage ?? 0,
                redis: service.rateLimit?.redis,
                toolLimits: service.rateLimit?.toolLimits || {},
            }
        });
    };

    // --- Resilience Helpers ---
    const updateResilience = (updates: any) => {
        const currentResilience = service.resilience || {
            timeout: undefined,
            retryPolicy: undefined,
            circuitBreaker: undefined
        };
        onChange({
            resilience: {
                ...currentResilience,
                ...updates
            }
        });
    };

    const updateRetryPolicy = (updates: any) => {
        const currentResilience = service.resilience || {};
        const currentRetry = currentResilience.retryPolicy || {
            numberOfRetries: 0,
            baseBackoff: undefined,
            maxBackoff: undefined,
            maxElapsedTime: undefined
        };

        onChange({
            resilience: {
                ...currentResilience,
                retryPolicy: {
                    ...currentRetry,
                    ...updates
                }
            }
        });
    };

    const updateCircuitBreaker = (updates: any) => {
        const currentResilience = service.resilience || {};
        const currentBreaker = currentResilience.circuitBreaker || {
            failureRateThreshold: 0,
            consecutiveFailures: 0,
            openDuration: undefined,
            halfOpenRequests: 0
        };

        onChange({
            resilience: {
                ...currentResilience,
                circuitBreaker: {
                    ...currentBreaker,
                    ...updates
                }
            }
        });
    };

    // --- Connection Pool Helpers ---
    const updateConnectionPool = (updates: any) => {
        const currentPool = service.connectionPool || {
            maxConnections: 0,
            maxIdleConnections: 0,
            idleTimeout: undefined
        };
        onChange({
            connectionPool: {
                ...currentPool,
                ...updates
            }
        });
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Rate Limiting */}
            <Card>
                <CardHeader>
                    <CardTitle>Rate Limiting</CardTitle>
                    <CardDescription>Protect the upstream service from excessive load.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center space-x-2">
                        <Switch
                            id="rl-enabled"
                            checked={service.rateLimit?.isEnabled || false}
                            onCheckedChange={(checked) => updateRateLimit({ isEnabled: checked })}
                        />
                        <Label htmlFor="rl-enabled">Enable Rate Limiting</Label>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="rl-rps">Requests / Sec</Label>
                            <Input
                                id="rl-rps"
                                type="number"
                                min="0"
                                step="0.1"
                                value={service.rateLimit?.requestsPerSecond || 0}
                                onChange={(e) => updateRateLimit({ requestsPerSecond: parseFloat(e.target.value) })}
                                disabled={!service.rateLimit?.isEnabled}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="rl-burst">Burst</Label>
                            <Input
                                id="rl-burst"
                                type="number"
                                min="0"
                                value={service.rateLimit?.burst || 0}
                                onChange={(e) => updateRateLimit({ burst: parseInt(e.target.value) })}
                                disabled={!service.rateLimit?.isEnabled}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Connection Pool */}
            <Card>
                <CardHeader>
                    <CardTitle>Connection Pool</CardTitle>
                    <CardDescription>Manage connection resources.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="pool-max">Max Connections</Label>
                            <Input
                                id="pool-max"
                                type="number"
                                min="0"
                                value={service.connectionPool?.maxConnections || 0}
                                onChange={(e) => updateConnectionPool({ maxConnections: parseInt(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="pool-idle">Max Idle</Label>
                            <Input
                                id="pool-idle"
                                type="number"
                                min="0"
                                value={service.connectionPool?.maxIdleConnections || 0}
                                onChange={(e) => updateConnectionPool({ maxIdleConnections: parseInt(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2 col-span-2">
                            <Label htmlFor="pool-timeout">Idle Timeout</Label>
                            <Input
                                id="pool-timeout"
                                placeholder="e.g. 10s"
                                value={formatDuration(service.connectionPool?.idleTimeout)}
                                onChange={(e) => updateConnectionPool({ idleTimeout: parseDuration(e.target.value) })}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Resilience: Timeout & Retry */}
            <Card>
                <CardHeader>
                    <CardTitle>Resilience: Retry</CardTitle>
                    <CardDescription>Handle transient failures.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="res-timeout">Global Timeout</Label>
                        <Input
                            id="res-timeout"
                            placeholder="e.g. 30s"
                            value={formatDuration(service.resilience?.timeout)}
                            onChange={(e) => updateResilience({ timeout: parseDuration(e.target.value) })}
                        />
                    </div>
                    <Separator />
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="retry-count">Max Retries</Label>
                            <Input
                                id="retry-count"
                                type="number"
                                min="0"
                                value={service.resilience?.retryPolicy?.numberOfRetries || 0}
                                onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="retry-max-time">Max Elapsed</Label>
                            <Input
                                id="retry-max-time"
                                placeholder="e.g. 10s"
                                value={formatDuration(service.resilience?.retryPolicy?.maxElapsedTime)}
                                onChange={(e) => updateRetryPolicy({ maxElapsedTime: parseDuration(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="retry-base">Base Backoff</Label>
                            <Input
                                id="retry-base"
                                placeholder="e.g. 100ms"
                                value={formatDuration(service.resilience?.retryPolicy?.baseBackoff)}
                                onChange={(e) => updateRetryPolicy({ baseBackoff: parseDuration(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="retry-max-backoff">Max Backoff</Label>
                            <Input
                                id="retry-max-backoff"
                                placeholder="e.g. 1s"
                                value={formatDuration(service.resilience?.retryPolicy?.maxBackoff)}
                                onChange={(e) => updateRetryPolicy({ maxBackoff: parseDuration(e.target.value) })}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Resilience: Circuit Breaker */}
            <Card>
                <CardHeader>
                    <CardTitle>Resilience: Circuit Breaker</CardTitle>
                    <CardDescription>Fail fast when upstream is down.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="cb-threshold">Failure Threshold (%)</Label>
                            <Input
                                id="cb-threshold"
                                type="number"
                                min="0"
                                max="100"
                                step="1"
                                placeholder="50"
                                value={(service.resilience?.circuitBreaker?.failureRateThreshold || 0) * 100}
                                onChange={(e) => updateCircuitBreaker({ failureRateThreshold: parseFloat(e.target.value) / 100 })}
                            />
                            <p className="text-xs text-muted-foreground">0-100%</p>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="cb-failures">Min Failures</Label>
                            <Input
                                id="cb-failures"
                                type="number"
                                min="1"
                                value={service.resilience?.circuitBreaker?.consecutiveFailures || 0}
                                onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="cb-open">Open Duration</Label>
                            <Input
                                id="cb-open"
                                placeholder="e.g. 1m"
                                value={formatDuration(service.resilience?.circuitBreaker?.openDuration)}
                                onChange={(e) => updateCircuitBreaker({ openDuration: parseDuration(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="cb-half">Half-Open Req</Label>
                            <Input
                                id="cb-half"
                                type="number"
                                min="1"
                                value={service.resilience?.circuitBreaker?.halfOpenRequests || 0}
                                onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) })}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
