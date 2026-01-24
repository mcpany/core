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
import { Slider } from "@/components/ui/slider";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CacheConfig, SemanticCacheConfig } from "@proto/config/v1/call";

interface CacheConfigEditorProps {
    config?: CacheConfig;
    onChange: (config: CacheConfig) => void;
}

export function CacheConfigEditor({ config, onChange }: CacheConfigEditorProps) {
    const safeConfig: CacheConfig = config || {
        isEnabled: false,
        ttl: undefined, // undefined usually implies 0 or default
        strategy: "lru",
        semanticConfig: undefined,
    };

    const update = (updates: Partial<CacheConfig>) => {
        onChange({ ...safeConfig, ...updates });
    };

    const updateSemantic = (updates: Partial<SemanticCacheConfig>) => {
        onChange({ ...safeConfig, semanticConfig: { ...safeConfig.semanticConfig!, ...updates } });
    };

    const getSemanticProvider = () => {
        if (safeConfig.semanticConfig?.openai) return 'openai';
        if (safeConfig.semanticConfig?.ollama) return 'ollama';
        if (safeConfig.semanticConfig?.http) return 'http';
        return 'openai';
    };

    const handleProviderChange = (provider: string) => {
        const current = safeConfig.semanticConfig || { similarityThreshold: 0.9, persistencePath: "" };
        const newConfig: SemanticCacheConfig = {
            ...current,
            // Defaults for required fields if they are missing on 'current' (shouldn't happen if we init correctly)
            similarityThreshold: current.similarityThreshold ?? 0.9,
            persistencePath: current.persistencePath ?? "",
            // Reset providers
            openai: undefined,
            ollama: undefined,
            http: undefined,
            provider: "", // deprecated but required by type?
            model: "", // deprecated but required by type?
            apiKey: undefined, // deprecated but required by type?
        };

        if (provider === 'openai') {
            newConfig.openai = {
                model: "text-embedding-3-small",
                apiKey: { plainText: "" } as any
            };
        }
        if (provider === 'ollama') {
            newConfig.ollama = {
                model: "nomic-embed-text",
                baseUrl: "http://localhost:11434"
            };
        }
        if (provider === 'http') {
            newConfig.http = {
                url: "",
                headers: {},
                bodyTemplate: "",
                responseJsonPath: ""
            };
        }

        update({ semanticConfig: newConfig });
    };

    return (
        <div className="space-y-4">
             <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div className="space-y-1">
                            <CardTitle className="text-base">Response Caching</CardTitle>
                            <CardDescription>Cache tool execution results to improve latency and reduce costs.</CardDescription>
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
                                <Label htmlFor="ttl">Time to Live (TTL)</Label>
                                <Input
                                    id="ttl"
                                    placeholder="e.g. 60s, 5m"
                                    defaultValue={(safeConfig.ttl as any) || "60s"}
                                    onChange={(e) => update({ ttl: e.target.value as any })}
                                />
                                <p className="text-xs text-muted-foreground">Duration to keep results in cache.</p>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="strategy">Strategy</Label>
                                <Select
                                    value={safeConfig.strategy || "lru"}
                                    onValueChange={(val) => update({ strategy: val })}
                                >
                                    <SelectTrigger id="strategy">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="lru">LRU (Least Recently Used)</SelectItem>
                                        <SelectItem value="lfu">LFU (Least Frequently Used)</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>
                    </CardContent>
                )}
            </Card>

            <Card>
                <CardHeader className="pb-3">
                     <div className="flex items-center justify-between">
                        <div className="space-y-1">
                            <CardTitle className="text-base">Semantic Caching</CardTitle>
                            <CardDescription>Use vector embeddings to cache similar queries.</CardDescription>
                        </div>
                         <Switch
                            checked={!!safeConfig.semanticConfig}
                            onCheckedChange={(checked) => {
                                if (checked) {
                                    handleProviderChange('openai'); // Default
                                } else {
                                    update({ semanticConfig: undefined });
                                }
                            }}
                        />
                    </div>
                </CardHeader>
                {safeConfig.semanticConfig && (
                    <CardContent className="space-y-4 pt-0">
                         <div className="space-y-4">
                            <div className="space-y-2">
                                <div className="flex justify-between">
                                    <Label>Similarity Threshold ({safeConfig.semanticConfig.similarityThreshold || 0.9})</Label>
                                </div>
                                <Slider
                                    defaultValue={[safeConfig.semanticConfig.similarityThreshold || 0.9]}
                                    max={1.0}
                                    min={0.0}
                                    step={0.01}
                                    onValueChange={(vals) => updateSemantic({ similarityThreshold: vals[0] })}
                                />
                                <p className="text-xs text-muted-foreground">Higher values require closer matches (0.0 - 1.0).</p>
                            </div>

                             <div className="space-y-2">
                                <Label>Embedding Provider</Label>
                                <Tabs value={getSemanticProvider()} onValueChange={handleProviderChange} className="w-full">
                                    <TabsList className="grid w-full grid-cols-3">
                                        <TabsTrigger value="openai">OpenAI</TabsTrigger>
                                        <TabsTrigger value="ollama">Ollama</TabsTrigger>
                                        <TabsTrigger value="http">Custom HTTP</TabsTrigger>
                                    </TabsList>
                                    <TabsContent value="openai" className="space-y-4 pt-4">
                                         <div className="grid gap-2">
                                            <Label>API Key</Label>
                                            <Input
                                                type="password"
                                                value={safeConfig.semanticConfig.openai?.apiKey?.plainText || ""}
                                                onChange={(e) => {
                                                    const openai = safeConfig.semanticConfig!.openai || { model: "", apiKey: {} as any };
                                                    updateSemantic({ openai: { ...openai, apiKey: { plainText: e.target.value } as any } });
                                                }}
                                                placeholder="sk-..."
                                            />
                                        </div>
                                         <div className="grid gap-2">
                                            <Label>Model</Label>
                                            <Input
                                                value={safeConfig.semanticConfig.openai?.model || "text-embedding-3-small"}
                                                onChange={(e) => {
                                                    const openai = safeConfig.semanticConfig!.openai || { model: "", apiKey: {} as any };
                                                    updateSemantic({ openai: { ...openai, model: e.target.value } });
                                                }}
                                            />
                                        </div>
                                    </TabsContent>
                                    <TabsContent value="ollama" className="space-y-4 pt-4">
                                        <div className="grid gap-2">
                                            <Label>Base URL</Label>
                                            <Input
                                                value={safeConfig.semanticConfig.ollama?.baseUrl || "http://localhost:11434"}
                                                onChange={(e) => {
                                                    const ollama = safeConfig.semanticConfig!.ollama || { model: "", baseUrl: "" };
                                                    updateSemantic({ ollama: { ...ollama, baseUrl: e.target.value } });
                                                }}
                                            />
                                        </div>
                                         <div className="grid gap-2">
                                            <Label>Model</Label>
                                            <Input
                                                value={safeConfig.semanticConfig.ollama?.model || "nomic-embed-text"}
                                                onChange={(e) => {
                                                    const ollama = safeConfig.semanticConfig!.ollama || { model: "", baseUrl: "" };
                                                    updateSemantic({ ollama: { ...ollama, model: e.target.value } });
                                                }}
                                            />
                                        </div>
                                    </TabsContent>
                                     <TabsContent value="http" className="space-y-4 pt-4">
                                         <div className="grid gap-2">
                                            <Label>URL</Label>
                                            <Input
                                                value={safeConfig.semanticConfig.http?.url || ""}
                                                onChange={(e) => {
                                                    const http = safeConfig.semanticConfig!.http || { url: "", headers: {}, bodyTemplate: "", responseJsonPath: "" };
                                                    updateSemantic({ http: { ...http, url: e.target.value } });
                                                }}
                                                placeholder="https://my-embedding-service/v1/embed"
                                            />
                                        </div>
                                    </TabsContent>
                                </Tabs>
                            </div>
                        </div>
                    </CardContent>
                )}
            </Card>
        </div>
    );
}
