/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";
import { Textarea } from "@/components/ui/textarea";
import { EnvVarEditor } from "@/components/services/env-var-editor";

interface InstantiateDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    templateConfig?: UpstreamServiceConfig;
    onComplete: () => void;
}

/**
 * InstantiateDialog.
 *
 * @param onComplete - The onComplete.
 */
export function InstantiateDialog({ open, onOpenChange, templateConfig, onComplete }: InstantiateDialogProps) {
    const { toast } = useToast();
    const router = useRouter();
    const [name, setName] = useState("");
    const [authId, setAuthId] = useState("none");
    const [loading, setLoading] = useState(false);
    const [credentials, setCredentials] = useState<any[]>([]);

    // Command & Env state
    const [command, setCommand] = useState("");
    const [envVars, setEnvVars] = useState<Record<string, { plainText?: string; secretId?: string }>>({});

    // Helper to determine if we should show CLI options
    // Community servers (which are created dynamically) usually have commandLineService populated.
    const isCommandLine = !!templateConfig?.commandLineService;

    useEffect(() => {
        if (open && templateConfig) {
            setName(`${templateConfig.name}-copy`);
            setAuthId("none");

            if (templateConfig.commandLineService) {
                // Initialize command
                setCommand(templateConfig.commandLineService.command || "");

                // Initialize env vars
                const newEnv: Record<string, { plainText?: string; secretId?: string }> = {};
                if (templateConfig.commandLineService.env) {
                    Object.entries(templateConfig.commandLineService.env).forEach(([k, v]) => {
                        const sv = v as any;
                        if (sv && typeof sv === 'object' && sv.plainText) {
                            newEnv[k] = { plainText: sv.plainText };
                        } else if (typeof v === 'string') {
                             // Fallback if the input config was just a record of strings
                             newEnv[k] = { plainText: v };
                        } else if (sv && typeof sv === 'object' && sv.environmentVariable) {
                             // If it's a server-side env var reference, we show it as plain text for now or handle appropriately
                             // For now, let's just map plainText
                        }
                    });
                }
                setEnvVars(newEnv);
            } else {
                setCommand("");
                setEnvVars({});
            }

            apiClient.listCredentials().then(setCredentials).catch(console.error);
        }
    }, [open, templateConfig]);

    const handleInstantiate = async () => {
        if (!templateConfig) return;
        setLoading(true);
        try {
            const newConfig = { ...templateConfig };
            newConfig.name = name;
            newConfig.id = name; // ID is name for now

            // Update CLI specific fields if applicable
            if (newConfig.commandLineService) {
                newConfig.commandLineService.command = command;

                // Map EnvVars back to API format
                const apiEnv: Record<string, any> = {};
                Object.entries(envVars).forEach(([k, v]) => {
                    if (v.plainText) {
                        apiEnv[k] = { plainText: v.plainText };
                    } else if (v.secretId) {
                        // Secret handling placeholder
                    }
                });
                newConfig.commandLineService.env = apiEnv;
            }

            if (authId !== 'none') {
                const cred = credentials.find(c => c.id === authId);
                if (cred && cred.authentication) {
                    newConfig.upstreamAuth = cred.authentication;
                }
            }

            await apiClient.registerService(newConfig);
            toast({ title: "Service Instantiated", description: `${name} is now running.` });
            onOpenChange(false);

            // Redirect to the new service page
            router.push(`/upstream-services/${name}`);
        } catch (e) {
            toast({ title: "Failed to instantiate", variant: "destructive", description: String(e) });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Instantiate Service</DialogTitle>
                    <DialogDescription>
                        Create a running instance from this template.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-6 py-4">
                    <div className="grid gap-2">
                        <Label htmlFor="service-name-input">Service Name</Label>
                        <Input id="service-name-input" value={name} onChange={e => setName(e.target.value)} />
                    </div>

                    {isCommandLine && (
                        <>
                            <div className="grid gap-2">
                                <Label htmlFor="command-input">Command</Label>
                                <Textarea
                                    id="command-input"
                                    value={command}
                                    onChange={e => setCommand(e.target.value)}
                                    className="font-mono text-sm"
                                    rows={3}
                                />
                                <p className="text-xs text-muted-foreground">
                                    The command to run the MCP server. Use absolute paths for executables if needed.
                                </p>
                            </div>

                            <div className="grid gap-2">
                                <EnvVarEditor initialEnv={envVars} onChange={setEnvVars} />
                            </div>
                        </>
                    )}

                    <div className="grid gap-2">
                        <Label>Authentication (Optional)</Label>
                        <Select value={authId} onValueChange={setAuthId}>
                            <SelectTrigger>
                                <SelectValue placeholder="Select credential" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="none">None</SelectItem>
                                {credentials.map(c => (
                                    <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleInstantiate} disabled={loading}>
                        {loading ? "Creating..." : "Create Instance"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
