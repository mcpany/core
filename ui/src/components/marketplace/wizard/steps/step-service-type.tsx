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

// Define the type for our wizard templates
interface WizardTemplate {
    id: string;
    name: string;
    description: string;
    config: {
        commandLineService?: {
            command: string;
            env: Record<string, any>;
            workingDirectory: string;
        };
        openapiService?: {
            address: string;
            specUrl: string;
            specContent: string;
            tools: any[];
        };
        configurationSchema?: string;
    };
    params: Record<string, string>;
}

// 1. Manual Template
const MANUAL_TEMPLATE: WizardTemplate = {
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
        configurationSchema: undefined // Clear schema
    },
    params: {}
};

// 2. OpenAPI Template
const OPENAPI_TEMPLATE: WizardTemplate = {
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
        configurationSchema: undefined // Clear schema
    },
    params: {}
};

// 3. Map Service Registry to Templates
const REGISTRY_TEMPLATES: WizardTemplate[] = SERVICE_REGISTRY.map(service => {
    // Extract default params from schema
    const defaultParams: Record<string, string> = {};
    if (service.configurationSchema && service.configurationSchema.properties) {
        Object.entries(service.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
            if (prop.default !== undefined) {
                defaultParams[key] = String(prop.default);
            }
        });
    }

    return {
        id: service.id,
        name: service.name,
        description: service.description,
        config: {
            commandLineService: {
                command: service.command,
                env: {}, // Will be filled from params later
                workingDirectory: ''
            },
            configurationSchema: JSON.stringify(service.configurationSchema),
            openapiService: undefined
        },
        params: defaultParams
    };
});

// Combine all templates
const ALL_TEMPLATES = [MANUAL_TEMPLATE, ...REGISTRY_TEMPLATES, OPENAPI_TEMPLATE];

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
            updateState({
                selectedTemplateId: val,
                params: template.params
            });
            updateConfig({
                ...template.config as any,
                name: config.name || template.name,
            });
        }
    };

    const selectedTemplate = ALL_TEMPLATES.find(t => t.id === (selectedTemplateId || 'manual'));

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
                        <SelectItem value="manual">Manual / Custom</SelectItem>
                        <SelectItem value="openapi">OpenAPI / Swagger Import</SelectItem>
                        {REGISTRY_TEMPLATES.map(t => (
                            <SelectItem key={t.id} value={t.id}>
                                {t.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>

                {selectedTemplate && (
                    <Card className="mt-2 bg-muted/50">
                        <CardHeader>
                            <CardTitle className="text-base">
                                {selectedTemplate.name}
                            </CardTitle>
                            <CardDescription>
                                {selectedTemplate.description}
                            </CardDescription>
                        </CardHeader>
                        {selectedTemplate.config.commandLineService?.command && (
                             <CardContent className="pt-0">
                                <code className="text-xs bg-muted p-1 rounded break-all">
                                    {selectedTemplate.config.commandLineService.command}
                                </code>
                             </CardContent>
                        )}
                    </Card>
                )}
            </div>
        </div>
    );
}
