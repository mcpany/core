/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { SourceEditor } from "@/components/services/editor/source-editor";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, Save, ArrowLeft, Play } from "lucide-react";
import Link from "next/link";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";

interface StackEditorProps {
    stackId?: string; // If undefined, we are creating a new stack
    initialData?: any;
}

const DEFAULT_YAML = `# MCP Any Stack Configuration
name: my-stack
description: A sample stack with a time service.
version: 1.0.0
services:
  - name: time-service
    command_line_service:
      command: npx -y @modelcontextprotocol/server-time
`;

export function StackEditor({ stackId, initialData }: StackEditorProps) {
    const [name, setName] = useState(initialData?.name || "");
    const [description, setDescription] = useState(initialData?.description || "");
    const [yamlContent, setYamlContent] = useState(DEFAULT_YAML);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();
    const router = useRouter();

    useEffect(() => {
        if (initialData) {
            setName(initialData.name);
            setDescription(initialData.description);
            // Fetch YAML config if editing
            fetchConfig(initialData.name);
        }
    }, [initialData]);

    const fetchConfig = async (id: string) => {
        try {
            // Check if apiClient has getStackConfig, otherwise fallback to getCollection
            // Based on client.ts, getStackConfig is a wrapper around getCollection but returns YAML?
            // Wait, server api_stacks.go returns YAML text/plain for getStackConfig.
            // client.ts getStackConfig calls getCollection which returns JSON.
            // Ah, client.ts getStackConfig implementation:
            // getStackConfig: async (stackId: string) => { return apiClient.getCollection(stackId); },
            // This returns JSON object (Collection).
            // But api_stacks.go returns YAML if I call /stacks/{id}/config.
            // client.ts doesn't seem to have a method pointing to /api/v1/stacks/{id}/config !
            // It points getStackConfig to getCollection (/api/v1/collections/{name}).

            // This is a mismatch. I should use getCollection and convert to YAML on client side for now,
            // OR blindly trust that I can fetch the config endpoint manually if I add a method.
            // But I am not supposed to modify client.ts unless I have to.
            // However, the previous plan step said "If not, add them".
            // I verified they exist, but `getStackConfig` maps to `getCollection` (JSON).

            // If I want the YAML source, I might need to construct it from JSON or fetch the text endpoint.
            // Let's check client.ts again.
            // It has `getStackConfig` calling `getCollection`.
            // So it returns an object.

            // I'll import 'js-yaml' to dump the JSON to YAML for the editor.
            const col = await apiClient.getStackConfig(id);
            // js-yaml dump
            const yaml = (await import("js-yaml")).default;
            setYamlContent(yaml.dump(col));
        } catch (e) {
            console.error("Failed to load stack config", e);
            setError("Failed to load stack configuration.");
        }
    };

    const handleSave = async () => {
        setLoading(true);
        setError(null);
        try {
            // We save by sending the YAML string.
            // client.ts `saveStackConfig` takes (id, config).
            // `config` can be string or object.
            // If string, it parses it as JSON?
            // "const collection = typeof config === 'string' ? JSON.parse(config) : config;"
            // It expects JSON string!
            // But I have YAML string.

            // So I must parse YAML to JSON client-side before sending to `saveStackConfig`.
            const yaml = (await import("js-yaml")).default;
            const parsed = yaml.load(yamlContent) as any;

            // Ensure name matches
            if (name) parsed.name = name;
            if (description) parsed.description = description;

            // If creating, use name as ID
            const targetId = stackId || name;

            await apiClient.saveStackConfig(targetId, parsed);

            toast({
                title: "Stack Deployed",
                description: `Stack ${targetId} has been successfully deployed.`,
            });

            if (!stackId) {
                router.push("/stacks");
            }
        } catch (e: any) {
            console.error(e);
            setError(e.message || "Failed to deploy stack.");
            toast({
                variant: "destructive",
                title: "Deployment Failed",
                description: e.message
            });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-6 h-full flex flex-col">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="icon" asChild>
                        <Link href="/stacks">
                            <ArrowLeft className="h-4 w-4" />
                        </Link>
                    </Button>
                    <div>
                        <h1 className="text-2xl font-bold tracking-tight">
                            {stackId ? `Edit Stack: ${stackId}` : "New Stack"}
                        </h1>
                        <p className="text-muted-foreground text-sm">
                            Define your service collection using YAML.
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <Button onClick={handleSave} disabled={loading}>
                        {loading ? <Play className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        {stackId ? "Redeploy Stack" : "Deploy Stack"}
                    </Button>
                </div>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            <div className="grid gap-6 md:grid-cols-[300px_1fr] flex-1 min-h-0">
                <div className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Details</CardTitle>
                            <CardDescription>Stack metadata.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="name">Name</Label>
                                <Input
                                    id="name"
                                    value={name}
                                    onChange={e => setName(e.target.value)}
                                    disabled={!!stackId}
                                    placeholder="my-stack"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="description">Description</Label>
                                <Textarea
                                    id="description"
                                    value={description}
                                    onChange={e => setDescription(e.target.value)}
                                    placeholder="Purpose of this stack..."
                                />
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/30">
                        <CardHeader>
                            <CardTitle>Documentation</CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-muted-foreground space-y-2">
                            <p>
                                Define services under the <code>services</code> key.
                            </p>
                            <p>
                                Supported service types:
                                <ul className="list-disc list-inside mt-1 ml-1">
                                    <li>command_line_service</li>
                                    <li>http_service</li>
                                    <li>mcp_service</li>
                                </ul>
                            </p>
                        </CardContent>
                    </Card>
                </div>

                <div className="h-full flex flex-col">
                    <Label className="mb-2">Configuration (YAML)</Label>
                    <div className="flex-1 border rounded-md overflow-hidden bg-background">
                        <SourceEditor value={yamlContent} onChange={(v) => setYamlContent(v || "")} />
                    </div>
                </div>
            </div>
        </div>
    );
}
