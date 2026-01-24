/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { RateLimitConfig, RateLimitConfig_Storage, RateLimitConfig_KeyBy } from "@proto/config/v1/profile";

interface RateLimitEditorProps {
    config?: RateLimitConfig;
    onChange: (config: RateLimitConfig) => void;
}

export function RateLimitEditor({ config, onChange }: RateLimitEditorProps) {
    // Default to empty object if undefined
    const safeConfig: RateLimitConfig = config || {
        isEnabled: false,
        requestsPerSecond: 0,
        burst: 0 as any, // Cast to any to handle Long type mismatch
        storage: RateLimitConfig_Storage.STORAGE_MEMORY,
        keyBy: RateLimitConfig_KeyBy.KEY_BY_GLOBAL,
        redis: undefined,
        costMetric: 0,
        toolLimits: {}
    };

    const update = (updates: Partial<RateLimitConfig>) => {
        onChange({ ...safeConfig, ...updates });
    };

    return (
        <Card>
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <div className="space-y-1">
                        <CardTitle className="text-base">Rate Limiting</CardTitle>
                        <CardDescription>Protect your upstream service from excessive traffic.</CardDescription>
                    </div>
                    <Switch
                        checked={safeConfig.isEnabled}
                        onCheckedChange={(checked) => update({ isEnabled: checked })}
                    />
                </div>
            </CardHeader>
            {safeConfig.isEnabled && (
                <CardContent className="grid gap-4 pt-0">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="rps">Requests Per Second</Label>
                            <Input
                                id="rps"
                                type="number"
                                min="0"
                                step="0.1"
                                value={safeConfig.requestsPerSecond || 0}
                                onChange={(e) => update({ requestsPerSecond: parseFloat(e.target.value) })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="burst">Burst Capacity</Label>
                            <Input
                                id="burst"
                                type="number"
                                min="0"
                                value={safeConfig.burst?.toString() || "0"}
                                onChange={(e) => update({ burst: parseInt(e.target.value) as any })}
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="key-by">Limit By</Label>
                            <Select
                                value={safeConfig.keyBy.toString()}
                                onValueChange={(val) => update({ keyBy: parseInt(val) })}
                            >
                                <SelectTrigger id="key-by">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={RateLimitConfig_KeyBy.KEY_BY_GLOBAL.toString()}>Global (All Users)</SelectItem>
                                    <SelectItem value={RateLimitConfig_KeyBy.KEY_BY_IP.toString()}>IP Address</SelectItem>
                                    <SelectItem value={RateLimitConfig_KeyBy.KEY_BY_USER_ID.toString()}>User ID</SelectItem>
                                    <SelectItem value={RateLimitConfig_KeyBy.KEY_BY_API_KEY.toString()}>API Key</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                         <div className="space-y-2">
                            <Label htmlFor="storage">Storage Backend</Label>
                            <Select
                                value={safeConfig.storage.toString()}
                                onValueChange={(val) => update({ storage: parseInt(val) })}
                            >
                                <SelectTrigger id="storage">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={RateLimitConfig_Storage.STORAGE_MEMORY.toString()}>In-Memory</SelectItem>
                                    <SelectItem value={RateLimitConfig_Storage.STORAGE_REDIS.toString()}>Redis</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    {safeConfig.storage === RateLimitConfig_Storage.STORAGE_REDIS && (
                        <div className="space-y-2 p-4 border rounded-md bg-muted/20">
                            <h4 className="font-medium text-sm mb-2">Redis Configuration</h4>
                            <div className="grid gap-2">
                                <Label htmlFor="redis-addr">Address</Label>
                                <Input
                                    id="redis-addr"
                                    placeholder="localhost:6379"
                                    value={safeConfig.redis?.address || ""}
                                    onChange={(e) => update({ redis: { ...safeConfig.redis, address: e.target.value } as any })}
                                />
                                <Label htmlFor="redis-pass">Password</Label>
                                <Input
                                    id="redis-pass"
                                    type="password"
                                    placeholder="Optional"
                                    value={safeConfig.redis?.password || ""}
                                    onChange={(e) => update({ redis: { ...safeConfig.redis, password: e.target.value } as any })}
                                />
                            </div>
                        </div>
                    )}
                </CardContent>
            )}
        </Card>
    );
}
