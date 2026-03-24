/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react';
import { useWizard } from '../wizard-context';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { apiClient, ServiceTemplate } from '@/lib/client';
import { Loader2 } from 'lucide-react';

const MANUAL_TEMPLATE: any = {
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

/**
 * StepServiceType component.
 * @returns The rendered component.
 */
export function StepServiceType() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, selectedTemplateId } = state;
    const [templates, setTemplates] = useState<any[]>([MANUAL_TEMPLATE]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        async function fetchTemplates() {
            try {
                setLoading(true);
                const data = await apiClient.getServiceTemplates();

                const mapped = data.map(t => {
                    // Extract params from config
                    const params: Record<string, string> = {};
                    const sc = t.serviceConfig;

                    if (sc.commandLineService && sc.commandLineService.env) {
                        for (const [key, val] of Object.entries(sc.commandLineService.env)) {
                            // Handle if val is string or object (EnvVarValue)
                            if (typeof val === 'string') {
                                params[key] = val;
                            } else if (val && typeof val === 'object' && (val as any).plainText) {
                                params[key] = (val as any).plainText;
                            } else {
                                // Fallback
                                params[key] = "";
                            }
                        }
                    }
                    // Handle HTTP service env? Usually not present, but good to check.

                    return {
                        ...t,
                        config: t.serviceConfig,
                        params: params
                    };
                });

                setTemplates([MANUAL_TEMPLATE, ...mapped]);
            } catch (e) {
                console.error("Failed to load templates", e);
            } finally {
                setLoading(false);
            }
        }
        fetchTemplates();
    }, []);

    const handleTemplateChange = (val: string) => {
        const template = templates.find(t => t.id === val);
        if (template) {
            updateState({
                selectedTemplateId: val,
                params: template.params || {}
            });
            // Don't overwrite name if user already typed one, unless it was empty or default
            const newName = config.name && config.name !== MANUAL_TEMPLATE.name ? config.name : template.name;
            updateConfig({
                ...template.config,
                name: newName,
            });
        }
    };

    if (loading) {
        return (
            <div className="flex h-40 items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

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
