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
import { SchemaForm } from '../../schema-form';

/**
 * StepParameters component.
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateState, updateConfig } = useWizard();
    const { params, config } = state;

    const parsedSchema = useMemo(() => {
        if (!config.configurationSchema) return null;
        try {
            return JSON.parse(config.configurationSchema);
        } catch (e) {
            console.error("Failed to parse configuration schema", e);
            return null;
        }
    }, [config.configurationSchema]);

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

        // Also update config env
        if (config.commandLineService) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
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

    const handleSchemaChange = (newValues: Record<string, string>) => {
        updateState({ params: newValues });
        if (config.commandLineService) {
            const env: any = {};
            Object.entries(newValues).forEach(([k, v]) => {
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

    const addParam = () => {
        const newParams = { ...params, "": "" };
        updateState({ params: newParams });
    };

    const removeParam = (key: string) => {
        const newParams = { ...params };
        delete newParams[key];
        updateState({ params: newParams });
         // Sync with config
         if (config.commandLineService) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
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

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                 <h3 className="text-lg font-medium">Configuration Parameters</h3>
                 {!parsedSchema && (
                    <Button size="sm" onClick={addParam}><Plus className="mr-2 h-4 w-4"/> Add Parameter</Button>
                 )}
            </div>

            {parsedSchema ? (
                <SchemaForm schema={parsedSchema} value={params} onChange={handleSchemaChange} />
            ) : (
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
                        disabled={!!parsedSchema} // Lock command if using schema/template? Maybe safer to allow edit but warn
                     />
                     {parsedSchema && <p className="text-xs text-muted-foreground">Command provided by template.</p>}

                 </div>
                 <div className="grid gap-2">
                     <Label>Arguments (Space separated)</Label>
                     {/* For now just command string editing is easiest if we don't strictly separate args */}
                     <p className="text-xs text-muted-foreground">Modify the command above to include arguments.</p>
                 </div>
             </div>
        </div>
    );
}
