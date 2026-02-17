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

const BUILTIN_TEMPLATES = [
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
             const defaults: Record<string, string> = {};
             if (s.configurationSchema && s.configurationSchema.properties) {
                 Object.entries(s.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
                     if (prop.default !== undefined) {
                         defaults[key] = String(prop.default);
                     } else {
                         defaults[key] = ""; // Initialize with empty string for controlled inputs
                     }
                 });
             }

            return {
                id: s.id,
                name: s.name,
                description: s.description,
                config: {
                    commandLineService: {
                        command: s.command,
                        env: {}, // Will be populated from params
                        workingDirectory: ''
                    },
                    openapiService: undefined,
                    configurationSchema: JSON.stringify(s.configurationSchema)
                },
                params: defaults
            };
        });

        return [...BUILTIN_TEMPLATES, ...registryTemplates];
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
