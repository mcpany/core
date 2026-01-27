/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

interface TrafficPolicyEditorProps {
    service: UpstreamServiceConfig;
    onChange: (updates: Partial<UpstreamServiceConfig>) => void;
}

// Helper to safely handle Duration which might be string (JSON) or object (gRPC)
// For UI editing, we want string.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getDurationString = (d: any): string => {
    if (typeof d === 'string') return d;
    if (d && typeof d === 'object') {
        // approximate for display if we ever get object
        // TODO: Better formatting if needed
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return (d as any).seconds ? `${(d as any).seconds}s` : '';
    }
    return '';
};

export function TrafficPolicyEditor({ service, onChange }: TrafficPolicyEditorProps) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const updateResilience = (updates: any) => {
        onChange({
            resilience: {
                ...service.resilience,
                ...updates
            }
        });
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const updateCircuitBreaker = (updates: any) => {
        onChange({
            resilience: {
                ...service.resilience,
                circuitBreaker: {
                    ...service.resilience?.circuitBreaker,
                    ...updates
                } as any // Cast to satisfy proto partial types
            }
        });
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const updateRetryPolicy = (updates: any) => {
        onChange({
            resilience: {
                ...service.resilience,
                retryPolicy: {
                    ...service.resilience?.retryPolicy,
                    ...updates
                } as any
            }
        });
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const updateRateLimit = (updates: any) => {
        onChange({
            rateLimit: {
                ...service.rateLimit,
                ...updates
            } as any
        });
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const updateConnectionPool = (updates: any) => {
        onChange({
            connectionPool: {
                ...service.connectionPool,
                ...updates
            } as any
        });
    };

    return (
        <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Rate Limiting */}
                <Card>
                    <CardHeader>
                        <CardTitle>Rate Limiting</CardTitle>
                        <CardDescription>Control the traffic volume to the upstream service.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label htmlFor="rl-enabled">Enable Rate Limiting</Label>
                            <Switch
                                id="rl-enabled"
                                checked={service.rateLimit?.isEnabled || false}
                                onCheckedChange={(checked) => updateRateLimit({ isEnabled: checked })}
                            />
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="rl-rps">Requests Per Second</Label>
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
                                <Label htmlFor="rl-burst">Burst Capacity</Label>
                                <Input
                                    id="rl-burst"
                                    type="number"
                                    min="0"
                                    value={service.rateLimit?.burst?.toString() || "0"}
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
                        <CardDescription>Manage connection reuse and concurrency.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="cp-max-conns">Max Connections</Label>
                                <Input
                                    id="cp-max-conns"
                                    type="number"
                                    min="0"
                                    value={service.connectionPool?.maxConnections || 0}
                                    onChange={(e) => updateConnectionPool({ maxConnections: parseInt(e.target.value) })}
                                    placeholder="0 (Unlimited)"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cp-max-idle">Max Idle Connections</Label>
                                <Input
                                    id="cp-max-idle"
                                    type="number"
                                    min="0"
                                    value={service.connectionPool?.maxIdleConnections || 0}
                                    onChange={(e) => updateConnectionPool({ maxIdleConnections: parseInt(e.target.value) })}
                                />
                            </div>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="cp-idle-timeout">Idle Timeout</Label>
                            <Input
                                id="cp-idle-timeout"
                                value={getDurationString(service.connectionPool?.idleTimeout)}
                                onChange={(e) => updateConnectionPool({ idleTimeout: e.target.value })}
                                placeholder="Example: 1m30s"
                            />
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Resilience */}
            <Card>
                <CardHeader>
                    <CardTitle>Resilience</CardTitle>
                    <CardDescription>Configure timeouts, retries, and circuit breaking.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    {/* Timeout */}
                    <div className="space-y-2">
                        <Label htmlFor="res-timeout">Request Timeout</Label>
                        <Input
                            id="res-timeout"
                            value={getDurationString(service.resilience?.timeout)}
                            onChange={(e) => updateResilience({ timeout: e.target.value })}
                            placeholder="Example: 30s"
                            className="max-w-md"
                        />
                    </div>

                    <Separator />

                    {/* Circuit Breaker */}
                    <div className="space-y-4">
                        <h4 className="text-sm font-medium">Circuit Breaker</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="cb-threshold">Failure Threshold (%)</Label>
                                <Input
                                    id="cb-threshold"
                                    type="number"
                                    min="0"
                                    max="1"
                                    step="0.1"
                                    value={service.resilience?.circuitBreaker?.failureRateThreshold || 0}
                                    onChange={(e) => updateCircuitBreaker({ failureRateThreshold: parseFloat(e.target.value) })}
                                    placeholder="0.5"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cb-failures">Consecutive Failures</Label>
                                <Input
                                    id="cb-failures"
                                    type="number"
                                    min="0"
                                    value={service.resilience?.circuitBreaker?.consecutiveFailures || 0}
                                    onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) })}
                                    placeholder="5"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cb-open-duration">Open Duration</Label>
                                <Input
                                    id="cb-open-duration"
                                    value={getDurationString(service.resilience?.circuitBreaker?.openDuration)}
                                    onChange={(e) => updateCircuitBreaker({ openDuration: e.target.value })}
                                    placeholder="10s"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="cb-half-open">Half-Open Requests</Label>
                                <Input
                                    id="cb-half-open"
                                    type="number"
                                    min="0"
                                    value={service.resilience?.circuitBreaker?.halfOpenRequests || 0}
                                    onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) })}
                                    placeholder="1"
                                />
                            </div>
                        </div>
                    </div>

                    <Separator />

                    {/* Retry Policy */}
                    <div className="space-y-4">
                        <h4 className="text-sm font-medium">Retry Policy</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="rp-retries">Max Retries</Label>
                                <Input
                                    id="rp-retries"
                                    type="number"
                                    min="0"
                                    value={service.resilience?.retryPolicy?.numberOfRetries || 0}
                                    onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) })}
                                    placeholder="3"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="rp-base">Base Backoff</Label>
                                <Input
                                    id="rp-base"
                                    value={getDurationString(service.resilience?.retryPolicy?.baseBackoff)}
                                    onChange={(e) => updateRetryPolicy({ baseBackoff: e.target.value })}
                                    placeholder="100ms"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="rp-max-backoff">Max Backoff</Label>
                                <Input
                                    id="rp-max-backoff"
                                    value={getDurationString(service.resilience?.retryPolicy?.maxBackoff)}
                                    onChange={(e) => updateRetryPolicy({ maxBackoff: e.target.value })}
                                    placeholder="1s"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="rp-max-elapsed">Max Elapsed Time</Label>
                                <Input
                                    id="rp-max-elapsed"
                                    value={getDurationString(service.resilience?.retryPolicy?.maxElapsedTime)}
                                    onChange={(e) => updateRetryPolicy({ maxElapsedTime: e.target.value })}
                                    placeholder="5s"
                                />
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
