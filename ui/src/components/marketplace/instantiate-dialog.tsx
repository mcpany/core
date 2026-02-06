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
import { analyzeRepository, AnalysisResult } from "@/lib/repo-analyzer";
import { Loader2, Sparkles, CheckCircle2, AlertTriangle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface InstantiateDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    templateConfig?: UpstreamServiceConfig;
    onComplete: () => void;
    repoUrl?: string;
}

/**
 * InstantiateDialog.
 *
 * @param onComplete - The onComplete.
 */
export function InstantiateDialog({ open, onOpenChange, templateConfig, onComplete, repoUrl }: InstantiateDialogProps) {
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

    // Analysis state
    const [isAnalyzing, setIsAnalyzing] = useState(false);
    const [analysisResult, setAnalysisResult] = useState<AnalysisResult | null>(null);

    // Helper to determine if we should show CLI options
    // Community servers (which are created dynamically) usually have commandLineService populated.
    // Also true if we are importing from Repo URL (which implies CLI).
    const isCommandLine = !!templateConfig?.commandLineService || !!repoUrl;

    useEffect(() => {
        if (open) {
            // Reset common state
            setAuthId("none");
            setAnalysisResult(null);

            if (repoUrl) {
                // Mode: Import from Repo URL
                setIsAnalyzing(true);
                analyzeRepository(repoUrl).then(result => {
                    setAnalysisResult(result);
                    if (result.name) setName(result.name);
                    if (result.command) setCommand(result.command);
                    if (result.envVars) {
                        setEnvVars(result.envVars);
                    }
                    setIsAnalyzing(false);

                    if (result.detectedType !== 'unknown') {
                         toast({ title: "Repository Analyzed", description: `Detected ${result.detectedType} project.` });
                    }
                });

                // Set defaults while analyzing or if analysis fails
                const parts = repoUrl.split('/');
                const repoName = parts[parts.length - 1] || "imported-service";
                setName(repoName);

            } else if (templateConfig) {
                // Mode: Template / Existing Config
                setName(`${templateConfig.name}-copy`);

                if (templateConfig.commandLineService) {
                    // Initialize command
                    setCommand(templateConfig.commandLineService.command || "");

                    // Initialize env vars
                    const newEnv: Record<string, SecretValue> = {};
                    const initialSchemaValues: Record<string, string> = {};

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
                    setEnvVars(newEnv);
                    setSchemaValues(initialSchemaValues);
                } else {
                    setCommand("");
                    setEnvVars({});
                    setSchemaValues({});
                }

                // Parse Schema if available
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

                    } catch (e) {
                        console.error("Failed to parse configuration schema", e);
                        setParsedSchema(null);
                    }
                } else {
                    setParsedSchema(null);
                    setIsSchemaValid(true);
                }
            }

            apiClient.listCredentials().then(setCredentials).catch(console.error);
        }
    }, [open, templateConfig, repoUrl]); // Added repoUrl to deps

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
        setLoading(true);
        try {
            // Construct config
            let newConfig: UpstreamServiceConfig;

            if (repoUrl) {
                // Construct from scratch
                newConfig = {
                    id: name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
                    name: name,
                    version: "1.0.0",
                    disable: false,
                    priority: 0,
                    loadBalancingStrategy: 0,
                    sanitizedName: name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
                    readOnly: false,
                    callPolicies: [],
                    preCallHooks: [],
                    postCallHooks: [],
                    prompts: [],
                    autoDiscoverTool: true,
                    configError: "",
                    tags: ["imported", "cli"],
                    commandLineService: {
                        command: command,
                        env: envVars,
                        workingDirectory: "",
                        tools: [],
                        resources: [],
                        prompts: [],
                        calls: {},
                        communicationProtocol: 0,
                        local: false
                    }
                };
            } else if (templateConfig) {
                 newConfig = { ...templateConfig };
                 newConfig.name = name;
                 newConfig.id = name; // ID is name for now

                 // Update CLI specific fields if applicable
                 if (newConfig.commandLineService) {
                     newConfig.commandLineService.command = command;
                     newConfig.commandLineService.env = envVars;
                 }
            } else {
                return; // Should not happen
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
                    <DialogTitle>{repoUrl ? "Import from Repository" : "Instantiate Service"}</DialogTitle>
                    <DialogDescription>
                        {repoUrl ? "Configure the imported service." : "Create a running instance from this template."}
                    </DialogDescription>
                </DialogHeader>

                {isAnalyzing && (
                    <div className="flex items-center justify-center p-8 text-muted-foreground animate-pulse">
                        <Loader2 className="mr-2 h-5 w-5 animate-spin text-primary" />
                        Analyzing repository structure...
                    </div>
                )}

                {!isAnalyzing && analysisResult && (
                     <Alert className="mb-4 bg-primary/5 border-primary/20">
                         <Sparkles className="h-4 w-4 text-primary" />
                         <AlertTitle>Smart Import Active</AlertTitle>
                         <AlertDescription className="text-xs">
                             Detected <strong>{analysisResult.detectedType}</strong> project.
                             Pre-filled command and found {Object.keys(analysisResult.envVars).length} potential environment variables.
                         </AlertDescription>
                     </Alert>
                )}

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
                                {parsedSchema ? (
                                    <>
                                        <Label>Configuration</Label>
                                        <SchemaForm schema={parsedSchema} value={schemaValues} onChange={handleSchemaChange} />
                                    </>
                                ) : (
                                    <EnvVarEditor initialEnv={envVars} onChange={setEnvVars} />
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
