/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

/**
 * StepOpenAPI component.
 * Allows users to configure an OpenAPI/Swagger upstream service.
 * @returns The rendered component.
 */
export function StepOpenAPI() {
    const { state, updateConfig } = useWizard();
    const { config } = state;
    const openapi = config.openapiService || { address: '', specUrl: '', specContent: '', tools: [] };

    const updateOpenAPI = (updates: Partial<typeof openapi>) => {
        updateConfig({
            openapiService: {
                ...openapi,
                ...updates
            }
        });
    };

    return (
        <div className="space-y-6">
            <Card className="bg-muted/5">
                <CardHeader>
                    <CardTitle className="text-base">OpenAPI Specification</CardTitle>
                    <CardDescription>
                        Provide the OpenAPI (Swagger) specification to automatically discover tools.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <Tabs defaultValue="url" className="w-full">
                        <TabsList className="grid w-full grid-cols-2">
                            <TabsTrigger value="url">Spec URL</TabsTrigger>
                            <TabsTrigger value="content">Paste JSON/YAML</TabsTrigger>
                        </TabsList>
                        <TabsContent value="url" className="mt-4 space-y-2">
                            <Label htmlFor="spec-url">Specification URL</Label>
                            <Input
                                id="spec-url"
                                placeholder="https://api.example.com/openapi.json"
                                value={openapi.specUrl || ''}
                                onChange={e => updateOpenAPI({ specUrl: e.target.value })}
                            />
                            <p className="text-xs text-muted-foreground">
                                The URL where the OpenAPI specification (v2 or v3) is hosted.
                            </p>
                        </TabsContent>
                        <TabsContent value="content" className="mt-4 space-y-2">
                            <Label htmlFor="spec-content">Specification Content</Label>
                            <Textarea
                                id="spec-content"
                                placeholder='{"openapi": "3.0.0", ...}'
                                className="font-mono text-xs min-h-[200px]"
                                value={openapi.specContent || ''}
                                onChange={e => updateOpenAPI({ specContent: e.target.value })}
                            />
                            <p className="text-xs text-muted-foreground">
                                Paste the raw JSON or YAML content of the specification here.
                            </p>
                        </TabsContent>
                    </Tabs>
                </CardContent>
            </Card>

            <div className="space-y-2">
                <Label htmlFor="base-address">Base Address (Optional)</Label>
                <Input
                    id="base-address"
                    placeholder="https://api.example.com"
                    value={openapi.address || ''}
                    onChange={e => updateOpenAPI({ address: e.target.value })}
                />
                <p className="text-sm text-muted-foreground">
                    Override the server URL defined in the spec. Useful for testing or staging environments.
                </p>
            </div>
        </div>
    );
}
