/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { EnvVarEditor } from "@/components/services/env-var-editor";
import { SecretValue } from "@proto/config/v1/auth";
import { ChevronRight, ChevronLeft, CheckCircle2, Package, Terminal, Settings2, ShieldCheck, Rocket, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { SchemaForm } from "../schema-form";

interface InstallWizardProps {
    initialRepo?: string;
    initialName?: string;
    initialDescription?: string;
    templateId?: string;
}

type WizardStep = "identity" | "connection" | "config" | "auth" | "review";

const STEPS: { id: WizardStep; label: string; icon: any }[] = [
    { id: "identity", label: "Identity", icon: Package },
    { id: "connection", label: "Connection", icon: Terminal },
    { id: "config", label: "Configuration", icon: Settings2 },
    { id: "auth", label: "Authentication", icon: ShieldCheck },
    { id: "review", label: "Review", icon: Rocket },
];

export function InstallWizard({ initialRepo, initialName, initialDescription, templateId }: InstallWizardProps) {
    const router = useRouter();
    const { toast } = useToast();
    const [currentStep, setCurrentStep] = useState<WizardStep>("identity");
    const [loading, setLoading] = useState(false);
    const [templateLoading, setTemplateLoading] = useState(!!templateId);
    const [credentials, setCredentials] = useState<any[]>([]);

    // Form State
    const [name, setName] = useState(initialName || "");
    const [description, setDescription] = useState(initialDescription || "");
    const [repoUrl, setRepoUrl] = useState(initialRepo || "");

    // Connection State
    const [command, setCommand] = useState("");
    const [workingDir, setWorkingDir] = useState("");

    // Config State
    const [envVars, setEnvVars] = useState<Record<string, SecretValue>>({});

    // Schema State
    const [parsedSchema, setParsedSchema] = useState<any>(null);
    const [schemaValues, setSchemaValues] = useState<Record<string, string>>({});
    const [isSchemaValid, setIsSchemaValid] = useState(true);

    // Auth State
    const [authId, setAuthId] = useState("none");

    // Load credentials on mount
    useEffect(() => {
        apiClient.listCredentials().then(setCredentials).catch(console.error);
    }, []);

    // Load template if ID provided
    useEffect(() => {
        if (!templateId) return;

        const fetchTemplate = async () => {
            try {
                const templates = await apiClient.listTemplates();
                // Match by ID or Name (fallback)
                const t = templates.find(t => t.id === templateId || t.name === templateId);
                if (t && t.serviceConfig) {
                    const conf = t.serviceConfig;
                    setName(`${conf.name}-copy`);
                    setDescription(conf.description || description);

                    if (conf.commandLineService) {
                        setCommand(conf.commandLineService.command || "");
                        setWorkingDir(conf.commandLineService.workingDirectory || "");

                        // Env Vars
                        if (conf.commandLineService.env) {
                            const newEnv: Record<string, SecretValue> = {};
                            const initialSchemaValues: Record<string, string> = {};
                            Object.entries(conf.commandLineService.env).forEach(([k, v]) => {
                                const sv = v as any;
                                let val = "";
                                if (sv && typeof sv === 'object' && sv.plainText) {
                                    val = sv.plainText;
                                    newEnv[k] = { plainText: val, validationRegex: "" };
                                } else {
                                    newEnv[k] = { plainText: String(v), validationRegex: "" };
                                    val = String(v);
                                }
                                initialSchemaValues[k] = val;
                            });
                            setEnvVars(newEnv);
                            setSchemaValues(initialSchemaValues);
                        }
                    }

                    // Schema Parsing
                    if (conf.configurationSchema) {
                        try {
                            const schema = JSON.parse(conf.configurationSchema);
                            setParsedSchema(schema);
                            // Apply schema defaults if needed?
                            // Logic from instantiate-dialog:
                            // Apply defaults if value not already present
                            // We already populated initialSchemaValues from env.
                            setSchemaValues(prev => {
                                const next = { ...prev };
                                if (schema.properties) {
                                    Object.entries(schema.properties).forEach(([k, v]: [string, any]) => {
                                        if (!next[k] && v.default) {
                                            next[k] = String(v.default);
                                        }
                                    });
                                }
                                return next;
                            });
                        } catch (e) {
                            console.error("Failed to parse schema", e);
                        }
                    }
                }
            } catch (e) {
                console.error("Failed to fetch template", e);
                toast({ variant: "destructive", title: "Error", description: "Failed to load template" });
            } finally {
                setTemplateLoading(false);
            }
        };
        fetchTemplate();
    }, [templateId, description, toast]);

    // Smart Detection Effect (only if no template loaded)
    useEffect(() => {
        if (!templateId && repoUrl && !command) {
            // Heuristic for command detection
            if (repoUrl.includes("github.com")) {
                const match = repoUrl.match(/github\.com\/([^/]+)\/([^/]+)/);
                if (match) {
                    const owner = match[1];
                    const repo = match[2];

                    // Python heuristic
                    const isPython = repo.includes("python") || repo.includes("py-");
                    if (isPython) {
                        setCommand(`uvx ${repo}`);
                    } else if (owner === 'modelcontextprotocol') {
                        setCommand(`npx -y @modelcontextprotocol/${repo}`);
                    } else {
                        setCommand(`npx -y ${repo}`); // Generic guess
                    }
                }
            }
        }
    }, [repoUrl, templateId, command]);

    // Initialize name from repo if empty and no template
    useEffect(() => {
        if (!templateId && repoUrl && !name) {
            const parts = repoUrl.split("/");
            const lastPart = parts[parts.length - 1];
            if (lastPart) {
                // remove .git if present
                const cleanName = lastPart.replace(".git", "").replace(/^server-/, "");
                // Title case
                setName(cleanName.charAt(0).toUpperCase() + cleanName.slice(1));
            }
        }
    }, [repoUrl, templateId, name]);

    const handleNext = () => {
        const idx = STEPS.findIndex(s => s.id === currentStep);
        if (idx < STEPS.length - 1) {
            setCurrentStep(STEPS[idx + 1].id);
        }
    };

    const handleBack = () => {
        const idx = STEPS.findIndex(s => s.id === currentStep);
        if (idx > 0) {
            setCurrentStep(STEPS[idx - 1].id);
        }
    };

    const handleSchemaChange = (newValues: Record<string, string>) => {
        setSchemaValues(newValues);
        // Sync to envVars
        const newEnv: Record<string, SecretValue> = { ...envVars };
        Object.entries(newValues).forEach(([k, v]) => {
            newEnv[k] = { plainText: v, validationRegex: "" };
        });
        setEnvVars(newEnv);
    };

    const handleDeploy = async () => {
        setLoading(true);
        try {
            const config: UpstreamServiceConfig = {
                id: name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
                name: name,
                description: description,
                version: "1.0.0",
                commandLineService: {
                    command: command,
                    env: envVars,
                    workingDirectory: workingDir,
                    tools: [],
                    resources: [],
                    prompts: [],
                    calls: {},
                    communicationProtocol: 0,
                    local: false
                },
                disable: false,
                sanitizedName: name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
                priority: 0,
                loadBalancingStrategy: 0,
                callPolicies: [],
                preCallHooks: [],
                postCallHooks: [],
                prompts: [],
                autoDiscoverTool: true,
                configError: "",
                tags: ["community", "wizard"],
                readOnly: false
            };

            if (authId !== 'none') {
                const cred = credentials.find(c => c.id === authId);
                if (cred && cred.authentication) {
                    config.upstreamAuth = cred.authentication;
                }
            }

            await apiClient.registerService(config);
            toast({
                title: "Service Deployed",
                description: `${name} has been successfully installed.`,
            });
            router.push(`/upstream-services/${config.name}`);
        } catch (e) {
            console.error(e);
            toast({
                variant: "destructive",
                title: "Deployment Failed",
                description: String(e),
            });
        } finally {
            setLoading(false);
        }
    };

    if (templateLoading) {
        return (
            <div className="flex flex-col items-center justify-center h-64 gap-4">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                <p className="text-muted-foreground">Loading template configuration...</p>
            </div>
        );
    }

    const currentStepIdx = STEPS.findIndex(s => s.id === currentStep);
    const progress = ((currentStepIdx + 1) / STEPS.length) * 100;

    return (
        <div className="flex flex-col gap-8 max-w-4xl mx-auto w-full">
            {/* Stepper */}
            <div className="relative">
                <div className="absolute top-1/2 left-0 w-full h-1 bg-muted -z-10 rounded-full" />
                <div
                    className="absolute top-1/2 left-0 h-1 bg-primary -z-10 rounded-full transition-all duration-500 ease-in-out"
                    style={{ width: `${progress}%` }}
                />
                <div className="flex justify-between">
                    {STEPS.map((step, idx) => {
                        const Icon = step.icon;
                        const isActive = idx <= currentStepIdx;
                        const isCurrent = idx === currentStepIdx;

                        return (
                            <div key={step.id} className="flex flex-col items-center gap-2 bg-background px-2">
                                <div className={cn(
                                    "w-10 h-10 rounded-full flex items-center justify-center border-2 transition-all duration-300",
                                    isActive ? "border-primary bg-primary text-primary-foreground shadow-lg scale-110" : "border-muted-foreground/30 text-muted-foreground bg-muted/10",
                                    isCurrent && "ring-4 ring-primary/20"
                                )}>
                                    <Icon className="w-5 h-5" />
                                </div>
                                <span className={cn(
                                    "text-xs font-medium transition-colors duration-300",
                                    isActive ? "text-foreground" : "text-muted-foreground"
                                )}>
                                    {step.label}
                                </span>
                            </div>
                        );
                    })}
                </div>
            </div>

            <Card className="flex-1 shadow-lg border-muted/40 overflow-hidden">
                <CardHeader className="border-b bg-muted/10 pb-6">
                    <CardTitle className="text-2xl">{STEPS[currentStepIdx].label}</CardTitle>
                    <CardDescription>
                        {currentStep === "identity" && "Define the identity of your new service."}
                        {currentStep === "connection" && "Configure how MCP Any connects to the service."}
                        {currentStep === "config" && "Set up environment variables and runtime configuration."}
                        {currentStep === "auth" && "Bind existing credentials for authentication."}
                        {currentStep === "review" && "Review your configuration before deployment."}
                    </CardDescription>
                </CardHeader>

                <CardContent className="p-8 min-h-[400px]">
                    <div className="max-w-2xl mx-auto space-y-6">
                        {currentStep === "identity" && (
                            <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
                                <div className="space-y-2">
                                    <Label htmlFor="name">Service Name</Label>
                                    <Input
                                        id="name"
                                        placeholder="e.g., Linear Integration"
                                        value={name}
                                        onChange={(e) => setName(e.target.value)}
                                        className="text-lg py-6"
                                        autoFocus
                                    />
                                    <p className="text-sm text-muted-foreground">The unique display name for this service.</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="desc">Description</Label>
                                    <Textarea
                                        id="desc"
                                        placeholder="What does this service do?"
                                        value={description}
                                        onChange={(e) => setDescription(e.target.value)}
                                        rows={3}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="repo">Source Repository</Label>
                                    <div className="relative">
                                        <Package className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                        <Input
                                            id="repo"
                                            placeholder="https://github.com/..."
                                            value={repoUrl}
                                            onChange={(e) => setRepoUrl(e.target.value)}
                                            className="pl-10 font-mono text-sm"
                                        />
                                    </div>
                                </div>
                            </div>
                        )}

                        {currentStep === "connection" && (
                            <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
                                <div className="bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900 rounded-lg p-4 mb-4">
                                    <div className="flex items-start gap-3">
                                        <Terminal className="h-5 w-5 text-blue-600 mt-0.5" />
                                        <div>
                                            <h4 className="font-medium text-blue-900 dark:text-blue-100">Stdio Integration</h4>
                                            <p className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                                                MCP Any will run this command as a subprocess and communicate via standard input/output.
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-2">
                                    <Label htmlFor="command">Command</Label>
                                    <Textarea
                                        id="command"
                                        placeholder="npx -y @modelcontextprotocol/server-..."
                                        value={command}
                                        onChange={(e) => setCommand(e.target.value)}
                                        className="font-mono text-sm bg-black/5 dark:bg-black/40 min-h-[100px]"
                                    />
                                    <p className="text-xs text-muted-foreground">
                                        Full command to execute. Use <code>uvx</code> for Python or <code>npx -y</code> for Node.js.
                                    </p>
                                </div>

                                <div className="space-y-2">
                                    <Label htmlFor="cwd">Working Directory (Optional)</Label>
                                    <Input
                                        id="cwd"
                                        placeholder="/app/data"
                                        value={workingDir}
                                        onChange={(e) => setWorkingDir(e.target.value)}
                                        className="font-mono text-sm"
                                    />
                                </div>
                            </div>
                        )}

                        {currentStep === "config" && (
                            <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
                                <p className="text-sm text-muted-foreground mb-4">
                                    Define environment variables required by the service (e.g., API Keys, Config Options).
                                </p>
                                {parsedSchema ? (
                                    <SchemaForm
                                        schema={parsedSchema}
                                        value={schemaValues}
                                        onChange={handleSchemaChange}
                                    />
                                ) : (
                                    <EnvVarEditor initialEnv={envVars} onChange={setEnvVars} />
                                )}
                            </div>
                        )}

                        {currentStep === "auth" && (
                            <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
                                <div className="space-y-2">
                                    <Label>Service Credential</Label>
                                    <Select value={authId} onValueChange={setAuthId}>
                                        <SelectTrigger className="w-full">
                                            <SelectValue placeholder="Select credential..." />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="none">
                                                <span className="text-muted-foreground italic">No Authentication</span>
                                            </SelectItem>
                                            {credentials.map((c) => (
                                                <SelectItem key={c.id} value={c.id} className="font-medium">
                                                    {c.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <p className="text-sm text-muted-foreground mt-2">
                                        Select a stored credential to authenticate requests <strong>from MCP Any to the Upstream Service</strong>.
                                        This is usually NOT required for Stdio services unless they expect specific auth headers (uncommon).
                                    </p>
                                </div>
                            </div>
                        )}

                        {currentStep === "review" && (
                            <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
                                <div className="bg-muted/30 rounded-lg p-6 space-y-4 border">
                                    <div className="flex justify-between items-start border-b pb-4">
                                        <div>
                                            <h3 className="text-lg font-bold">{name}</h3>
                                            <p className="text-sm text-muted-foreground">{description || "No description provided"}</p>
                                        </div>
                                        {repoUrl && (
                                            <a href={repoUrl} target="_blank" rel="noreferrer" className="text-xs text-blue-500 hover:underline">
                                                View Source
                                            </a>
                                        )}
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                                        <div>
                                            <span className="text-muted-foreground block mb-1">Command</span>
                                            <code className="bg-black/80 text-green-400 p-1.5 rounded text-xs block break-all">
                                                {command || "No command specified"}
                                            </code>
                                        </div>
                                        <div>
                                            <span className="text-muted-foreground block mb-1">Environment Vars</span>
                                            <div className="font-mono text-xs">
                                                {Object.keys(envVars).length > 0 ? (
                                                    <ul className="list-disc list-inside">
                                                        {Object.keys(envVars).map(k => (
                                                            <li key={k}>{k}</li>
                                                        ))}
                                                    </ul>
                                                ) : (
                                                    <span className="italic opacity-50">None configured</span>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex items-center gap-2 p-4 bg-yellow-50 dark:bg-yellow-900/10 text-yellow-800 dark:text-yellow-200 rounded-lg text-sm">
                                    <CheckCircle2 className="h-5 w-5 shrink-0" />
                                    <p>Ready to deploy. The service will be registered and started immediately.</p>
                                </div>
                            </div>
                        )}
                    </div>
                </CardContent>

                <CardFooter className="flex justify-between border-t p-6 bg-muted/10">
                    <Button
                        variant="outline"
                        onClick={handleBack}
                        disabled={currentStep === "identity"}
                        className="w-32"
                    >
                        <ChevronLeft className="mr-2 h-4 w-4" /> Back
                    </Button>

                    {currentStep === "review" ? (
                        <Button
                            onClick={handleDeploy}
                            disabled={loading}
                            className="w-32 bg-green-600 hover:bg-green-700 text-white"
                        >
                            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Rocket className="mr-2 h-4 w-4" />}
                            Deploy
                        </Button>
                    ) : (
                        <Button
                            onClick={handleNext}
                            disabled={!name && currentStep === "identity"}
                            className="w-32"
                        >
                            Next <ChevronRight className="ml-2 h-4 w-4" />
                        </Button>
                    )}
                </CardFooter>
            </Card>
        </div>
    );
}
