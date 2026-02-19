/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { SERVICE_REGISTRY } from '@/lib/service-registry';

// Map Registry items to Template format
const REGISTRY_TEMPLATES = SERVICE_REGISTRY.map(s => ({
    id: s.id,
    name: s.name,
    description: s.description,
    config: {
        commandLineService: {
            command: s.command,
            env: {},
            workingDirectory: ''
        },
        // We attach the schema string so StepParameters can use it to render a form
        configurationSchema: JSON.stringify(s.configurationSchema),
        openapiService: undefined
    },
    params: {} // Defaults will be extracted dynamically
}));

const MANUAL_TEMPLATES = [
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
            configurationSchema: undefined
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
            configurationSchema: undefined
        },
        params: {}
    }
];

const TEMPLATES = [...MANUAL_TEMPLATES, ...REGISTRY_TEMPLATES];

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;


    const handleTemplateChange = (val: string) => {
        const template = TEMPLATES.find(t => t.id === val);
        if (template) {
            const params = { ...(template.params as Record<string, string>) };

            // Extract defaults from schema if present
            if (template.config.configurationSchema) {
                try {
                    const schema = JSON.parse(template.config.configurationSchema);
                    if (schema.properties) {
                        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        Object.entries(schema.properties).forEach(([k, v]: [string, any]) => {
                            if (v.default !== undefined) {
                                params[k] = String(v.default);
                            }
                        });
                    }
                } catch (e) {
                    console.error("Failed to parse schema for defaults", e);
                }
            }

            updateState({
                selectedTemplateId: val,
                params: params
            });

            // We cast to any because template.config might have extra fields or slight type mismatches
            // with Partial<UpstreamServiceConfig> but we trust our mapping.
            updateConfig({
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
                    <SelectContent>
                        {TEMPLATES.map(t => (
                            <SelectItem key={t.id} value={t.id}>
                                {t.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Card className="mt-2 bg-muted/50">
                    <CardHeader>
                        <CardTitle className="text-base">
                            {TEMPLATES.find(t => t.id === (selectedTemplateId || 'manual'))?.name}
                        </CardTitle>
                        <CardDescription>
                            {TEMPLATES.find(t => t.id === (selectedTemplateId || 'manual'))?.description}
                        </CardDescription>
                    </CardHeader>
                </Card>
            </div>
        </div>
    );
}
