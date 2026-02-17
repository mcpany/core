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

    // Parse schema if available
    const schema = useMemo(() => {
        // Check if configurationSchema is present in the partial config
        const schemaStr = (config as any).configurationSchema;
        if (schemaStr) {
            try {
                return JSON.parse(schemaStr);
            } catch (e) {
                console.error("Failed to parse configuration schema", e);
                return null;
            }
        }
        return null;
    }, [config]);

    const syncToConfig = (newParams: Record<string, string>) => {
        if (config.commandLineService) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
                // If the schema defines this field, we might want to respect its type?
                // But for now, everything is environment variable string.
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
        syncToConfig(newParams);
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
        syncToConfig(newParams);
    };

    const addParam = () => {
        const newParams = { ...params, "": "" };
        updateState({ params: newParams });
    };

    const removeParam = (key: string) => {
        const newParams = { ...params };
        delete newParams[key];
        updateState({ params: newParams });
        syncToConfig(newParams);
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                 <h3 className="text-lg font-medium">
                     {schema ? "Configuration" : "Environment Variables / Parameters"}
                 </h3>
                 {!schema && (
                    <Button size="sm" onClick={addParam}><Plus className="mr-2 h-4 w-4"/> Add Parameter</Button>
                 )}
            </div>

            {schema ? (
                <div className="border rounded-lg p-4 bg-muted/20">
                     <p className="text-sm text-muted-foreground mb-4">
                        Configure the service using the provided schema.
                     </p>
                    <SchemaForm schema={schema} value={params} onChange={handleSchemaChange} />
                </div>
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
                     />
                 </div>
                 <div className="grid gap-2">
                     <p className="text-xs text-muted-foreground">
                        {schema ?
                            "The command is pre-configured by the template but can be overridden." :
                            "Modify the command above to include arguments if needed."
                        }
                     </p>
                 </div>
             </div>
        </div>
    );
}
