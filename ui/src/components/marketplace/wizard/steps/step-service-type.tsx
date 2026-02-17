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

// Define the template interface
interface Template {
    id: string;
    name: string;
    description: string;
    config: any;
    params: Record<string, string>;
}

const MANUAL_TEMPLATE: Template = {
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
};

const OPENAPI_TEMPLATE: Template = {
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
};

// Map Service Registry items to templates
const REGISTRY_TEMPLATES: Template[] = SERVICE_REGISTRY.map(s => ({
    id: s.id,
    name: s.name,
    description: s.description,
    config: {
        commandLineService: {
            command: s.command,
            env: {}, // Start empty, will be filled by params
            workingDirectory: ''
        },
        configurationSchema: JSON.stringify(s.configurationSchema),
        openapiService: undefined
    },
    params: {} // Defaults will be extracted on selection
}));

const TEMPLATES = [
    MANUAL_TEMPLATE,
    OPENAPI_TEMPLATE,
    ...REGISTRY_TEMPLATES
];

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
            // Extract defaults from schema if available
            const defaultParams: Record<string, string> = { ...template.params };

            if (template.config.configurationSchema) {
                try {
                    const schema = JSON.parse(template.config.configurationSchema);
                    if (schema && schema.properties) {
                         Object.entries(schema.properties).forEach(([key, prop]: [string, any]) => {
                            if (prop.default !== undefined) {
                                defaultParams[key] = String(prop.default);
                            } else if (!defaultParams[key]) {
                                // Initialize empty for required fields to make them visible in params list
                                defaultParams[key] = "";
                            }
                         });
                    }
                } catch (e) {
                    console.error("Failed to parse schema for defaults", e);
                }
            }

            // Also set env vars in config based on defaults
            const env: any = {};
            Object.entries(defaultParams).forEach(([k, v]) => {
                 if (v) env[k] = { plainText: v };
            });

            const newConfig = {
                ...template.config,
                name: config.name || template.name,
            };

            if (newConfig.commandLineService) {
                newConfig.commandLineService.env = env;
            }

            updateState({
                selectedTemplateId: val,
                params: defaultParams
            });
            updateConfig(newConfig);
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
