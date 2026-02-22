/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus, Eye, EyeOff, GripVertical } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent } from '@/components/ui/card';

/**
 * StepParameters component.
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateState, updateConfig } = useWizard();
    const { params, config } = state;

    // Detect mode
    const isMcpStdio = !!config.mcpService?.connectionType?.stdioConnection;
    const isCommandLine = !!config.commandLineService;

    // If neither is set (e.g. Manual start), default to CommandLineService (Legacy/Simple)
    // unless we want to promote McpService.
    // Let's check if we have a template selected.

    // Helper to update Env Vars
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

        // Sync to Config
        const env: any = {};
        Object.entries(newParams).forEach(([k, v]) => {
            env[k] = { plainText: v };
        });

        if (isMcpStdio) {
             updateConfig({
                mcpService: {
                    ...config.mcpService,
                    connectionType: {
                        stdioConnection: {
                            ...(config.mcpService?.connectionType?.stdioConnection || { command: '', args: [] }),
                            env
                        }
                    }
                }
            });
        } else {
            updateConfig({
                commandLineService: {
                    ...(config.commandLineService || { command: '', workingDirectory: '', tools: [], resources: [], calls: {}, prompts: [], communicationProtocol: 0, local: false }),
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

        // Sync removal
        const env: any = {};
        Object.entries(newParams).forEach(([k, v]) => {
            env[k] = { plainText: v };
        });

        if (isMcpStdio) {
             updateConfig({
                mcpService: {
                    ...config.mcpService,
                    connectionType: {
                        stdioConnection: {
                            ...(config.mcpService?.connectionType?.stdioConnection || { command: '', args: [] }),
                            env
                        }
                    }
                }
            });
        } else {
             updateConfig({
                commandLineService: {
                    ...(config.commandLineService || { command: '', workingDirectory: '', tools: [], resources: [], calls: {}, prompts: [], communicationProtocol: 0, local: false }),
                    env
                }
            });
        }
    };

    // --- MCP Stdio Specific Handlers ---

    const handleArgChange = (index: number, value: string) => {
        const currentArgs = config.mcpService?.connectionType?.stdioConnection?.args || [];
        const newArgs = [...currentArgs];
        newArgs[index] = value;
        updateConfig({
            mcpService: {
                ...config.mcpService,
                connectionType: {
                    stdioConnection: {
                        ...(config.mcpService?.connectionType?.stdioConnection || { command: '', env: {} }),
                        args: newArgs
                    }
                }
            }
        });
    };

    const addArg = () => {
        const currentArgs = config.mcpService?.connectionType?.stdioConnection?.args || [];
        updateConfig({
            mcpService: {
                ...config.mcpService,
                connectionType: {
                    stdioConnection: {
                        ...(config.mcpService?.connectionType?.stdioConnection || { command: '', env: {} }),
                        args: [...currentArgs, ""]
                    }
                }
            }
        });
    };

    const removeArg = (index: number) => {
        const currentArgs = config.mcpService?.connectionType?.stdioConnection?.args || [];
        const newArgs = [...currentArgs];
        newArgs.splice(index, 1);
        updateConfig({
            mcpService: {
                ...config.mcpService,
                connectionType: {
                    stdioConnection: {
                        ...(config.mcpService?.connectionType?.stdioConnection || { command: '', env: {} }),
                        args: newArgs
                    }
                }
            }
        });
    };

    return (
        <div className="space-y-8">
            {/* Command Section */}
            <div className="space-y-4">
                 <h3 className="text-lg font-medium">Command Configuration</h3>

                 {isMcpStdio ? (
                     // MCP Stdio Editor
                     <div className="grid gap-4">
                         <div className="grid gap-2">
                             <Label>Executable</Label>
                             <Input
                                value={config.mcpService?.connectionType?.stdioConnection?.command || ''}
                                onChange={e => updateConfig({
                                    mcpService: {
                                        ...config.mcpService,
                                        connectionType: {
                                            stdioConnection: {
                                                ...(config.mcpService?.connectionType?.stdioConnection || { args: [], env: {} }),
                                                command: e.target.value
                                            }
                                        }
                                    }
                                })}
                                placeholder="npx, python3, docker"
                             />
                         </div>

                         <div className="space-y-2">
                             <div className="flex items-center justify-between">
                                 <Label>Arguments</Label>
                                 <Button size="sm" variant="outline" onClick={addArg} className="h-7 text-xs">
                                     <Plus className="mr-1 h-3 w-3"/> Add Argument
                                 </Button>
                             </div>
                             <div className="space-y-2">
                                 {(config.mcpService?.connectionType?.stdioConnection?.args || []).map((arg, idx) => (
                                     <div key={idx} className="flex gap-2 items-center group">
                                         <div className="text-muted-foreground/30 cursor-grab">
                                             <GripVertical className="h-4 w-4" />
                                         </div>
                                         <Input
                                             value={arg}
                                             onChange={e => handleArgChange(idx, e.target.value)}
                                             placeholder={`Argument ${idx + 1}`}
                                             className="font-mono text-sm"
                                         />
                                         <Button variant="ghost" size="icon" onClick={() => removeArg(idx)} className="h-8 w-8 text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity">
                                             <Trash2 className="h-4 w-4" />
                                         </Button>
                                     </div>
                                 ))}
                                 {(config.mcpService?.connectionType?.stdioConnection?.args || []).length === 0 && (
                                     <div className="text-sm text-muted-foreground italic border border-dashed rounded p-4 text-center bg-muted/20">
                                         No arguments defined.
                                     </div>
                                 )}
                             </div>
                         </div>
                     </div>
                 ) : (
                     // Legacy Command Line Editor
                     <div className="grid gap-2">
                         <Label>Full Command</Label>
                         <Input
                            value={config.commandLineService?.command || ''}
                            onChange={e => updateConfig({
                                commandLineService: {
                                    ...(config.commandLineService || { env: {}, workingDirectory: '', tools: [], resources: [], calls: {}, prompts: [], communicationProtocol: 0, local: false }),
                                    command: e.target.value
                                }
                            })}
                            placeholder="npx -y package-name OR /usr/bin/python3 script.py"
                         />
                         <p className="text-xs text-muted-foreground">
                             Enter the full command string including arguments.
                         </p>
                     </div>
                 )}
            </div>

            {/* Env Vars Section (Shared) */}
            <div className="space-y-4">
                <div className="flex items-center justify-between">
                     <h3 className="text-lg font-medium">Environment Variables</h3>
                     <Button size="sm" variant="outline" onClick={addParam}>
                         <Plus className="mr-2 h-4 w-4"/> Add Variable
                     </Button>
                </div>

                <div className="border rounded-lg overflow-hidden">
                    <Table>
                        <TableHeader>
                            <TableRow className="bg-muted/50">
                                <TableHead className="w-[40%]">Key</TableHead>
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
                                            className="font-mono text-xs uppercase"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <div className="relative">
                                            <Input
                                                value={value}
                                                placeholder="Value"
                                                onChange={e => handleParamChange(key, e.target.value)}
                                                type={key.match(/key|secret|token|password/i) ? "password" : "text"}
                                            />
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
