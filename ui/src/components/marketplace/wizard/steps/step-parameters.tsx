/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus, GripVertical } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

/**
 * StepParameters component.
 * Supports both MCP Stdio Connection (Command + Args + Env) and Legacy Command Line Service (Command + Env).
 * @returns The rendered component.
 */
export function StepParameters() {
    const { state, updateConfig } = useWizard();
    const { config } = state;

    const isMcpStdio = !!config.mcpService?.stdioConnection;
    const isLegacyCommandLine = !!config.commandLineService;

    // --- Helper for MCP Stdio Arguments ---

    const handleArgChange = (index: number, value: string) => {
        if (!config.mcpService?.stdioConnection) return;
        const newArgs = [...(config.mcpService.stdioConnection.args || [])];
        newArgs[index] = value;
        updateConfig({
            mcpService: {
                ...config.mcpService,
                stdioConnection: {
                    ...config.mcpService.stdioConnection,
                    args: newArgs
                }
            }
        });
    };

    const addArg = () => {
        if (!config.mcpService?.stdioConnection) return;
        const newArgs = [...(config.mcpService.stdioConnection.args || []), ""];
        updateConfig({
            mcpService: {
                ...config.mcpService,
                stdioConnection: {
                    ...config.mcpService.stdioConnection,
                    args: newArgs
                }
            }
        });
    };

    const removeArg = (index: number) => {
        if (!config.mcpService?.stdioConnection) return;
        const newArgs = [...(config.mcpService.stdioConnection.args || [])];
        newArgs.splice(index, 1);
        updateConfig({
            mcpService: {
                ...config.mcpService,
                stdioConnection: {
                    ...config.mcpService.stdioConnection,
                    args: newArgs
                }
            }
        });
    };

    // --- Helper for Environment Variables ---

    const getEnvFromConfig = () => {
        if (isMcpStdio) return config.mcpService?.stdioConnection?.env || {};
        if (isLegacyCommandLine) return config.commandLineService?.env || {};
        return {};
    };

    // Local state for env vars to prevent UI jumping when keys change order in map
    const [localEnv, setLocalEnv] = useState<{ id: string, key: string, value: string }[]>(() => {
        const env = getEnvFromConfig();
        return Object.entries(env).map(([k, v]) => ({
            id: Math.random().toString(36).substring(7),
            key: k,
            value: v.plainText || ""
        }));
    });

    // Sync local state to config whenever it changes
    useEffect(() => {
        const newEnvMap: Record<string, any> = {};
        localEnv.forEach(entry => {
            if (entry.key) {
                newEnvMap[entry.key] = { plainText: entry.value };
            }
        });

        if (isMcpStdio && config.mcpService?.stdioConnection) {
            // Avoid infinite loop by checking if meaningful change?
            // Actually useEffect dependency on localEnv is enough.
            // But we need to make sure we don't trigger re-render of this component that resets state?
            // No, localEnv is state.
            updateConfig({
                mcpService: {
                    ...config.mcpService,
                    stdioConnection: {
                        ...config.mcpService.stdioConnection,
                        env: newEnvMap
                    }
                }
            });
        } else if (isLegacyCommandLine && config.commandLineService) {
            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env: newEnvMap
                }
            });
        }
    }, [localEnv]); // Only run when localEnv changes

    const handleLocalEnvChange = (id: string, field: 'key' | 'value', val: string) => {
        setLocalEnv(prev => prev.map(entry => {
            if (entry.id === id) {
                return { ...entry, [field]: val };
            }
            return entry;
        }));
    };

    const addLocalEnv = () => {
        setLocalEnv(prev => [...prev, { id: Math.random().toString(36).substring(7), key: "", value: "" }]);
    };

    const removeLocalEnv = (id: string) => {
        setLocalEnv(prev => prev.filter(e => e.id !== id));
    };


    // --- Render ---

    if (!isMcpStdio && !isLegacyCommandLine) {
        return <div className="text-muted-foreground p-4">No command line configuration available for this service type.</div>;
    }

    return (
        <div className="space-y-8">
            {/* Command Section */}
            <div className="space-y-4">
                <div className="flex flex-col gap-2">
                    <Label className="text-base font-semibold">Executable Command</Label>
                    <p className="text-sm text-muted-foreground">The binary or script to run (e.g., `npx`, `python`, `docker`).</p>
                    <Input
                        value={
                            isMcpStdio ? config.mcpService?.stdioConnection?.command :
                            isLegacyCommandLine ? config.commandLineService?.command : ''
                        }
                        onChange={e => {
                            if (isMcpStdio && config.mcpService?.stdioConnection) {
                                updateConfig({
                                    mcpService: {
                                        ...config.mcpService,
                                        stdioConnection: { ...config.mcpService.stdioConnection, command: e.target.value }
                                    }
                                });
                            } else if (isLegacyCommandLine && config.commandLineService) {
                                updateConfig({
                                    commandLineService: { ...config.commandLineService, command: e.target.value }
                                });
                            }
                        }}
                        placeholder="npx"
                        className="font-mono"
                    />
                </div>
            </div>

            {/* Arguments Section (MCP Stdio Only) */}
            {isMcpStdio && (
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <div>
                            <Label className="text-base font-semibold">Arguments</Label>
                            <p className="text-sm text-muted-foreground">Arguments passed to the command.</p>
                        </div>
                        <Button size="sm" variant="outline" onClick={addArg} className="gap-2">
                            <Plus className="h-4 w-4" /> Add Argument
                        </Button>
                    </div>

                    <div className="space-y-2">
                        {config.mcpService?.stdioConnection?.args?.map((arg: string, idx: number) => (
                            <div key={idx} className="flex gap-2 items-center group">
                                <div className="text-muted-foreground/30 cursor-move">
                                    <GripVertical className="h-4 w-4" />
                                </div>
                                <Input
                                    value={arg}
                                    onChange={e => handleArgChange(idx, e.target.value)}
                                    className="font-mono text-sm"
                                    placeholder={`Arg ${idx + 1}`}
                                />
                                <Button variant="ghost" size="icon" onClick={() => removeArg(idx)} className="opacity-0 group-hover:opacity-100 transition-opacity">
                                    <Trash2 className="h-4 w-4 text-destructive" />
                                </Button>
                            </div>
                        ))}
                        {(!config.mcpService?.stdioConnection?.args || config.mcpService.stdioConnection.args.length === 0) && (
                            <div className="text-sm text-muted-foreground italic border border-dashed rounded-md p-4 text-center">
                                No arguments.
                            </div>
                        )}
                    </div>

                    {/* Preview Command */}
                    <div className="mt-4 p-3 bg-muted rounded-md font-mono text-xs break-all">
                        <span className="text-muted-foreground select-none">$ </span>
                        <span className="text-primary">{config.mcpService?.stdioConnection?.command}</span>
                        {config.mcpService?.stdioConnection?.args?.map((arg: string, i: number) => (
                            <span key={i} className="ml-2">{arg.includes(' ') ? `"${arg}"` : arg}</span>
                        ))}
                    </div>
                </div>
            )}

            {/* Environment Variables Section */}
            <div className="space-y-4">
                <div className="flex items-center justify-between">
                     <div>
                        <Label className="text-base font-semibold">Environment Variables</Label>
                        <p className="text-sm text-muted-foreground">Key-value pairs set in the process environment.</p>
                     </div>
                     <Button size="sm" variant="outline" onClick={addLocalEnv} className="gap-2">
                        <Plus className="h-4 w-4"/> Add Variable
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
                            {localEnv.map((entry) => (
                                <TableRow key={entry.id}>
                                    <TableCell>
                                        <Input
                                            value={entry.key}
                                            placeholder="KEY"
                                            onChange={e => handleLocalEnvChange(entry.id, 'key', e.target.value)}
                                            className="font-mono text-xs h-8"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <Input
                                            value={entry.value}
                                            placeholder="Value"
                                            onChange={e => handleLocalEnvChange(entry.id, 'value', e.target.value)}
                                            className="font-mono text-xs h-8"
                                            type={entry.key.toUpperCase().includes("KEY") || entry.key.toUpperCase().includes("SECRET") ? "password" : "text"}
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <Button variant="ghost" size="icon" onClick={() => removeLocalEnv(entry.id)} className="h-8 w-8">
                                            <Trash2 className="h-4 w-4 text-destructive" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                            {localEnv.length === 0 && (
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
