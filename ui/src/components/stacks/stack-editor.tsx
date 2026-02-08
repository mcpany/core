/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Editor from "@monaco-editor/react";
import yaml from "js-yaml";
import { Save, X, AlertCircle, CheckCircle2, Loader2, FileCode } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { useToast } from "@/hooks/use-toast";

interface StackEditorProps {
    initialStack?: ServiceCollection;
    isNew?: boolean;
}

const DEFAULT_TEMPLATE = `name: my-new-stack
description: A collection of MCP services
version: 1.0.0
author: user
services:
  - name: weather-service
    disable: false
    commandLineService:
      command: npx -y @modelcontextprotocol/server-weather
`;

export function StackEditor({ initialStack, isNew = false }: StackEditorProps) {
    const router = useRouter();
    const { toast } = useToast();
    const [content, setContent] = useState("");
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (initialStack) {
            try {
                setContent(yaml.dump(initialStack));
            } catch (e) {
                console.error("Failed to dump stack to YAML", e);
                setError("Failed to load stack configuration.");
            }
        } else {
            setContent(DEFAULT_TEMPLATE);
        }
    }, [initialStack]);

    const handleSave = async () => {
        setSaving(true);
        setError(null);
        try {
            const parsed = yaml.load(content) as ServiceCollection;

            // Basic Validation
            if (!parsed || typeof parsed !== 'object') {
                throw new Error("Invalid YAML: Root must be an object.");
            }
            if (!parsed.name) {
                throw new Error("Validation Error: 'name' field is required.");
            }
            if (!parsed.services || !Array.isArray(parsed.services)) {
                throw new Error("Validation Error: 'services' must be a list.");
            }

            await apiClient.saveCollection(parsed);

            toast({
                title: isNew ? "Stack Created" : "Stack Updated",
                description: `Successfully saved stack "${parsed.name}".`,
                action: <CheckCircle2 className="h-5 w-5 text-green-500" />
            });

            router.push(`/stacks/${parsed.name}`);
        } catch (e: any) {
            console.error("Failed to save stack", e);
            setError(e.message || "Failed to save stack.");
            toast({
                variant: "destructive",
                title: "Error",
                description: e.message || "Failed to save stack."
            });
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-full gap-4">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">{isNew ? "Create Stack" : "Edit Stack"}</h2>
                    <p className="text-muted-foreground">
                        {isNew ? "Define a new service collection using YAML." : "Modify the service collection configuration."}
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={() => router.back()} disabled={saving}>
                        <X className="mr-2 h-4 w-4" /> Cancel
                    </Button>
                    <Button onClick={handleSave} disabled={saving}>
                        {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        {isNew ? "Create Stack" : "Save Changes"}
                    </Button>
                </div>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Validation Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            <Card className="flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm">
                <CardHeader className="py-3 px-4 border-b bg-muted/20 flex flex-row items-center justify-between">
                    <div className="flex items-center gap-2">
                        <FileCode className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-medium">stack.yaml</span>
                    </div>
                    <div className="text-xs text-muted-foreground">
                        YAML
                    </div>
                </CardHeader>
                <CardContent className="p-0 flex-1 relative">
                    <Editor
                        height="100%"
                        defaultLanguage="yaml"
                        value={content}
                        onChange={(val) => setContent(val || "")}
                        theme="vs-dark"
                        options={{
                            minimap: { enabled: false },
                            scrollBeyondLastLine: false,
                            fontSize: 14,
                            fontFamily: "var(--font-mono)",
                            padding: { top: 16, bottom: 16 },
                        }}
                        loading={
                            <div className="flex items-center justify-center h-full text-muted-foreground">
                                <Loader2 className="h-6 w-6 animate-spin mr-2" /> Loading Editor...
                            </div>
                        }
                    />
                </CardContent>
            </Card>
        </div>
    );
}
