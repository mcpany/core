/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard, ParamValue } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

/**
 * StepParameters component.
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateState, updateConfig } = useWizard();
    const { params, config } = state;

    const syncConfig = (newParams: Record<string, ParamValue>) => {
        if (config.commandLineService) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
                if (v.type === 'environmentVariable') {
                    env[k] = { environmentVariable: v.value };
                } else {
                    env[k] = { plainText: v.value };
                }
            });
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        }
    };

    const handleParamChange = (key: string, updates: Partial<ParamValue>, newKey?: string) => {
        const newParams = { ...params };
        const currentParam = params[key];

        if (newKey !== undefined && newKey !== key) {
             // Key change
             delete newParams[key];
             newParams[newKey] = { ...currentParam, ...updates };
        } else {
            newParams[key] = { ...currentParam, ...updates };
        }
        updateState({ params: newParams });
        syncConfig(newParams);
    };

    const addParam = () => {
        const newParams = { ...params, "": { type: 'plainText', value: '' } as ParamValue };
        updateState({ params: newParams });
        // Don't sync config yet as key is empty
    };

    const removeParam = (key: string) => {
        const newParams = { ...params };
        delete newParams[key];
        updateState({ params: newParams });
        syncConfig(newParams);
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
                            <TableHead className="w-[30%]">Key</TableHead>
                            <TableHead className="w-[20%]">Type</TableHead>
                            <TableHead>Value</TableHead>
                            <TableHead className="w-[50px]"></TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {Object.entries(params).map(([key, param], idx) => (
                            <TableRow key={idx}>
                                <TableCell>
                                    <Input
                                        value={key}
                                        placeholder="VAR_NAME"
                                        onChange={e => handleParamChange(key, {}, e.target.value)}
                                    />
                                </TableCell>
                                <TableCell>
                                    <Select
                                        value={param.type}
                                        onValueChange={(val: 'plainText' | 'environmentVariable') => handleParamChange(key, { type: val })}
                                    >
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="plainText">Plain Text</SelectItem>
                                            <SelectItem value="environmentVariable">Host Env Var</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </TableCell>
                                <TableCell>
                                    <Input
                                        value={param.value}
                                        placeholder={param.type === 'environmentVariable' ? 'HOST_VAR_NAME' : 'Value'}
                                        onChange={e => handleParamChange(key, { value: e.target.value })}
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
                                <TableCell colSpan={4} className="text-center text-muted-foreground h-24">
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
