/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { SERVICE_REGISTRY } from '@/lib/service-registry';

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
            openapiService: undefined
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
            commandLineService: undefined
        },
        params: {}
    }
];

// Map Service Registry items to Template format
const REGISTRY_TEMPLATES = SERVICE_REGISTRY.map(service => ({
    id: service.id,
    name: service.name,
    description: service.description,
    config: {
        commandLineService: {
            command: service.command,
            env: {},
            workingDirectory: ''
        },
        // Store schema string for StepParameters to use
        configurationSchema: JSON.stringify(service.configurationSchema),
        openapiService: undefined
    },
    params: {} // Will be populated from schema defaults on selection
}));

// Combine templates, filtering out duplicates if any ID conflicts (prefer registry)
const ALL_TEMPLATES = [...MANUAL_TEMPLATES, ...REGISTRY_TEMPLATES.filter(r => !MANUAL_TEMPLATES.find(m => m.id === r.id))];

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;


    const handleTemplateChange = (val: string) => {
        const template = ALL_TEMPLATES.find(t => t.id === val);
        if (template) {
            // Extract defaults from schema if available
            const defaultParams: Record<string, string> = {};
            if (template.config.configurationSchema) {
                try {
                    const schema = JSON.parse(template.config.configurationSchema);
                    if (schema.properties) {
                        Object.entries(schema.properties).forEach(([key, prop]: [string, any]) => {
                            if (prop.default !== undefined) {
                                defaultParams[key] = String(prop.default);
                            }
                        });
                    }
                } catch (e) {
                    console.error("Failed to parse schema defaults", e);
                }
            }

            updateState({
                selectedTemplateId: val,
                params: defaultParams
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
                    <SelectContent>
                        {ALL_TEMPLATES.map(t => (
                            <SelectItem key={t.id} value={t.id}>
                                {t.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Card className="mt-2 bg-muted/50">
                    <CardHeader>
                        <CardTitle className="text-base">
                            {ALL_TEMPLATES.find(t => t.id === (selectedTemplateId || 'manual'))?.name}
                        </CardTitle>
                        <CardDescription>
                            {ALL_TEMPLATES.find(t => t.id === (selectedTemplateId || 'manual'))?.description}
                        </CardDescription>
                    </CardHeader>
                </Card>
            </div>
        </div>
    );
}
