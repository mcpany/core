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
import { apiClient, UpstreamServiceConfig } from '@/lib/client';
import { Loader2 } from 'lucide-react';

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
        configurationSchema: ""
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
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const loadTemplates = async () => {
            setLoading(true);
            try {
                const fetched = await apiClient.listTemplates();
                // Map API templates to wizard format
                const mapped = fetched.map((t: UpstreamServiceConfig) => {
                    // Extract default params from schema or env
                    const initialParams: Record<string, string> = {};
                    if (t.commandLineService?.env) {
                        Object.entries(t.commandLineService.env).forEach(([k, v]) => {
                             initialParams[k] = (v as any).plainText || "";
                        });
                    }

                    // If schema exists, defaults might be in schema, but we'll rely on env for now
                    // or let schema form handle defaults later.

                    return {
                        id: t.id || t.name,
                        name: t.name,
                        description: `Template for ${t.name}`, // Description not always in API yet
                        config: t,
                        params: initialParams
                    };
                });
                setTemplates([MANUAL_TEMPLATE, ...mapped]);
            } catch (e) {
                console.error("Failed to load templates", e);
            } finally {
                setLoading(false);
            }
        };
        loadTemplates();
    }, []);


    const handleTemplateChange = (val: string) => {
        const template = templates.find(t => t.id === val);
        if (template) {
            updateState({
                selectedTemplateId: val,
                params: template.params as Record<string, string>
            });
            updateConfig({
                ...template.config,
                name: config.name || template.name,
                configurationSchema: template.config.configurationSchema // Ensure schema is passed
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
                <div className="flex gap-2 items-center">
                    <Select value={selectedTemplateId || 'manual'} onValueChange={handleTemplateChange} disabled={loading}>
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
                    {loading && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                </div>
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
