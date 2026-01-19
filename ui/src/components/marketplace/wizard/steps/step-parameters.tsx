/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

export function StepParameters() {
    const { state, updateState, updateConfig } = useWizard();
    const { params, config } = state;

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
        // TODO: Sync `params` to `config.commandLineService.env` more robustly
        // For now we just update basic env
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
                     <Label>Arguments (Space separated or JSON array coming soon)</Label>
                     {/* For now just command string editing is easiest if we don't strictly separate args */}
                     <p className="text-xs text-muted-foreground">Modify the command above to include arguments.</p>
                 </div>
             </div>
        </div>
    );
}
