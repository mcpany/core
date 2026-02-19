/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { SchemaForm } from '@/components/marketplace/schema-form';
import { EnvVarEditor } from '@/components/services/env-var-editor';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';

/**
 * StepParameters component.
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateConfig, updateState } = useWizard();
    const { config, params } = state;

    // Determine if we have a schema to render
    // The schema string is stored in config.configurationSchema (if from template)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let schema: any = null;
    try {
        if (config.configurationSchema) {
            schema = JSON.parse(config.configurationSchema);
        }
    } catch (_e) {
        // ignore
    }

    const handleSchemaChange = (newParams: Record<string, string>) => {
        updateState({ params: newParams });

        // Also map back to config env vars if commandLineService exists
        // This mapping depends on the template logic. Usually schema fields map to env vars.
        // The SERVICE_REGISTRY uses uppercase keys for env vars directly as properties.
        // So we can assume keys match.
        if (config.commandLineService) {
            const newEnv = { ...(config.commandLineService.env || {}) };
            Object.entries(newParams).forEach(([k, v]) => {
                newEnv[k] = v; // Simple string assignment
            });
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    env: newEnv as any
                }
            });
        }
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const handleEnvChange = (env: Record<string, any>) => {
        if (config.commandLineService) {
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        }
    };

    return (
        <div className="space-y-6">
            <h3 className="text-lg font-medium">Service Configuration</h3>

            {schema ? (
                <SchemaForm
                    schema={schema}
                    value={params || {}}
                    onChange={handleSchemaChange}
                />
            ) : config.commandLineService ? (
                <div className="space-y-4">
                     <div className="grid gap-2">
                        <Label htmlFor="command">Command</Label>
                        <Input
                            id="command"
                            value={config.commandLineService.command || ''}
                            onChange={e => updateConfig({
                                commandLineService: {
                                    ...config.commandLineService!,
                                    command: e.target.value
                                }
                            })}
                        />
                     </div>
                     <EnvVarEditor
                        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        initialEnv={config.commandLineService.env as any}
                        onChange={handleEnvChange}
                     />
                </div>
            ) : (
                <div className="text-sm text-muted-foreground">
                    No specific parameters required or available for this template type.
                </div>
            )}
        </div>
    );
}
