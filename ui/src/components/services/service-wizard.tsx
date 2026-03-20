/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { ServiceTemplateSelector } from "@/components/services/service-template-selector";
import { ServiceTemplate } from "@/lib/templates";
import { useToast } from "@/hooks/use-toast";
import { ChevronLeft, ChevronRight, Loader2, CheckCircle2 } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface ServiceWizardProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSuccess: () => void;
}

export function ServiceWizard({ open, onOpenChange, onSuccess }: ServiceWizardProps) {
    const [step, setStep] = useState(1);
    const [config, setConfig] = useState<Partial<UpstreamServiceConfig>>({
        name: "",
        id: "",
        version: "1.0.0",
        disable: false,
        priority: 0,
    });
    const [selectedTemplate, setSelectedTemplate] = useState<ServiceTemplate | null>(null);
    const [loading, setLoading] = useState(false);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<any>(null);
    const { toast } = useToast();

    // Specific auth state to build the upstreamAuth object
    const [authType, setAuthType] = useState<"none" | "apiKey" | "bearerToken" | "basic" | "oauth2">("none");
    const [apiKeyHeader, setApiKeyHeader] = useState("Authorization");
    const [apiKeyValue, setApiKeyValue] = useState("");
    const [bearerToken, setBearerToken] = useState("");

    const reset = () => {
        setStep(1);
        setConfig({
            name: "",
            id: "",
            version: "1.0.0",
            disable: false,
            priority: 0,
        });
        setSelectedTemplate(null);
        setAuthType("none");
        setTestResult(null);
    };

    const handleOpenChange = (newOpen: boolean) => {
        if (!newOpen) {
            reset();
        }
        onOpenChange(newOpen);
    };

    const handleTemplateSelect = (template: ServiceTemplate) => {
        setSelectedTemplate(template);
        // Deep clone config
        const newConfig = structuredClone(template.serviceConfig);
        newConfig.id = ""; // Ensure it's treated as new
        setConfig(newConfig);

        // Pre-fill auth type if template has it
        if (template.authType) {
             if (template.authType === 'apiKey') setAuthType('apiKey');
             else if (template.authType === 'token') setAuthType('bearerToken');
             else if (template.authType === 'basic') setAuthType('basic');
             else if (template.authType === 'oauth2') setAuthType('oauth2');
        }

        setStep(2);
    };

    const buildFinalConfig = (): UpstreamServiceConfig => {
        const finalConfig = { ...config } as UpstreamServiceConfig;

        // Build UpstreamAuth based on wizard state if it was modified
        if (authType !== "none") {
            finalConfig.upstreamAuth = finalConfig.upstreamAuth || {};
            if (authType === "apiKey" && apiKeyValue) {
                finalConfig.upstreamAuth.apiKey = {
                    headerName: apiKeyHeader,
                    value: apiKeyValue
                };
            } else if (authType === "bearerToken" && bearerToken) {
                 finalConfig.upstreamAuth.bearerToken = {
                     token: bearerToken
                 };
            }
        }

        // Generate ID from name if missing
        if (!finalConfig.id && finalConfig.name) {
            finalConfig.id = finalConfig.name.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
        }

        return finalConfig;
    };

    const handleTest = async () => {
        setTesting(true);
        setTestResult(null);
        try {
            const finalConfig = buildFinalConfig();
            const res = await apiClient.validateService(finalConfig);
            setTestResult(res);
            if (!res.valid) {
                 toast({
                    title: "Validation Failed",
                    description: res.error || res.message || "Configuration failed validation.",
                    variant: "destructive"
                });
            } else {
                toast({
                    title: "Validation Successful",
                    description: "Configuration is valid.",
                });
            }
        } catch (e: any) {
            setTestResult({ valid: false, error: e.message || String(e) });
            toast({
                title: "Validation Failed",
                description: e.message || String(e),
                variant: "destructive"
            });
        } finally {
            setTesting(false);
        }
    };

    const handleSave = async () => {
        setLoading(true);
        try {
            const finalConfig = buildFinalConfig();
            await apiClient.registerService(finalConfig);
            toast({ title: "Service Created", description: "New service registered successfully." });
            onSuccess();
            handleOpenChange(false);
        } catch (e: any) {
            console.error("Failed to save service", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: e.message || "Failed to save service configuration."
            });
        } finally {
            setLoading(false);
        }
    };

    const renderStepContent = () => {
        switch (step) {
            case 1:
                return (
                    <div className="h-[400px] overflow-y-auto pr-2">
                        <ServiceTemplateSelector onSelect={handleTemplateSelect} />
                        <div className="mt-4 flex justify-center">
                             <Button variant="outline" onClick={() => {
                                 setSelectedTemplate(null);
                                 setConfig({ name: "", id: "", version: "1.0.0", disable: false, priority: 0 });
                                 setStep(2);
                             }}>
                                 Start from Blank Service
                             </Button>
                        </div>
                    </div>
                );
            case 2:
                return (
                    <div className="space-y-4 py-4 h-[400px] overflow-y-auto pr-2">
                        <div className="space-y-2">
                            <Label htmlFor="name">Service Name <span className="text-destructive">*</span></Label>
                            <Input
                                id="name"
                                value={config.name || ""}
                                onChange={(e) => setConfig({ ...config, name: e.target.value })}
                                placeholder="e.g. My Custom API"
                                autoFocus
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="id">Service ID <span className="text-muted-foreground text-xs font-normal">(Optional, auto-generated if empty)</span></Label>
                            <Input
                                id="id"
                                value={config.id || ""}
                                onChange={(e) => setConfig({ ...config, id: e.target.value })}
                                placeholder="my-custom-api"
                            />
                        </div>
                         <div className="space-y-2">
                            <Label htmlFor="version">Version</Label>
                            <Input
                                id="version"
                                value={config.version || ""}
                                onChange={(e) => setConfig({ ...config, version: e.target.value })}
                                placeholder="1.0.0"
                            />
                        </div>
                        <div className="space-y-2 pt-4 border-t">
                            <Label className="text-base font-semibold">Base Configuration</Label>
                            {/* Simplified type selector if blank, else just show what it is */}
                            {!selectedTemplate ? (
                                <div className="space-y-4">
                                   <div className="space-y-2">
                                        <Label>Service Type</Label>
                                        <Select
                                            value={config.httpService ? "http" : config.commandLineService ? "cmd" : "none"}
                                            onValueChange={(v) => {
                                                if (v === "http") {
                                                    setConfig({...config, httpService: { baseUrl: "" }, commandLineService: undefined, mcpService: undefined});
                                                } else if (v === "cmd") {
                                                    setConfig({...config, commandLineService: { command: "", workingDirectory: "", env: {} }, httpService: undefined, mcpService: undefined});
                                                }
                                            }}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="Select Type" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="none">Not Configured</SelectItem>
                                                <SelectItem value="http">HTTP / REST</SelectItem>
                                                <SelectItem value="cmd">Command Line</SelectItem>
                                            </SelectContent>
                                        </Select>
                                   </div>

                                   {config.httpService && (
                                       <div className="space-y-2">
                                           <Label>Base URL <span className="text-destructive">*</span></Label>
                                           <Input
                                               value={config.httpService.baseUrl || ""}
                                               onChange={(e) => setConfig({
                                                   ...config,
                                                   httpService: { ...config.httpService, baseUrl: e.target.value }
                                               })}
                                               placeholder="https://api.example.com"
                                           />
                                       </div>
                                   )}
                                   {config.commandLineService && (
                                        <div className="space-y-2">
                                           <Label>Command <span className="text-destructive">*</span></Label>
                                           <Input
                                               value={config.commandLineService.command || ""}
                                               onChange={(e) => setConfig({
                                                   ...config,
                                                   commandLineService: { ...config.commandLineService, command: e.target.value }
                                               })}
                                               placeholder="npx @mcpx/server"
                                           />
                                       </div>
                                   )}
                                </div>
                            ) : (
                                <div className="text-sm text-muted-foreground p-3 bg-muted/20 rounded-md border">
                                    Pre-configured via template: <strong>{selectedTemplate.name}</strong>
                                    {config.httpService && <div className="mt-1 font-mono text-xs">{config.httpService.baseUrl}</div>}
                                    {config.commandLineService && <div className="mt-1 font-mono text-xs">{config.commandLineService.command}</div>}
                                </div>
                            )}
                        </div>
                    </div>
                );
            case 3:
                return (
                    <div className="space-y-4 py-4 h-[400px] overflow-y-auto pr-2">
                        <div className="space-y-2">
                            <Label>Authentication Type</Label>
                            <Select value={authType} onValueChange={(v: any) => setAuthType(v)}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="none">No Authentication</SelectItem>
                                    <SelectItem value="apiKey">API Key (Header)</SelectItem>
                                    <SelectItem value="bearerToken">Bearer Token</SelectItem>
                                    <SelectItem value="basic" disabled>Basic Auth (Not supported in wizard yet)</SelectItem>
                                    <SelectItem value="oauth2" disabled>OAuth 2.0 (Configure via Editor later)</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        {authType === "apiKey" && (
                            <div className="space-y-4 pt-4 border-t">
                                <div className="space-y-2">
                                    <Label>Header Name</Label>
                                    <Input
                                        value={apiKeyHeader}
                                        onChange={(e) => setApiKeyHeader(e.target.value)}
                                        placeholder="X-API-Key"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>API Key Value (can use {'${ENV_VAR}'})</Label>
                                    <Input
                                        value={apiKeyValue}
                                        onChange={(e) => setApiKeyValue(e.target.value)}
                                        placeholder="my-secret-key or ${API_KEY}"
                                        type="password"
                                    />
                                </div>
                            </div>
                        )}

                         {authType === "bearerToken" && (
                            <div className="space-y-4 pt-4 border-t">
                                <div className="space-y-2">
                                    <Label>Token (can use {'${ENV_VAR}'})</Label>
                                    <Input
                                        value={bearerToken}
                                        onChange={(e) => setBearerToken(e.target.value)}
                                        placeholder="eyJh... or ${BEARER_TOKEN}"
                                        type="password"
                                    />
                                </div>
                            </div>
                        )}
                    </div>
                );
            case 4:
                const finalConfig = buildFinalConfig();
                return (
                    <div className="space-y-6 py-4 h-[400px] overflow-y-auto pr-2">
                        <div className="rounded-lg border bg-card p-4 space-y-3">
                            <div>
                                <h4 className="font-semibold text-sm text-muted-foreground uppercase tracking-wider mb-1">Identity</h4>
                                <div className="grid grid-cols-2 gap-2 text-sm">
                                    <span className="font-medium">Name:</span> <span>{finalConfig.name || <span className="text-destructive">Missing</span>}</span>
                                    <span className="font-medium">ID:</span> <span className="font-mono">{finalConfig.id || <span className="text-muted-foreground italic">Will be generated</span>}</span>
                                </div>
                            </div>
                            <div className="pt-3 border-t">
                                <h4 className="font-semibold text-sm text-muted-foreground uppercase tracking-wider mb-1">Connection</h4>
                                <div className="text-sm">
                                    {finalConfig.httpService ? (
                                        <div className="flex gap-2"><span className="font-medium">HTTP:</span> <span className="font-mono">{finalConfig.httpService.baseUrl}</span></div>
                                    ) : finalConfig.commandLineService ? (
                                         <div className="flex gap-2"><span className="font-medium">CMD:</span> <span className="font-mono">{finalConfig.commandLineService.command}</span></div>
                                    ) : (
                                        <span className="text-destructive">No connection type configured.</span>
                                    )}
                                </div>
                            </div>
                            <div className="pt-3 border-t">
                                <h4 className="font-semibold text-sm text-muted-foreground uppercase tracking-wider mb-1">Authentication</h4>
                                <div className="text-sm">
                                    {authType === "none" ? "None" : authType === "apiKey" ? "API Key" : "Bearer Token"}
                                </div>
                            </div>
                        </div>

                        <div className="flex flex-col gap-2">
                            <Button variant="secondary" onClick={handleTest} disabled={testing || !finalConfig.name} className="w-full">
                                {testing ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <CheckCircle2 className="h-4 w-4 mr-2" />}
                                Test Connection
                            </Button>
                            {testResult && (
                                <div className={`p-3 text-sm rounded border ${testResult.valid ? 'bg-green-50 text-green-700 border-green-200' : 'bg-red-50 text-red-700 border-red-200'}`}>
                                    {testResult.valid ? "Validation successful. Tools can be discovered." : `Failed: ${testResult.error || testResult.message}`}
                                </div>
                            )}
                        </div>
                    </div>
                );
            default:
                return null;
        }
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogContent className="sm:max-w-xl backdrop-blur-md">
                <DialogHeader>
                    <DialogTitle>
                        {step === 1 && "Choose a Template"}
                        {step === 2 && "Service Details"}
                        {step === 3 && "Authentication"}
                        {step === 4 && "Review & Create"}
                    </DialogTitle>
                    <DialogDescription>
                         {step === 1 && "Start from a pre-configured template or create a blank service."}
                         {step === 2 && "Provide the core identity and connection details."}
                         {step === 3 && "Configure how MCP Any will authenticate with the upstream service."}
                         {step === 4 && "Review your configuration before saving."}
                    </DialogDescription>
                </DialogHeader>

                <div className="flex items-center gap-2 mb-4">
                     {[1, 2, 3, 4].map(s => (
                         <div key={s} className="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
                             <div className={`h-full bg-primary transition-all duration-300 ${s <= step ? 'w-full' : 'w-0'}`} />
                         </div>
                     ))}
                </div>

                {renderStepContent()}

                <div className="flex justify-between items-center mt-6 pt-4 border-t">
                    <Button
                        variant="ghost"
                        onClick={() => setStep(step - 1)}
                        disabled={step === 1 || loading}
                    >
                        <ChevronLeft className="h-4 w-4 mr-1" /> Back
                    </Button>

                    {step < 4 ? (
                         <Button onClick={() => setStep(step + 1)} disabled={step === 2 && !config.name}>
                            Next <ChevronRight className="h-4 w-4 ml-1" />
                        </Button>
                    ) : (
                        <Button onClick={handleSave} disabled={loading || !config.name}>
                            {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                            Create Service
                        </Button>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    );
}
