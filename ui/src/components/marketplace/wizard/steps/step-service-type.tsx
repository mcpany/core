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
        openapiService: undefined
    },
    params: {}
};

const OPENAPI_TEMPLATE = {
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

// Map registry items to wizard templates
const REGISTRY_TEMPLATES = SERVICE_REGISTRY.map(s => {
    // Extract defaults from schema if any
    const defaultParams: Record<string, string> = {};
    if (s.configurationSchema && s.configurationSchema.properties) {
        Object.entries(s.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
            if (prop.default !== undefined) {
                defaultParams[key] = String(prop.default);
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
                env: {}, // Will be populated from params later
                workingDirectory: ''
            },
            configurationSchema: JSON.stringify(s.configurationSchema),
            openapiService: undefined
        },
        params: defaultParams
    };
});

const TEMPLATES = [MANUAL_TEMPLATE, ...REGISTRY_TEMPLATES, OPENAPI_TEMPLATE];

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
            updateState({
                selectedTemplateId: val,
                params: template.params as Record<string, string>
            });

            // We need to be careful not to overwrite the commandLineService env if we switch templates
            // But actually, switching templates SHOULD reset params, so this is fine.

            // However, we must ensure `configurationSchema` is passed along.
            // The spread `...template.config` handles that because we mapped it above.

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
