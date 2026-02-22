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

const TEMPLATES = [
    {
        id: 'mcp-stdio',
        name: 'Generic MCP Server (Stdio)',
        description: 'Run any MCP server via command line (stdio). Recommended for most users.',
        config: {
            mcpService: {
                stdioConnection: {
                    command: 'npx',
                    args: ['-y', '@modelcontextprotocol/server-name'],
                    env: {}
                }
            },
            commandLineService: undefined,
            openapiService: undefined
        },
        params: {}
    },
    {
        id: 'postgres',
        name: 'PostgreSQL Database',
        description: 'Connect to a PostgreSQL database.',
        config: {
            mcpService: {
                stdioConnection: {
                    command: 'npx',
                    args: ['-y', '@modelcontextprotocol/server-postgres', 'postgresql://user:password@localhost:5432/dbname'],
                    env: {}
                }
            },
            commandLineService: undefined,
            openapiService: undefined
        },
        params: {}
    },
    {
        id: 'filesystem',
        name: 'Filesystem',
        description: 'Expose a local directory.',
        config: {
            mcpService: {
                stdioConnection: {
                    command: 'npx',
                    args: ['-y', '@modelcontextprotocol/server-filesystem', '/home/user'],
                    env: {}
                }
            },
            commandLineService: undefined,
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
            mcpService: undefined,
            commandLineService: undefined
        },
        params: {}
    },
    {
        id: 'manual',
        name: 'Legacy Command Line (Raw)',
        description: 'Configure a raw command line service wrapper (Advanced).',
        config: {
            commandLineService: {
                command: '',
                env: {},
                workingDirectory: ''
            },
            mcpService: undefined,
            openapiService: undefined
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


    const handleTemplateChange = (val: string) => {
        const template = TEMPLATES.find(t => t.id === val);
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
                <Select value={selectedTemplateId || 'mcp-stdio'} onValueChange={handleTemplateChange}>
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
                            {TEMPLATES.find(t => t.id === (selectedTemplateId || 'mcp-stdio'))?.name}
                        </CardTitle>
                        <CardDescription>
                            {TEMPLATES.find(t => t.id === (selectedTemplateId || 'mcp-stdio'))?.description}
                        </CardDescription>
                    </CardHeader>
                </Card>
            </div>
        </div>
    );
}
