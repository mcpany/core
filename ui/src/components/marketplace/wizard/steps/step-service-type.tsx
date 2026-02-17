/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
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

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;

    const templates = useMemo(() => {
        const registryTemplates = SERVICE_REGISTRY.map(s => {
            // Extract defaults from schema
            const params: Record<string, string> = {};
            const env: Record<string, any> = {};

            if (s.configurationSchema && s.configurationSchema.properties) {
                Object.entries(s.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
                     // Check if default exists, otherwise empty string
                     const defaultVal = prop.default !== undefined ? String(prop.default) : "";
                     params[key] = defaultVal;
                     // Initialize env config structure
                     env[key] = { plainText: defaultVal, validationRegex: "" };
                });
            }

            return {
                id: s.id,
                name: s.name,
                description: s.description,
                config: {
                    commandLineService: {
                        command: s.command,
                        env: env,
                        workingDirectory: ''
                    },
                    // Store the schema string so StepParameters can use it
                    configurationSchema: JSON.stringify(s.configurationSchema),
                    openapiService: undefined
                },
                params: params
            };
        });

        // Filter out manual templates that might conflict by ID (though unlikely with 'manual' and 'openapi')
        return [...MANUAL_TEMPLATES, ...registryTemplates];
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

    const selectedTemplate = templates.find(t => t.id === (selectedTemplateId || 'manual'));

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
                        <SelectItem value="manual" className="font-semibold">Manual / Custom</SelectItem>
                        <SelectItem value="openapi" className="font-semibold">OpenAPI / Swagger</SelectItem>
                        <div className="h-px bg-muted my-1" />
                        <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
                            Official Registry
                        </div>
                        {templates.filter(t => t.id !== 'manual' && t.id !== 'openapi').map(t => (
                            <SelectItem key={t.id} value={t.id}>
                                {t.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Card className="mt-2 bg-muted/50">
                    <CardHeader>
                        <CardTitle className="text-base">
                            {selectedTemplate?.name}
                        </CardTitle>
                        <CardDescription>
                            {selectedTemplate?.description}
                        </CardDescription>
                    </CardHeader>
                </Card>
            </div>
        </div>
    );
}
