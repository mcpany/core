/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Trash2, Plus, Info } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { SchemaForm } from '@/components/marketplace/schema-form';
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { SecretValue } from "@proto/config/v1/auth";

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
            console.error("Failed to parse config schema", e);
            return null;
        }
    }, [config.configurationSchema]);

    const handleSchemaChange = (newParams: Record<string, string>) => {
        updateState({ params: newParams });

        // Sync to config env
        if (config.commandLineService) {
            // Merge with existing env to preserve hidden variables
            const env: Record<string, SecretValue> = { ...(config.commandLineService.env || {}) };

            Object.entries(newParams).forEach(([k, v]) => {
                env[k] = { plainText: v, validationRegex: "" };
            });

            updateConfig({
                commandLineService: {
                    ...config.commandLineService,
                    env
                }
            });
        }
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

        // Also update config env
        if (config.commandLineService) {
            const env: Record<string, SecretValue> = { ...(config.commandLineService.env || {}) };

            // Note: For manual params, we might want to clear old env keys if they were renamed/removed.
            // But since we are mapping from `params` -> `env`, maybe we should just rebuild `env` from `params`?
            // If we are in "Manual" mode, `params` IS the source of truth.
            // So destructive update is actually correct for manual mode, BUT "Smart" mode might have hidden keys.
            // Since "Smart" mode doesn't allow adding/removing params via the table (it uses SchemaForm),
            // we should separate the logic.

            if (!parsedSchema) {
                // Manual Mode: Rebuild from params to ensure deletions are reflected
                const newEnv: Record<string, SecretValue> = {};
                Object.entries(newParams).forEach(([k, v]) => {
                    newEnv[k] = { plainText: v, validationRegex: "" };
                });
                // If we want to preserve hidden keys, we can't easily distinguish them from deleted keys here.
                // However, Manual mode implies full control.
                // So let's stick to overwriting for manual mode to allow deletions.
                updateConfig({
                    commandLineService: {
                        ...config.commandLineService,
                        env: newEnv
                    }
                });
            } else {
                // Smart Mode (Schema): Only update keys present in params (which come from schema)
                Object.entries(newParams).forEach(([k, v]) => {
                    env[k] = { plainText: v, validationRegex: "" };
                });
                updateConfig({
                    commandLineService: {
                        ...config.commandLineService,
                        env
                    }
                });
            }
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
            if (!parsedSchema) {
                const newEnv: Record<string, SecretValue> = {};
                Object.entries(newParams).forEach(([k, v]) => {
                    newEnv[k] = { plainText: v, validationRegex: "" };
                });
                updateConfig({
                    commandLineService: {
                        ...config.commandLineService,
                        env: newEnv
                    }
                });
            }
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                 <h3 className="text-lg font-medium">Environment Variables / Parameters</h3>
                 {!parsedSchema && <Button size="sm" onClick={addParam}><Plus className="mr-2 h-4 w-4"/> Add Parameter</Button>}
            </div>

            {parsedSchema ? (
                <div className="space-y-4">
                    <Alert>
                        <Info className="h-4 w-4" />
                        <AlertTitle>Smart Configuration</AlertTitle>
                        <AlertDescription>
                            This service provides a configuration schema. Please fill in the required fields below.
                        </AlertDescription>
                    </Alert>
                    <SchemaForm schema={parsedSchema} value={params} onChange={handleSchemaChange} />
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
                     <Label>Arguments (Space separated or JSON array coming soon)</Label>
                     {/* For now just command string editing is easiest if we don't strictly separate args */}
                     <p className="text-xs text-muted-foreground">Modify the command above to include arguments.</p>
                 </div>
             </div>
        </div>
    );
}
