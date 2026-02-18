/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { SchemaForm } from '@/components/marketplace/schema-form';

/**
 * StepParameters component.
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateState, updateConfig } = useWizard();
    const { params, config } = state;

    // Parse schema if available to determine if we should show the form or the raw editor
    const schema = useMemo(() => {
        if (!config.configurationSchema) return null;
        try {
            return JSON.parse(config.configurationSchema);
        } catch (e) {
            console.error("Failed to parse schema", e);
            return null;
        }
    }, [config.configurationSchema]);

    const syncEnv = (currentParams: Record<string, string>) => {
        if (config.commandLineService) {
            const env: any = {};
            Object.entries(currentParams).forEach(([k, v]) => {
                env[k] = { plainText: v };
            });
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        }
    };

    const handleSchemaChange = (newParams: Record<string, string>) => {
        updateState({ params: newParams });
        syncEnv(newParams);
    };

    const handleParamChange = (key: string, value: string, newKey?: string) => {
        const newParams = { ...params };
        if (newKey !== undefined && newKey !== key) {
             // Key change
             delete newParams[key];
             newParams[newKey] = value;
        } else {
            newParams[key] = value;
        }
        updateState({ params: newParams });
        syncEnv(newParams);
    };

    const addParam = () => {
        const newParams = { ...params, "": "" };
        updateState({ params: newParams });
    };

    const removeParam = (key: string) => {
        const newParams = { ...params };
        delete newParams[key];
        updateState({ params: newParams });
        syncEnv(newParams);
    };

    return (
        <div className="space-y-6">
            {schema ? (
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                         <h3 className="text-lg font-medium">Configuration</h3>
                    </div>
                    <p className="text-sm text-muted-foreground">
                        Configure the service using the form below. Required fields are marked with an asterisk.
                    </p>
                    <SchemaForm schema={schema} value={params} onChange={handleSchemaChange} />
                </div>
            ) : (
                <>
                    <div className="flex items-center justify-between">
                         <h3 className="text-lg font-medium">Environment Variables / Parameters</h3>
                         <Button size="sm" onClick={addParam}><Plus className="mr-2 h-4 w-4"/> Add Parameter</Button>
                    </div>

                    <div className="border rounded-lg">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Key</TableHead>
                                    <TableHead>Value</TableHead>
                                    <TableHead className="w-[50px]"></TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {Object.entries(params).map(([key, value], idx) => (
                                    <TableRow key={idx}>
                                        <TableCell>
                                            <Input
                                                value={key}
                                                placeholder="VAR_NAME"
                                                onChange={e => handleParamChange(key, value, e.target.value)}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <Input
                                                value={value}
                                                placeholder="Value"
                                                onChange={e => handleParamChange(key, e.target.value)}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <Button variant="ghost" size="icon" onClick={() => removeParam(key)}>
                                                <Trash2 className="h-4 w-4 text-destructive" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                                {Object.keys(params).length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={3} className="text-center text-muted-foreground h-24">
                                            No parameters configured.
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </>
            )}

             <div className="space-y-4 pt-4 border-t">
                 <h3 className="text-lg font-medium">Command</h3>
                 <div className="grid gap-2">
                     <Label>Executable</Label>

                     <Input
                        value={config.commandLineService?.command || ''}
                        onChange={e => updateConfig({
                            commandLineService: {
                                ...(config.commandLineService || { env: {}, workingDirectory: '', tools: [], resources: [], calls: {}, prompts: [], communicationProtocol: 0, local: false }),
                                command: e.target.value
                            }
                        })}
                        placeholder="npx -y package-name OR /usr/bin/python3"
                     />
                     <p className="text-xs text-muted-foreground">
                        The command to execute. For Node.js, use `npx -y package-name`. For Python, use `uvx package-name`.
                     </p>
                 </div>
             </div>
        </div>
    );
}
