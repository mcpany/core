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
import { SchemaForm } from "./schema-form";
import { SecretValue } from "@proto/config/v1/auth";
import { COMMUNITY_MANIFESTS } from "@/lib/community-manifests";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Sparkles } from "lucide-react";

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
    const [envVars, setEnvVars] = useState<Record<string, SecretValue>>({});

    // Schema state
    const [parsedSchema, setParsedSchema] = useState<any>(null);
    const [schemaValues, setSchemaValues] = useState<Record<string, string>>({});
    const [isSchemaValid, setIsSchemaValid] = useState(true);

    // Manifest State
    const [manifestApplied, setManifestApplied] = useState<string | null>(null);
    const [manifestDescription, setManifestDescription] = useState<string | null>(null);
    const [suggestedKeys, setSuggestedKeys] = useState<Record<string, string>>({});

    // Helper to determine if we should show CLI options
    // Community servers (which are created dynamically) usually have commandLineService populated.
    const isCommandLine = !!templateConfig?.commandLineService;

    useEffect(() => {
        if (open && templateConfig) {
            setName(`${templateConfig.name}-copy`);
            setAuthId("none");
            setManifestApplied(null);
            setManifestDescription(null);
            setSuggestedKeys({});

            let initialCommand = "";
            const newEnv: Record<string, SecretValue> = {};
            const initialSchemaValues: Record<string, string> = {};

            if (templateConfig.commandLineService) {
                // Initialize command
                initialCommand = templateConfig.commandLineService.command || "";

                if (templateConfig.commandLineService.env) {
                    Object.entries(templateConfig.commandLineService.env).forEach(([k, v]) => {
                        const sv = v as any;
                        let val = "";
                        if (sv && typeof sv === 'object' && sv.plainText) {
                            val = sv.plainText;
                            newEnv[k] = { plainText: val, validationRegex: "" };
                        } else if (typeof v === 'string') {
                             // Fallback if the input config was just a record of strings
                             val = v;
                             newEnv[k] = { plainText: val, validationRegex: "" };
                        } else {
                             // Default empty
                             newEnv[k] = { plainText: "", validationRegex: "" };
                        }
                        initialSchemaValues[k] = val;
                    });
                }
            }

            // Check for Community Manifest match
            const manifest = COMMUNITY_MANIFESTS[templateConfig.name] ||
                             COMMUNITY_MANIFESTS[templateConfig.sanitizedName] ||
                             Object.entries(COMMUNITY_MANIFESTS).find(([k]) => templateConfig.name.includes(k))?.[1];

            if (manifest && !templateConfig.configurationSchema) {
                // Apply manifest if no schema exists
                setManifestApplied(templateConfig.name);
                if (manifest.description) setManifestDescription(manifest.description);

                // Merge command if current is generic "npx -y package" or similar, or just prefer manifest
                // We trust the manifest more for command structure
                if (manifest.command) {
                    initialCommand = manifest.command;
                }

                // Prepare suggested keys
                if (manifest.env) {
                    setSuggestedKeys(manifest.env);
                }
            }

            setCommand(initialCommand);
            setEnvVars(newEnv);
            setSchemaValues(initialSchemaValues);

            // Parse Schema if available (Overrides manifest)
            if (templateConfig.configurationSchema) {
                try {
                    const schema = JSON.parse(templateConfig.configurationSchema);
                    setParsedSchema(schema);

                    // Apply defaults if value not already present
                    const valuesWithDefaults = { ...initialSchemaValues };
                    if (schema.properties) {
                        Object.entries(schema.properties).forEach(([k, v]: [string, any]) => {
                            if (!valuesWithDefaults[k] && v.default) {
                                valuesWithDefaults[k] = String(v.default);
                            }
                        });
                    }
                    setSchemaValues(valuesWithDefaults);
                    // Sync defaults to envVars immediately
                    const envWithDefaults = { ...newEnv };
                    Object.entries(valuesWithDefaults).forEach(([k, v]) => {
                        envWithDefaults[k] = { plainText: v, validationRegex: "" };
                    });
                    setEnvVars(envWithDefaults);
                    checkSchemaValidity(valuesWithDefaults, schema);

                    // Reset manifest if schema present
                    setManifestApplied(null);
                    setSuggestedKeys({});

                } catch (e) {
                    console.error("Failed to parse configuration schema", e);
                    setParsedSchema(null);
                }
            } else {
                setParsedSchema(null);
                setIsSchemaValid(true);
            }

            apiClient.listCredentials().then(setCredentials).catch(console.error);
        }
    }, [open, templateConfig]);

    const checkSchemaValidity = (values: Record<string, string>, schema: any) => {
        if (!schema || !schema.required) {
            setIsSchemaValid(true);
            return;
        }
        const missing = schema.required.some((field: string) => !values[field]);
        setIsSchemaValid(!missing);
    };

    const handleSchemaChange = (newValues: Record<string, string>) => {
        setSchemaValues(newValues);
        // Sync to envVars for compatibility with handleInstantiate logic
        const newEnv: Record<string, SecretValue> = { ...envVars }; // Preserve existing secrets if any (though schema mode might overwrite)

        // Update with new values
        Object.entries(newValues).forEach(([k, v]) => {
            newEnv[k] = { plainText: v, validationRegex: "" };
        });

        setEnvVars(newEnv);
        checkSchemaValidity(newValues, parsedSchema);
    };

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
                newConfig.commandLineService.env = envVars;
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
            if (onComplete) onComplete();

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
                            {manifestApplied && (
                                <Alert className="bg-blue-50/50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800">
                                    <Sparkles className="h-4 w-4 text-blue-500" />
                                    <AlertTitle className="text-blue-700 dark:text-blue-300">Configuration Auto-Detected</AlertTitle>
                                    <AlertDescription className="text-blue-600/80 dark:text-blue-400/80 text-xs">
                                        We've pre-filled the configuration based on community standards.
                                        {manifestDescription && <span className="block mt-1 font-medium">{manifestDescription}</span>}
                                    </AlertDescription>
                                </Alert>
                            )}

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
                                {parsedSchema ? (
                                    <>
                                        <Label>Configuration</Label>
                                        <SchemaForm schema={parsedSchema} value={schemaValues} onChange={handleSchemaChange} />
                                    </>
                                ) : (
                                    <EnvVarEditor
                                        key={name}
                                        initialEnv={envVars}
                                        suggestedKeys={suggestedKeys}
                                        onChange={setEnvVars}
                                    />
                                )}
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
                    <Button onClick={handleInstantiate} disabled={loading || (parsedSchema && !isSchemaValid)}>
                        {loading ? "Creating..." : "Create Instance"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
