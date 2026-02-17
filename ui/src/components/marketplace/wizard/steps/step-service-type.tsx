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

// Define the structure for internal templates used by the wizard
interface WizardTemplate {
    id: string;
    name: string;
    description: string;
    config: any; // Using any for flexibility with UpstreamServiceConfig partials
    params: Record<string, string>;
}

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;

    // Merge static templates with dynamic registry items
    const templates = useMemo<WizardTemplate[]>(() => {
        const staticTemplates: WizardTemplate[] = [
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

        const dynamicTemplates: WizardTemplate[] = SERVICE_REGISTRY.map(item => {
            // Extract default values from schema properties
            const defaults: Record<string, string> = {};
            if (item.configurationSchema && item.configurationSchema.properties) {
                Object.entries(item.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
                    if (prop.default !== undefined) {
                        defaults[key] = String(prop.default);
                    } else {
                         // Initialize required fields with empty string so they show up in params map
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
                        env: {}, // To be populated from params
                        workingDirectory: ''
                    },
                    openapiService: undefined,
                    // Save schema string for StepParameters to use
                    configurationSchema: JSON.stringify(item.configurationSchema)
                },
                params: defaults
            };
        });

        return [...staticTemplates, ...dynamicTemplates];
    }, []);

    const handleTemplateChange = (val: string) => {
        const template = templates.find(t => t.id === val);
        if (template) {
            updateState({
                selectedTemplateId: val,
                params: template.params
            });
            updateConfig({
                ...template.config,
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
