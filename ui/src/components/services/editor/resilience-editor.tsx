/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Slider } from "@/components/ui/slider";
import { ResilienceConfig } from "@proto/config/v1/upstream_service";

interface ResilienceEditorProps {
    resilience?: ResilienceConfig;
    onChange: (resilience: ResilienceConfig) => void;
}

/**
 * Helper to convert duration to string (mock implementation if type is object).
 * In most proto-json configs, duration is a string like "10s".
 */
const durationToString = (d: any): string => {
    if (!d) return "";
    if (typeof d === 'string') return d;
    if (typeof d === 'object' && 'seconds' in d) return `${d.seconds}s`;
    return "";
};

/**
 * ResilienceEditor component.
 * Allows configuring Retry Policy, Circuit Breaker, and Timeouts.
 */
export function ResilienceEditor({ resilience, onChange }: ResilienceEditorProps) {
    // Local state to handle UI interactions before committing to parent
    // We initialize from props.
    // Note: We need to handle deep updates carefully.

    const updateCircuitBreaker = (updates: any) => {
        const current = resilience?.circuitBreaker || {};
        onChange({
            ...resilience,
            circuitBreaker: { ...current, ...updates }
        });
    };

    const updateRetryPolicy = (updates: any) => {
        const current = resilience?.retryPolicy || {};
        onChange({
            ...resilience,
            retryPolicy: { ...current, ...updates }
        });
    };

    const updateTimeout = (val: string) => {
        onChange({
            ...resilience,
            timeout: val as any // Assuming string is accepted by proto-json mapping
        });
    };

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle>Timeout</CardTitle>
                    <CardDescription>Maximum duration for a request before it is cancelled.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid w-full max-w-sm items-center gap-1.5">
                        <Label htmlFor="timeout">Global Timeout</Label>
                        <Input
                            id="timeout"
                            placeholder="e.g. 30s"
                            value={durationToString(resilience?.timeout)}
                            onChange={(e) => updateTimeout(e.target.value)}
                        />
                        <p className="text-[10px] text-muted-foreground">Format: 30s, 1m, 500ms</p>
                    </div>
                </CardContent>
            </Card>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle>Circuit Breaker</CardTitle>
                        <CardDescription>Prevents cascading failures by stopping requests to failing services.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label>Failure Rate Threshold ({((resilience?.circuitBreaker?.failureRateThreshold || 0.5) * 100).toFixed(0)}%)</Label>
                            <Slider
                                defaultValue={[resilience?.circuitBreaker?.failureRateThreshold || 0.5]}
                                max={1}
                                step={0.05}
                                onValueChange={(vals) => updateCircuitBreaker({ failureRateThreshold: vals[0] })}
                            />
                            <p className="text-[10px] text-muted-foreground">Open circuit if error rate exceeds this ratio.</p>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="consecutive-failures">Consecutive Failures</Label>
                            <Input
                                id="consecutive-failures"
                                type="number"
                                value={resilience?.circuitBreaker?.consecutiveFailures || 5}
                                onChange={(e) => updateCircuitBreaker({ consecutiveFailures: parseInt(e.target.value) || 0 })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="open-duration">Open Duration</Label>
                            <Input
                                id="open-duration"
                                placeholder="e.g. 60s"
                                value={durationToString(resilience?.circuitBreaker?.openDuration)}
                                onChange={(e) => updateCircuitBreaker({ openDuration: e.target.value })}
                            />
                            <p className="text-[10px] text-muted-foreground">Time to wait before testing connectivity (Half-Open).</p>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="half-open-requests">Half-Open Requests</Label>
                            <Input
                                id="half-open-requests"
                                type="number"
                                value={resilience?.circuitBreaker?.halfOpenRequests || 1}
                                onChange={(e) => updateCircuitBreaker({ halfOpenRequests: parseInt(e.target.value) || 1 })}
                            />
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Retry Policy</CardTitle>
                        <CardDescription>Automatically retry failed requests with exponential backoff.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="retries">Number of Retries</Label>
                            <Input
                                id="retries"
                                type="number"
                                value={resilience?.retryPolicy?.numberOfRetries || 0}
                                onChange={(e) => updateRetryPolicy({ numberOfRetries: parseInt(e.target.value) || 0 })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="base-backoff">Base Backoff</Label>
                            <Input
                                id="base-backoff"
                                placeholder="e.g. 100ms"
                                value={durationToString(resilience?.retryPolicy?.baseBackoff)}
                                onChange={(e) => updateRetryPolicy({ baseBackoff: e.target.value })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="max-backoff">Max Backoff</Label>
                            <Input
                                id="max-backoff"
                                placeholder="e.g. 5s"
                                value={durationToString(resilience?.retryPolicy?.maxBackoff)}
                                onChange={(e) => updateRetryPolicy({ maxBackoff: e.target.value })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="max-elapsed">Max Elapsed Time</Label>
                            <Input
                                id="max-elapsed"
                                placeholder="e.g. 10s"
                                value={durationToString(resilience?.retryPolicy?.maxElapsedTime)}
                                onChange={(e) => updateRetryPolicy({ maxElapsedTime: e.target.value })}
                            />
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
