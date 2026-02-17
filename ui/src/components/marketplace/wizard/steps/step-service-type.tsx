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

// Define the static templates (Manual, OpenAPI)
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

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;

    // Merge Static templates with Registry templates
    const allTemplates = useMemo(() => {
        const registryTemplates = SERVICE_REGISTRY.map(item => {
            // Extract defaults from schema
            const defaults: Record<string, string> = {};
            if (item.configurationSchema && item.configurationSchema.properties) {
                Object.entries(item.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
                    if (prop.default !== undefined) {
                        defaults[key] = String(prop.default);
                    } else {
                        // Initialize empty string for required fields to improve UX
                         defaults[key] = "";
                    }
                });
            }

            return {
                id: item.id,
                name: item.name,
                description: item.description,
                config: {
                    commandLineService: {
                        command: item.command,
                        env: {}, // Will be filled from params
                        workingDirectory: ''
                    },
                    openapiService: undefined,
                    configurationSchema: JSON.stringify(item.configurationSchema)
                },
                params: defaults
            };
        });

        // Filter out any static templates that might duplicate registry IDs if we kept them
        return [...STATIC_TEMPLATES, ...registryTemplates];
    }, []);

    const handleTemplateChange = (val: string) => {
        const template = allTemplates.find(t => t.id === val);
        if (template) {
            // If the user already typed a name, keep it. If not (or if it matches old template name), suggest new one.
            const currentName = config.name || '';
            const newName = (!currentName || currentName === 'New Service') ? template.name : currentName;

            updateState({
                selectedTemplateId: val,
                params: template.params as Record<string, string>
            });

            const newConfig: any = {
                ...template.config,
                name: newName
            };

            // If switching TO commandLine, clear openapi
            if (template.config.commandLineService) {
                newConfig.openapiService = undefined;
            }
            // If switching TO openapi, clear commandLine
            if (template.config.openapiService) {
                newConfig.commandLineService = undefined;
            }

            updateConfig(newConfig);
        }
    };

    const selectedTemplate = allTemplates.find(t => t.id === (selectedTemplateId || 'manual'));

    return (
        <div className="space-y-6">
            <div className="space-y-2">
                <Label htmlFor="service-name">Service Name</Label>
                <Input
                    id="service-name"
                    placeholder="e.g. My Service"
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
                        {allTemplates.map(t => (
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
                    </Card>
                )}
            </div>
        </div>
    );
}
