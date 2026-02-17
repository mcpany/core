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

const MANUAL_TEMPLATE = {
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
};

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;

    const templates = React.useMemo(() => {
        try {
            // Defensive check for SERVICE_REGISTRY
            if (!Array.isArray(SERVICE_REGISTRY)) {
                console.error("SERVICE_REGISTRY is not an array");
                return [MANUAL_TEMPLATE];
            }

            return [
                MANUAL_TEMPLATE,
                ...SERVICE_REGISTRY.map(s => ({
                    id: s.id,
                    name: s.name,
                    description: s.description,
                    config: {
                        commandLineService: {
                            command: s.command,
                            env: {}, // Will be filled from params
                            workingDirectory: ''
                        },
                        openapiService: undefined,
                        configurationSchema: s.configurationSchema ? JSON.stringify(s.configurationSchema) : undefined
                    },
                    params: {} // Will be filled with defaults dynamically
                }))
            ];
        } catch (e) {
            console.error("Failed to load templates from registry", e);
            return [MANUAL_TEMPLATE];
        }
    }, []);

    const handleTemplateChange = (val: string) => {
        const template = templates.find(t => t.id === val);
        if (template) {
            let initialParams = { ...template.params };

            // Extract defaults from schema if available
            if (template.config.configurationSchema) {
                try {
                    const schema = JSON.parse(template.config.configurationSchema);
                    if (schema && typeof schema === 'object' && schema.properties) {
                         Object.entries(schema.properties).forEach(([k, v]: [string, any]) => {
                            if (v && v.default !== undefined) {
                                initialParams[k] = String(v.default);
                            }
                        });
                    }
                } catch (e) {
                    console.error("Failed to parse schema defaults", e);
                }
            }

            updateState({
                selectedTemplateId: val,
                params: initialParams as Record<string, string>
            });

            // Update config name and structure
            // We use 'as any' because UpstreamServiceConfig type might be strict about undefined fields
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
