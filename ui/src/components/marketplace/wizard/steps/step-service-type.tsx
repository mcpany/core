/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useEffect, useState } from 'react';
import { useWizard } from '../wizard-context';
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { apiClient, ServiceTemplate } from '@/lib/client';
import { SERVICE_REGISTRY } from '@/lib/service-registry';
import { Loader2 } from 'lucide-react';

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
    const [remoteTemplates, setRemoteTemplates] = useState<ServiceTemplate[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                let templates = await apiClient.getServiceTemplates();

                // SEEDING LOGIC: If no templates, seed from Registry
                if (templates.length === 0) {
                    console.log("No templates found. Seeding database from Registry...");
                    // We seed a few core ones to avoid spamming
                    const coreIds = ['postgres', 'sqlite', 'filesystem'];
                    const seedPromises = SERVICE_REGISTRY
                        .filter(r => coreIds.includes(r.id))
                        .map(item => {
                            const config = {
                                id: item.id,
                                name: item.name,
                                description: item.description,
                                version: "1.0.0",
                                disable: false,
                                commandLineService: {
                                    command: item.command,
                                    env: {},
                                    workingDirectory: ''
                                },
                                configurationSchema: JSON.stringify(item.configurationSchema)
                            } as any;
                            return apiClient.saveTemplate(config).catch(e => console.warn(`Failed to seed ${item.id}`, e));
                        });

                    await Promise.all(seedPromises);

                    // Also ensure default profile exists to prevent backend warnings/errors
                    try {
                        await apiClient.createProfile({ id: 'default', name: 'default', description: 'Default Profile' });
                        console.log("Seeded default profile.");
                    } catch (e) {
                        // Profile likely exists or API not ready, ignore
                    }

                    // Fetch again
                    templates = await apiClient.getServiceTemplates();
                }

                setRemoteTemplates(templates);
            } catch (e) {
                console.error("Failed to fetch templates", e);
            } finally {
                setLoading(false);
            }
        };
        fetchTemplates();
    }, []);

    // Merge Static templates with Remote templates
    const allTemplates = useMemo(() => {
        const mappedRemote = remoteTemplates.map(t => {
            // Extract defaults from schema if available
            const defaults: Record<string, string> = {};
            try {
                if (t.serviceConfig.configurationSchema) {
                    const schema = JSON.parse(t.serviceConfig.configurationSchema);
                    if (schema.properties) {
                        Object.entries(schema.properties).forEach(([key, prop]: [string, any]) => {
                            if (prop.default !== undefined) {
                                defaults[key] = String(prop.default);
                            } else {
                                defaults[key] = "";
                            }
                        });
                    }
                }
            } catch (e) {
                // ignore invalid schema
            }

            return {
                id: t.id,
                name: t.name,
                description: t.description,
                config: t.serviceConfig,
                params: defaults
            };
        });

        return [...STATIC_TEMPLATES, ...mappedRemote];
    }, [remoteTemplates]);

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
                <div className="relative">
                    <Select value={selectedTemplateId || 'manual'} onValueChange={handleTemplateChange} disabled={loading}>
                        <SelectTrigger id="service-template">
                            <SelectValue placeholder={loading ? "Loading templates..." : "Select a template"} />
                        </SelectTrigger>
                        <SelectContent className="max-h-[300px]">
                            {allTemplates.map(t => (
                                <SelectItem key={t.id} value={t.id}>
                                    {t.name}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    {loading && <Loader2 className="h-4 w-4 animate-spin absolute right-10 top-3 text-muted-foreground" />}
                </div>

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
