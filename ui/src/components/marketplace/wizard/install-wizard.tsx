/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { EnvVarEditor } from "@/components/services/env-var-editor";
import { SecretValue } from "@proto/config/v1/auth";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";
import { CheckCircle2, ChevronRight, ChevronLeft, Loader2, Rocket, Settings2, ShieldCheck, Terminal } from "lucide-react";
import { cn } from "@/lib/utils";

// Step Definitions
const STEPS = [
    { id: "identity", title: "Identity", icon: Rocket, description: "Name and describe your service." },
    { id: "connection", title: "Connection", icon: Terminal, description: "Configure execution command." },
    { id: "config", title: "Configuration", icon: Settings2, description: "Set environment variables." },
    { id: "review", title: "Review", icon: CheckCircle2, description: "Verify and deploy." },
];

export function InstallWizard() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { toast } = useToast();

    const [currentStep, setCurrentStep] = useState(0);
    const [loading, setLoading] = useState(false);
    const [credentials, setCredentials] = useState<any[]>([]);

    // Form State
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [command, setCommand] = useState("");
    const [workingDir, setWorkingDir] = useState("");
    const [envVars, setEnvVars] = useState<Record<string, SecretValue>>({});
    const [authId, setAuthId] = useState("none");

    // Initialize from URL params
    useEffect(() => {
        const repoParam = searchParams.get("repo");
        const nameParam = searchParams.get("name");
        const descParam = searchParams.get("description");

        if (nameParam) setName(nameParam);
        if (descParam) setDescription(descParam);

        // Heuristic: Suggest command based on Repo
        if (repoParam) {
            // E.g. https://github.com/modelcontextprotocol/server-filesystem
            if (repoParam.includes("modelcontextprotocol/server-")) {
                const pkg = repoParam.split("server-")[1];
                setCommand(`npx -y @modelcontextprotocol/server-${pkg}`);
            } else if (repoParam.includes("github.com")) {
                const parts = repoParam.split("/");
                const repoName = parts[parts.length - 1];
                setCommand(`npx -y ${repoName}`); // Best guess
            }
        }

        // Load credentials
        apiClient.listCredentials().then(setCredentials).catch(console.error);
    }, [searchParams]);

    const handleNext = () => {
        if (currentStep < STEPS.length - 1) {
            setCurrentStep(prev => prev + 1);
        }
    };

    const handleBack = () => {
        if (currentStep > 0) {
            setCurrentStep(prev => prev - 1);
        }
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
                    workingDirectory: workingDir,
                    env: envVars,
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
                tags: ["community", "wizard"]
            };

            if (authId !== 'none') {
                const cred = credentials.find(c => c.id === authId);
                if (cred && cred.authentication) {
                    config.upstreamAuth = cred.authentication;
                }
            }

            await apiClient.registerService(config);
            toast({ title: "Service Deployed", description: `${name} is now running.` });
            router.push(`/upstream-services/${config.name}`);
        } catch (e: any) {
            toast({
                title: "Deployment Failed",
                variant: "destructive",
                description: e.message || String(e)
            });
        } finally {
            setLoading(false);
        }
    };

    // Render Steps
    const renderStepContent = () => {
        switch (STEPS[currentStep].id) {
            case "identity":
                return (
                    <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
                        <div className="grid gap-2">
                            <Label htmlFor="name">Service Name</Label>
                            <Input
                                id="name"
                                value={name}
                                onChange={e => setName(e.target.value)}
                                placeholder="e.g. Linear Integration"
                                autoFocus
                            />
                            <p className="text-xs text-muted-foreground">Unique identifier for this service.</p>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="description">Description</Label>
                            <Textarea
                                id="description"
                                value={description}
                                onChange={e => setDescription(e.target.value)}
                                placeholder="What does this service do?"
                            />
                        </div>
                    </div>
                );
            case "connection":
                return (
                    <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
                        <div className="grid gap-2">
                            <Label htmlFor="command">Execution Command</Label>
                            <div className="relative">
                                <Textarea
                                    id="command"
                                    value={command}
                                    onChange={e => setCommand(e.target.value)}
                                    className="font-mono text-sm bg-black/5 dark:bg-black/50 min-h-[100px]"
                                    placeholder="npx -y @modelcontextprotocol/server-..."
                                />
                                <div className="absolute bottom-2 right-2 flex gap-1">
                                    <Button variant="ghost" size="xs" onClick={() => setCommand("npx -y ")}>npx</Button>
                                    <Button variant="ghost" size="xs" onClick={() => setCommand("uvx ")}>uvx</Button>
                                    <Button variant="ghost" size="xs" onClick={() => setCommand("docker run -i --rm ")}>docker</Button>
                                </div>
                            </div>
                            <p className="text-xs text-muted-foreground">
                                The command to start the MCP server using Stdio.
                            </p>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="cwd">Working Directory (Optional)</Label>
                            <Input
                                id="cwd"
                                value={workingDir}
                                onChange={e => setWorkingDir(e.target.value)}
                                placeholder="/app"
                            />
                        </div>
                    </div>
                );
            case "config":
                return (
                    <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
                        <div className="flex items-center justify-between">
                            <Label>Environment Variables</Label>
                            <Button variant="ghost" size="sm" onClick={() => {
                                // Add common keys hint
                                const common = ["API_KEY", "TOKEN", "SECRET"];
                                // Logic to suggest? For now just manual.
                            }}>
                                <ShieldCheck className="mr-2 h-4 w-4 text-green-500" />
                                Secure Storage Enabled
                            </Button>
                        </div>
                        <EnvVarEditor initialEnv={envVars} onChange={setEnvVars} />

                        <div className="grid gap-2 mt-6 pt-6 border-t">
                            <Label>Authentication (Optional)</Label>
                            <Select value={authId} onValueChange={setAuthId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Link existing credential" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="none">No Auth / Managed by Env Vars</SelectItem>
                                    {credentials.map(c => (
                                        <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <p className="text-xs text-muted-foreground">
                                Link a shared credential (e.g. OAuth) to this service.
                            </p>
                        </div>
                    </div>
                );
            case "review":
                return (
                    <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
                        <div className="rounded-lg border bg-muted/10 p-6 space-y-4">
                            <div className="grid grid-cols-3 gap-4">
                                <div className="text-sm font-medium text-muted-foreground">Name</div>
                                <div className="col-span-2 font-medium">{name}</div>

                                <div className="text-sm font-medium text-muted-foreground">Command</div>
                                <div className="col-span-2 font-mono text-xs bg-muted p-2 rounded break-all">{command}</div>

                                <div className="text-sm font-medium text-muted-foreground">Environment</div>
                                <div className="col-span-2 text-sm">{Object.keys(envVars).length} variables configured</div>

                                <div className="text-sm font-medium text-muted-foreground">Auth</div>
                                <div className="col-span-2 text-sm">{authId === 'none' ? 'None' : 'Linked Credential'}</div>
                            </div>
                        </div>
                        <div className="bg-yellow-50 dark:bg-yellow-900/10 border-l-4 border-yellow-500 p-4">
                            <p className="text-sm text-yellow-700 dark:text-yellow-400">
                                This will start a new process on the server. Ensure the command is safe and the environment is trusted.
                            </p>
                        </div>
                    </div>
                );
            default:
                return null;
        }
    };

    return (
        <div className="max-w-4xl mx-auto py-10 px-4">
            <div className="mb-8">
                <h1 className="text-3xl font-bold tracking-tight mb-2">Install Service</h1>
                <p className="text-muted-foreground">
                    Configure and deploy a new MCP service from the marketplace.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
                {/* Stepper */}
                <div className="space-y-1">
                    {STEPS.map((step, idx) => {
                        const Icon = step.icon;
                        const isActive = idx === currentStep;
                        const isCompleted = idx < currentStep;

                        return (
                            <div
                                key={step.id}
                                className={cn(
                                    "flex items-center gap-3 p-3 rounded-lg transition-colors relative overflow-hidden",
                                    isActive ? "bg-primary/10 text-primary" : "text-muted-foreground",
                                    isCompleted ? "text-foreground" : ""
                                )}
                            >
                                <div className={cn(
                                    "flex items-center justify-center w-8 h-8 rounded-full border text-xs font-bold transition-all",
                                    isActive ? "border-primary bg-primary text-primary-foreground" : "border-muted-foreground/30",
                                    isCompleted ? "border-primary/50 bg-primary/20 text-primary" : ""
                                )}>
                                    {isCompleted ? <CheckCircle2 className="w-4 h-4" /> : idx + 1}
                                </div>
                                <div className="flex-1">
                                    <div className="text-sm font-medium">{step.title}</div>
                                    {isActive && <div className="text-[10px] opacity-80 line-clamp-1">{step.description}</div>}
                                </div>
                                {isActive && (
                                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary" />
                                )}
                            </div>
                        );
                    })}
                </div>

                {/* Content */}
                <div className="md:col-span-3">
                    <Card className="min-h-[400px] flex flex-col shadow-lg border-muted/40">
                        <CardHeader className="border-b bg-muted/5">
                            <CardTitle>{STEPS[currentStep].title}</CardTitle>
                            <CardDescription>{STEPS[currentStep].description}</CardDescription>
                        </CardHeader>
                        <CardContent className="flex-1 p-6">
                            {renderStepContent()}
                        </CardContent>
                        <CardFooter className="flex justify-between border-t p-6 bg-muted/5">
                            <Button
                                variant="outline"
                                onClick={handleBack}
                                disabled={currentStep === 0 || loading}
                            >
                                <ChevronLeft className="mr-2 h-4 w-4" /> Back
                            </Button>

                            {currentStep === STEPS.length - 1 ? (
                                <Button onClick={handleDeploy} disabled={loading}>
                                    {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Rocket className="mr-2 h-4 w-4" />}
                                    Deploy Service
                                </Button>
                            ) : (
                                <Button onClick={handleNext} disabled={!name && currentStep === 0}>
                                    Next <ChevronRight className="ml-2 h-4 w-4" />
                                </Button>
                            )}
                        </CardFooter>
                    </Card>
                </div>
            </div>
        </div>
    );
}
