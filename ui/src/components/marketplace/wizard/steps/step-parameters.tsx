/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus, GripVertical, Eye, EyeOff } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

/**
 * StepParameters component.
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateState, updateConfig } = useWizard();
    const { params, config } = state;
    const [maskedParams, setMaskedParams] = useState<Record<string, boolean>>({});

    // Check which mode we are in
    // Note: The UI state (params) is used for editing Env Vars, and we sync to config on change.
    // Ideally we should sync from config on mount too?
    // But `StepServiceType` sets `params` in state when template is selected.

    const isStdio = config.mcpService?.connectionType?.stdioConnection;
    const isLegacy = config.commandLineService;

    const handleParamChange = (key: string, value: string, newKey?: string) => {
        const newParams = { ...params };
        if (newKey !== undefined && newKey !== key) {
             // Key change
             delete newParams[key];
             newParams[newKey] = value;
             // Preserve mask state
             if (maskedParams[key]) {
                 setMaskedParams(prev => {
                     const next = { ...prev };
                     delete next[key];
                     next[newKey] = true;
                     return next;
                 });
             }
        } else {
            newParams[key] = value;
        }
        updateState({ params: newParams });

        // Sync to config
        syncEnvToConfig(newParams);
    };

    const syncEnvToConfig = (newParams: Record<string, string>) => {
        if (isLegacy) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
                env[k] = { plainText: v };
            });
            updateConfig({
                commandLineService: {
                    ...config.commandLineService!,
                    env
                }
            });
        } else if (isStdio) {
            const env: any = {};
            Object.entries(newParams).forEach(([k, v]) => {
                env[k] = { plainText: v };
            });
            // We must preserve existing structure
            const currentStdio = config.mcpService!.connectionType!.stdioConnection!;
            updateConfig({
                mcpService: {
                    ...config.mcpService,
                    connectionType: {
                        ...config.mcpService!.connectionType,
                        stdioConnection: {
                            ...currentStdio,
                            env
                        }
                    }
                }
            });
        }
    };

    const addParam = () => {
        const newParams = { ...params, "": "" };
        updateState({ params: newParams });
        syncEnvToConfig(newParams);
    };

    const removeParam = (key: string) => {
        const newParams = { ...params };
        delete newParams[key];
        updateState({ params: newParams });
        syncEnvToConfig(newParams);
    };

    const toggleMask = (key: string) => {
        setMaskedParams(prev => ({ ...prev, [key]: !prev[key] }));
    };

    // Stdio Argument Management
    const args = isStdio?.args || [];

    const updateArg = (index: number, value: string) => {
        const newArgs = [...args];
        newArgs[index] = value;
        updateStdioArgs(newArgs);
    };

    const addArg = () => {
        updateStdioArgs([...args, ""]);
    };

    const removeArg = (index: number) => {
        const newArgs = [...args];
        newArgs.splice(index, 1);
        updateStdioArgs(newArgs);
    };

    const updateStdioArgs = (newArgs: string[]) => {
        if (!isStdio) return;
        updateConfig({
            mcpService: {
                ...config.mcpService,
                connectionType: {
                    ...config.mcpService!.connectionType,
                    stdioConnection: {
                        ...config.mcpService!.connectionType!.stdioConnection!,
                        args: newArgs
                    }
                }
            }
        });
    };

    const updateCommand = (val: string) => {
        if (isStdio) {
             updateConfig({
                mcpService: {
                    ...config.mcpService,
                    connectionType: {
                        ...config.mcpService!.connectionType,
                        stdioConnection: {
                            ...config.mcpService!.connectionType!.stdioConnection!,
                            command: val
                        }
                    }
                }
            });
        } else if (isLegacy) {
             updateConfig({
                commandLineService: {
                    ...config.commandLineService!,
                    command: val
                }
            });
        }
    };

    if (!isStdio && !isLegacy) {
        return <div className="text-muted-foreground p-4">No CLI configuration available for this service type.</div>;
    }

    return (
        <div className="space-y-8">
             {/* Command Configuration */}
             <div className="space-y-4">
                 <h3 className="text-lg font-medium">Command Configuration</h3>
                 <div className="grid gap-4">
                     <div className="space-y-2">
                         <Label>Executable</Label>
                         <Input
                            value={isStdio ? isStdio.command : isLegacy!.command}
                            onChange={e => updateCommand(e.target.value)}
                            placeholder={isStdio ? "npx" : "npx -y @modelcontextprotocol/server-postgres"}
                         />
                         <p className="text-xs text-muted-foreground">
                             {isStdio ? "The binary or command to execute (e.g. npx, python, docker)." : "The full command string to execute."}
                         </p>
                     </div>

                     {/* Argument Builder (Only for Stdio) */}
                     {isStdio && (
                         <div className="space-y-2">
                             <div className="flex items-center justify-between">
                                 <Label>Arguments</Label>
                                 <Button size="sm" variant="outline" onClick={addArg} className="h-7 text-xs">
                                     <Plus className="mr-2 h-3 w-3"/> Add Argument
                                 </Button>
                             </div>
                             <Card className="bg-muted/5">
                                 <CardContent className="p-2 space-y-2">
                                     {args.length === 0 && (
                                         <div className="text-xs text-muted-foreground text-center py-4 italic">
                                             No arguments defined.
                                         </div>
                                     )}
                                     {args.map((arg, idx) => (
                                         <div key={idx} className="flex gap-2 items-center group">
                                             <div className="text-muted-foreground/30 cursor-move">
                                                 <GripVertical className="h-4 w-4" />
                                             </div>
                                             <Input
                                                 value={arg}
                                                 onChange={e => updateArg(idx, e.target.value)}
                                                 className="h-8 text-sm"
                                                 placeholder={`Arg ${idx + 1}`}
                                             />
                                             <Button variant="ghost" size="icon" onClick={() => removeArg(idx)} className="h-8 w-8 text-muted-foreground hover:text-destructive">
                                                 <Trash2 className="h-4 w-4" />
                                             </Button>
                                         </div>
                                     ))}
                                 </CardContent>
                             </Card>
                             {/* Preview */}
                             <div className="mt-2 p-2 bg-black/90 text-white font-mono text-xs rounded overflow-x-auto">
                                 <span className="text-green-400">$</span> {isStdio.command} {args.map(a => /\s/.test(a) ? `"${a}"` : a).join(" ")}
                             </div>
                         </div>
                     )}
                 </div>
             </div>

            {/* Environment Variables */}
            <div className="space-y-4 pt-4 border-t">
                <div className="flex items-center justify-between">
                     <div className="space-y-1">
                        <h3 className="text-lg font-medium">Environment Variables</h3>
                        <p className="text-sm text-muted-foreground">Configure environment variables and secrets.</p>
                     </div>
                     <Button size="sm" onClick={addParam}><Plus className="mr-2 h-4 w-4"/> Add Variable</Button>
                </div>

                <div className="border rounded-lg overflow-hidden">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[40%]">Key</TableHead>
                                <TableHead>Value</TableHead>
                                <TableHead className="w-[80px]"></TableHead>
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
                                            className="font-mono text-xs"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <div className="relative">
                                            <Input
                                                value={value}
                                                type={maskedParams[key] ? "password" : "text"}
                                                placeholder="Value"
                                                onChange={e => handleParamChange(key, e.target.value)}
                                                className="pr-8 text-sm"
                                            />
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="absolute right-0 top-0 h-full px-2 hover:bg-transparent"
                                                onClick={() => toggleMask(key)}
                                            >
                                                {maskedParams[key] ? <Eye className="h-3 w-3 opacity-50" /> : <EyeOff className="h-3 w-3 opacity-50" />}
                                            </Button>
                                        </div>
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
                                        No environment variables configured.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </div>
    );
}
