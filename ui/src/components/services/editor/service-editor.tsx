/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { EnvVarEditor } from "@/components/services/env-var-editor";
import { OAuthConfig } from "@/components/services/editor/oauth-config";
import { OAuthConnect } from "@/components/services/editor/oauth-connect";
import { ScrollArea } from "@/components/ui/scroll-area";
import { AlertCircle, Plus, Trash2, CheckCircle2, XCircle, Loader2, Key } from "lucide-react";
import { SecretPicker } from "@/components/secrets/secret-picker";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { ServiceDiagnostics } from "@/components/services/editor/service-diagnostics";
import { PolicyEditor } from "@/components/services/editor/policy-editor";
import { ServiceInspector } from "@/components/services/editor/service-inspector";
import { SourceEditor } from "@/components/services/editor/source-editor";
import { ResilienceEditor } from "@/components/services/editor/resilience-editor";
import yaml from "js-yaml";

interface ServiceEditorProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
    onSave: () => void;
    onCancel: () => void;
}

/**
 * ServiceEditor.
 *
 * @param onCancel - The onCancel.
 */
export function ServiceEditor({ service, onChange, onSave, onCancel }: ServiceEditorProps) {
    const [activeTab, setActiveTab] = useState("general");
    const [validating, setValidating] = useState(false);
    const [yamlContent, setYamlContent] = useState("");
    const [yamlError, setYamlError] = useState<string | null>(null);
    const { toast } = useToast();

    const updateService = (updates: Partial<UpstreamServiceConfig>) => {
        onChange({ ...service, ...updates });
    };

    const handleTabChange = (value: string) => {
        setActiveTab(value);
        if (value === "source") {
            try {
                // Clone and strip runtime fields
                const cleanService = { ...service };
                delete cleanService.connectionPool;
                delete cleanService.lastError;
                delete cleanService.toolCount;

                const yamlStr = yaml.dump(cleanService);
                setYamlContent(yamlStr);
                setYamlError(null);
            } catch (e) {
                console.error("Failed to dump YAML", e);
                setYamlError("Failed to generate YAML from current configuration.");
            }
        }
    };

    const handleYamlChange = (value: string | undefined) => {
        if (value === undefined) return;
        setYamlContent(value);
        try {
            const parsed = yaml.load(value) as UpstreamServiceConfig;
            if (typeof parsed !== 'object' || parsed === null) {
                throw new Error("Invalid YAML: Must be an object");
            }
            setYamlError(null);
            // Update parent state
            onChange(parsed);
        } catch (e: any) {
            setYamlError(e.message || "Invalid YAML");
        }
    };

    const handleValidate = async () => {
        setValidating(true);
        try {
            const result = await apiClient.validateService(service);
            if (result.valid) {
                toast({
                    title: "Configuration Valid",
                    description: "The service configuration is valid and reachable.",
                    action: <CheckCircle2 className="h-5 w-5 text-green-500" />
                });
            } else {
                 toast({
                    variant: "destructive",
                    title: "Validation Failed",
                    description: result.error || "Unknown validation error.",
                    action: <XCircle className="h-5 w-5 text-destructive-foreground" />
                });
            }
        } catch (e: any) {
            toast({
                variant: "destructive",
                title: "Validation Error",
                description: e.message || "Failed to validate service.",
            });
        } finally {
            setValidating(false);
        }
    };

    const handleTypeChange = (type: string) => {
        const newService = { ...service };
        // Clear old configs
        delete newService.httpService;
        delete newService.grpcService;
        delete newService.commandLineService;
        delete newService.mcpService;
        delete newService.openapiService;

        // Initialize new config with defaults
        if (type === 'http') newService.httpService = { address: "", tools: [], calls: {}, resources: [], prompts: [] };
        if (type === 'grpc') newService.grpcService = { address: "", useReflection: true, tools: [], resources: [], calls: {}, prompts: [], protoDefinitions: [], protoCollection: [] };
        if (type === 'cmd') newService.commandLineService = { command: "", workingDirectory: "", local: false, env: {}, tools: [], resources: [], prompts: [], communicationProtocol: 0, calls: {} };
        if (type === 'mcp') newService.mcpService = { toolAutoDiscovery: true, tools: [], resources: [], calls: {}, prompts: [] };
        if (type === 'openapi') newService.openapiService = { address: "", specUrl: "", tools: [], resources: [], calls: {}, prompts: [] };

        onChange(newService);
    };

    const getType = () => {
        if (service.httpService) return 'http';
        if (service.grpcService) return 'grpc';
        if (service.commandLineService) return 'cmd';
        if (service.mcpService) return 'mcp';
        if (service.openapiService) return 'openapi';
        return 'http'; // Default
    };

    return (
        <div className="flex flex-col h-full">
            <div className="flex-1 overflow-y-auto">
                <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
                    <div className="border-b px-4">
                        <TabsList className="bg-transparent">
                            <TabsTrigger value="general">General</TabsTrigger>
                            <TabsTrigger value="connection">Connection</TabsTrigger>
                            <TabsTrigger value="auth">Authentication</TabsTrigger>
                            <TabsTrigger value="policies">Policies</TabsTrigger>
                            <TabsTrigger value="advanced">Advanced</TabsTrigger>
                            <TabsTrigger value="diagnostics">Diagnostics</TabsTrigger>
                            {service.id && <TabsTrigger value="inspector">Inspector</TabsTrigger>}
                            <TabsTrigger value="source">Source</TabsTrigger>
                        </TabsList>
                    </div>

                    <div className="p-4 space-y-6">
                        <TabsContent value="source" className="space-y-4 mt-0">
                            <div className="space-y-4">
                                <div className="flex items-center justify-between">
                                    <h3 className="text-lg font-medium">Configuration Source</h3>
                                    {yamlError ? (
                                        <div className="flex items-center gap-2 text-destructive text-sm bg-destructive/10 px-3 py-1 rounded">
                                            <AlertCircle className="h-4 w-4" />
                                            {yamlError}
                                        </div>
                                    ) : (
                                        <div className="flex items-center gap-2 text-green-500 text-sm bg-green-500/10 px-3 py-1 rounded">
                                            <CheckCircle2 className="h-4 w-4" />
                                            Valid YAML
                                        </div>
                                    )}
                                </div>
                                <p className="text-sm text-muted-foreground">
                                    View and edit the raw YAML configuration for this service. Changes are synced automatically.
                                </p>
                                <SourceEditor value={yamlContent} onChange={handleYamlChange} />
                            </div>
                        </TabsContent>
                        <TabsContent value="general" className="space-y-4 mt-0">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="name">Service Name</Label>
                                    <Input
                                        id="name"
                                        value={service.name}
                                        onChange={(e) => updateService({ name: e.target.value })}
                                        placeholder="My Service"
                                    />
                                    <p className="text-xs text-muted-foreground">Unique identifier for this service.</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="version">Version</Label>
                                    <Input
                                        id="version"
                                        value={service.version}
                                        onChange={(e) => updateService({ version: e.target.value })}
                                        placeholder="1.0.0"
                                    />
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="priority">Priority</Label>
                                    <Input
                                        id="priority"
                                        type="number"
                                        value={service.priority || 0}
                                        onChange={(e) => updateService({ priority: parseInt(e.target.value) })}
                                    />
                                    <p className="text-xs text-muted-foreground">Lower numbers have higher priority (0 is highest).</p>
                                </div>
                                <div className="flex items-center space-x-2 pt-8">
                                    <Switch
                                        id="disable"
                                        checked={!service.disable}
                                        onCheckedChange={(checked) => updateService({ disable: !checked })}
                                    />
                                    <Label htmlFor="disable">{!service.disable ? 'Enabled' : 'Disabled'}</Label>
                                </div>
                            </div>
                        </TabsContent>

                        <TabsContent value="connection" className="space-y-4 mt-0">
                            <div className="space-y-2">
                                <Label htmlFor="service-type">Service Type</Label>
                                <Select value={getType()} onValueChange={handleTypeChange}>
                                    <SelectTrigger id="service-type">
                                        <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="http">HTTP Service</SelectItem>
                                        <SelectItem value="grpc">gRPC Service</SelectItem>
                                        <SelectItem value="cmd">Command Line (Stdio)</SelectItem>
                                        <SelectItem value="mcp">MCP Proxy</SelectItem>
                                        <SelectItem value="openapi">OpenAPI / Swagger</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            <Separator />

                            {service.httpService && (
                                <div className="space-y-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="http-address">Base URL</Label>
                                        <Input
                                            id="http-address"
                                            value={service.httpService.address}
                                            onChange={(e) => onChange({ ...service, httpService: { ...service.httpService!, address: e.target.value } })}
                                            placeholder="https://api.example.com"
                                        />
                                    </div>
                                </div>
                            )}

                            {service.grpcService && (
                                <div className="space-y-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="grpc-address">Address</Label>
                                        <Input
                                            id="grpc-address"
                                            value={service.grpcService.address}
                                            onChange={(e) => onChange({ ...service, grpcService: { ...service.grpcService!, address: e.target.value } })}
                                            placeholder="localhost:9090"
                                        />
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        <Switch
                                            id="use-reflection"
                                            checked={service.grpcService.useReflection}
                                            onCheckedChange={(checked) => onChange({ ...service, grpcService: { ...service.grpcService!, useReflection: checked } })}
                                        />
                                        <Label htmlFor="use-reflection">Use Reflection</Label>
                                    </div>
                                </div>
                            )}

                            {service.commandLineService && (
                                <div className="space-y-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="command">Command</Label>
                                        <Input
                                            id="command"
                                            value={service.commandLineService.command}
                                            onChange={(e) => onChange({ ...service, commandLineService: { ...service.commandLineService!, command: e.target.value } })}
                                            placeholder="docker"
                                        />
                                    </div>
                                     <div className="space-y-2">
                                        <Label htmlFor="working-directory">Working Directory</Label>
                                        <Input
                                            id="working-directory"
                                            value={service.commandLineService.workingDirectory || ""}
                                            onChange={(e) => onChange({ ...service, commandLineService: { ...service.commandLineService!, workingDirectory: e.target.value } })}
                                            placeholder="/app"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <Label>Environment Variables</Label>
                                        <EnvVarEditor
                                            initialEnv={service.commandLineService.env as any}
                                            onChange={(newEnv) => onChange({ ...service, commandLineService: { ...service.commandLineService!, env: newEnv } })}
                                        />
                                    </div>
                                </div>
                            )}

                             {service.mcpService && (
                                <div className="space-y-4">
                                     <Alert>
                                        <AlertCircle className="h-4 w-4" />
                                        <AlertTitle>MCP Proxy</AlertTitle>
                                        <AlertDescription>
                                            Connects to an existing MCP server via HTTP (SSE) or Stdio.
                                        </AlertDescription>
                                    </Alert>
                                    <div className="flex items-center space-x-2">
                                        <Switch
                                            id="auto-discover"
                                            checked={service.mcpService.toolAutoDiscovery}
                                            onCheckedChange={(checked) => onChange({ ...service, mcpService: { ...service.mcpService!, toolAutoDiscovery: checked } })}
                                        />
                                        <Label htmlFor="auto-discover">Auto-discover Tools</Label>
                                    </div>
                                </div>
                            )}

                            {service.openapiService && (
                                <div className="space-y-4">
                                     <div className="space-y-2">
                                        <Label htmlFor="openapi-address">Base Address</Label>
                                        <Input
                                            id="openapi-address"
                                            value={service.openapiService.address}
                                            onChange={(e) => onChange({ ...service, openapiService: { ...service.openapiService!, address: e.target.value } })}
                                            placeholder="https://api.example.com"
                                        />
                                    </div>
                                     <div className="space-y-2">
                                        <Label htmlFor="openapi-spec">Spec URL</Label>
                                         <Input
                                            id="openapi-spec"
                                            value={service.openapiService.specUrl || ""}
                                            onChange={(e) => onChange({ ...service, openapiService: { ...service.openapiService!, specUrl: e.target.value } })}
                                            placeholder="https://api.example.com/openapi.json"
                                        />
                                    </div>
                                </div>
                            )}
                        </TabsContent>

                        <TabsContent value="policies" className="space-y-4 mt-0">
                            <div className="grid grid-cols-1 gap-6">
                                <PolicyEditor
                                    title="Tool Export Policy"
                                    description="Control which tools are exposed to the AI client."
                                    policy={service.toolExportPolicy}
                                    onChange={(policy) => updateService({ toolExportPolicy: policy })}
                                />
                                <PolicyEditor
                                    title="Prompt Export Policy"
                                    description="Control which prompts are exposed to the AI client."
                                    policy={service.promptExportPolicy}
                                    onChange={(policy) => updateService({ promptExportPolicy: policy })}
                                />
                                <PolicyEditor
                                    title="Resource Export Policy"
                                    description="Control which resources are exposed to the AI client."
                                    policy={service.resourceExportPolicy}
                                    onChange={(policy) => updateService({ resourceExportPolicy: policy })}
                                />
                            </div>
                        </TabsContent>

                        <TabsContent value="auth" className="space-y-4 mt-0">
                            <Card>
                                <CardHeader>
                                    <CardTitle>Upstream Authentication</CardTitle>
                                    <CardDescription>
                                        How MCP Any authenticates with the upstream service.
                                    </CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="auth-type">Authentication Type</Label>
                                        <Select
                                            value={service.upstreamAuth ? (service.upstreamAuth.apiKey ? 'apikey' : service.upstreamAuth.bearerToken ? 'bearer' : service.upstreamAuth.oauth2 ? 'oauth2' : 'none') : 'none'}
                                            onValueChange={(val) => {
                                                if (val === 'none') {
                                                    updateService({ upstreamAuth: undefined });
                                                } else if (val === 'apikey') {
                                                    updateService({ upstreamAuth: { apiKey: { paramName: "", value: { plainText: "", validationRegex: "" }, in: 0, verificationValue: "" } } });
                                                } else if (val === 'bearer') {
                                                    updateService({ upstreamAuth: { bearerToken: { token: { plainText: "", validationRegex: "" } } } });
                                                } else if (val === 'oauth2') {
                                                    updateService({ upstreamAuth: { oauth2: { clientId: { plainText: "", validationRegex: "" }, clientSecret: { plainText: "", validationRegex: "" }, tokenUrl: "", authorizationUrl: "", scopes: "", issuerUrl: "", audience: "" } } });
                                                }
                                            }}
                                        >
                                            <SelectTrigger id="auth-type">
                                                <SelectValue placeholder="No Authentication" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="none">No Authentication</SelectItem>
                                                <SelectItem value="apikey">API Key (Header/Query)</SelectItem>
                                                <SelectItem value="bearer">Bearer Token</SelectItem>
                                                <SelectItem value="oauth2">OAuth 2.0</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    {service.upstreamAuth?.oauth2 && (
                                        <>
                                            <OAuthConfig
                                                auth={service.upstreamAuth.oauth2}
                                                onChange={(newAuth) => updateService({ upstreamAuth: { ...service.upstreamAuth, oauth2: newAuth } })}
                                            />
                                            {/* Show Connect button if we have an ID (saved) */}
                                            {service.name && (
                                                <div className="border-t pt-4 mt-4">
                                                    <OAuthConnect
                                                        serviceId={service.name}
                                                        serviceName={service.name}
                                                        isSaved={!!service.id}
                                                    />
                                                </div>
                                            )}
                                        </>
                                    )}

                                    {service.upstreamAuth?.apiKey && (
                                        <div className="space-y-4 border-l-2 border-primary/20 pl-4">
                                            <div className="space-y-2">
                                                <Label htmlFor="api-key-name">Key Name (Header/Param Name)</Label>
                                                <Input
                                                    id="api-key-name"
                                                    value={service.upstreamAuth.apiKey.paramName}
                                                    onChange={(e) => updateService({ upstreamAuth: { ...service.upstreamAuth, apiKey: { ...service.upstreamAuth!.apiKey!, paramName: e.target.value } } })}
                                                    placeholder="X-API-Key"
                                                />
                                            </div>
                                             <div className="space-y-2">
                                                <Label htmlFor="api-key-value">Value</Label>
                                                <div className="flex gap-2">
                                                    <Input
                                                        id="api-key-value"
                                                        type="password"
                                                        value={service.upstreamAuth.apiKey.value?.plainText || ""}
                                                        onChange={(e) => updateService({ upstreamAuth: { ...service.upstreamAuth, apiKey: { ...service.upstreamAuth!.apiKey!, value: { plainText: e.target.value, validationRegex: "" } } } })}
                                                        placeholder="secret-key-123"
                                                    />
                                                    <SecretPicker
                                                        onSelect={(key) => {
                                                            updateService({ upstreamAuth: { ...service.upstreamAuth, apiKey: { ...service.upstreamAuth!.apiKey!, value: { plainText: `\${${key}}`, validationRegex: "" } } } });
                                                        }}
                                                    >
                                                        <Button variant="outline" size="icon" title="Insert Secret Reference">
                                                            <Key className="h-4 w-4" />
                                                        </Button>
                                                    </SecretPicker>
                                                </div>
                                                <p className="text-[10px] text-muted-foreground">Use <code>{"${SECRET_NAME}"}</code> to reference a stored secret.</p>
                                            </div>
                                             <div className="space-y-2">
                                                <Label htmlFor="api-key-location">Location</Label>
                                                <Select
                                                     value={service.upstreamAuth.apiKey.in?.toString() || "0"}
                                                     onValueChange={(val) => updateService({ upstreamAuth: { ...service.upstreamAuth, apiKey: { ...service.upstreamAuth!.apiKey!, in: parseInt(val) } } })}
                                                >
                                                    <SelectTrigger id="api-key-location">
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value="0">Header</SelectItem>
                                                        <SelectItem value="1">Query Parameter</SelectItem>
                                                        <SelectItem value="2">Cookie</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </div>
                                        </div>
                                    )}

                                    {service.upstreamAuth?.bearerToken && (
                                        <div className="space-y-4 border-l-2 border-primary/20 pl-4">
                                             <div className="space-y-2">
                                                <Label htmlFor="bearer-token">Token</Label>
                                                <div className="flex gap-2">
                                                    <Input
                                                        id="bearer-token"
                                                        type="password"
                                                        value={service.upstreamAuth.bearerToken.token?.plainText || ""}
                                                        onChange={(e) => updateService({ upstreamAuth: { ...service.upstreamAuth, bearerToken: { ...service.upstreamAuth!.bearerToken!, token: { plainText: e.target.value, validationRegex: "" } } } })}
                                                        placeholder="ey..."
                                                    />
                                                    <SecretPicker
                                                        onSelect={(key) => {
                                                            updateService({ upstreamAuth: { ...service.upstreamAuth, bearerToken: { ...service.upstreamAuth!.bearerToken!, token: { plainText: `\${${key}}`, validationRegex: "" } } } });
                                                        }}
                                                    >
                                                        <Button variant="outline" size="icon" title="Insert Secret Reference">
                                                            <Key className="h-4 w-4" />
                                                        </Button>
                                                    </SecretPicker>
                                                </div>
                                                <p className="text-[10px] text-muted-foreground">Use <code>{"${SECRET_NAME}"}</code> to reference a stored secret.</p>
                                            </div>
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </TabsContent>

                        <TabsContent value="advanced" className="space-y-4 mt-0">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="space-y-6">
                                    <ResilienceEditor
                                        config={service.resilience}
                                        onChange={(resilience) => updateService({ resilience })}
                                    />
                                </div>
                                <div className="space-y-6">
                                    <Card>
                                        <CardHeader>
                                            <CardTitle>Export Policy</CardTitle>
                                        </CardHeader>
                                        <CardContent className="space-y-4">
                                            <div className="flex items-center space-x-2">
                                                <Switch
                                                    id="export-auto-discover"
                                                    checked={service.autoDiscoverTool}
                                                    onCheckedChange={(checked) => updateService({ autoDiscoverTool: checked })}
                                                />
                                                <Label htmlFor="export-auto-discover">Auto Discover Tools</Label>
                                            </div>
                                        </CardContent>
                                    </Card>
                                </div>
                            </div>
                        </TabsContent>

                        <TabsContent value="diagnostics" className="space-y-4 mt-0">
                            <ServiceDiagnostics service={service} />
                        </TabsContent>

                        {service.id && (
                            <TabsContent value="inspector" className="space-y-4 mt-0">
                                <ServiceInspector service={service} />
                            </TabsContent>
                        )}
                    </div>
                </Tabs>
            </div>

            <div className="border-t p-4 flex justify-end gap-2 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
                <Button variant="outline" onClick={handleValidate} disabled={validating}>
                    {validating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                    Validate
                </Button>
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={onSave}>Save Changes</Button>
            </div>
        </div>
    );
}
