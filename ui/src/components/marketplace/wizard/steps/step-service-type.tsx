/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
import { useWizard } from '../wizard-context';
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { SERVICE_REGISTRY } from '@/lib/service-registry';

// Static templates that are not in the registry (e.g. manual/custom)
const STATIC_TEMPLATES = [
    {
        id: 'manual',
        name: 'Manual / Custom',
        description: 'Configure everything from scratch.',
        config: {
            commandLineService: {
                command: '',
                env: {},
                workingDirectory: ''
            },
            openapiService: undefined,
            configurationSchema: ""
        },
        params: {}
    },
    {
        id: 'openapi',
        name: 'OpenAPI / Swagger Import',
        description: 'Import tools from an OpenAPI specification.',
        config: {
            openapiService: {
                address: "",
                specUrl: "",
                specContent: "",
                tools: []
            },
            commandLineService: undefined,
            configurationSchema: ""
        },
        params: {}
    }
];

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;

    // Merge static templates with registry
    const templates = useMemo(() => {
        const registryTemplates = SERVICE_REGISTRY.map(item => {
            // Extract defaults from schema
            const defaults: Record<string, string> = {};
            if (item.configurationSchema && item.configurationSchema.properties) {
                Object.entries(item.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
                    if (prop.default) {
                        defaults[key] = String(prop.default);
                    } else {
                        // Initialize empty string for required fields to hint existence
                        defaults[key] = "";
                    }
                });
            }

            // Build config env object from defaults
            const env: Record<string, any> = {};
            Object.keys(defaults).forEach(k => {
                env[k] = { plainText: defaults[k] };
            });

            return {
                id: item.id,
                name: item.name,
                description: item.description,
                config: {
                    commandLineService: {
                        command: item.command,
                        env: env,
                        workingDirectory: ''
                    },
                    openapiService: undefined,
                    configurationSchema: JSON.stringify(item.configurationSchema)
                },
                params: defaults
            };
        });

        // Filter out duplicates if any. Registry takes precedence.
        // We assume STATIC_TEMPLATES IDs (manual, openapi) don't conflict with registry IDs.
        return [...STATIC_TEMPLATES, ...registryTemplates];
    }, []);

    const handleTemplateChange = (val: string) => {
        const template = templates.find(t => t.id === val);
        if (template) {
            updateState({
                selectedTemplateId: val,
                params: template.params as Record<string, string>
            });
            updateConfig({
                ...template.config as any,
                name: config.name || template.name,
            });
        }
    };


    return (
        <div className="space-y-6">
            <div className="space-y-2">
                <Label htmlFor="service-name">Service Name</Label>
                <Input
                    id="service-name"
                    placeholder="e.g. My Postgres DB"
                    value={config.name || ''}
                    onChange={e => updateConfig({ name: e.target.value })}
                />
                <p className="text-sm text-muted-foreground">Unique identifier for this service.</p>
            </div>

            <div className="space-y-2">
                <Label htmlFor="service-template">Template</Label>
                <Select value={selectedTemplateId || 'manual'} onValueChange={handleTemplateChange}>
                    <SelectTrigger id="service-template">
                        <SelectValue placeholder="Select a template" />
                    </SelectTrigger>
                    <SelectContent className="max-h-[300px]">
                        {templates.map(t => (
                            <SelectItem key={t.id} value={t.id}>
                                {t.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Card className="mt-2 bg-muted/50">
                    <CardHeader>
                        <CardTitle className="text-base">
                            {templates.find(t => t.id === (selectedTemplateId || 'manual'))?.name}
                        </CardTitle>
                        <CardDescription>
                            {templates.find(t => t.id === (selectedTemplateId || 'manual'))?.description}
                        </CardDescription>
                    </CardHeader>
                </Card>
            </div>
        </div>
    );
}
