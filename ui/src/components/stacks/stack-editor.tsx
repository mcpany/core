/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useCallback } from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { Loader2, Save, FileJson, AlertCircle } from "lucide-react";
import yaml from "js-yaml";
import { useToast } from "@/hooks/use-toast";

interface StackEditorProps {
    initialYaml?: string;
    onSave: (yamlContent: string, parsedJson: any) => Promise<void>;
    onCancel: () => void;
    isSaving?: boolean;
}

const DEFAULT_YAML = `# MCP Any Stack Configuration
name: my-stack
description: A collection of services
version: 1.0.0
services:
  - name: example-service
    httpService:
      address: https://api.example.com
    disable: false
`;

/**
 * StackEditor component.
 * @returns The rendered component.
 */
export function StackEditor({ initialYaml = DEFAULT_YAML, onSave, onCancel, isSaving }: StackEditorProps) {
    const [value, setValue] = useState(initialYaml);
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    const handleEditorChange = (newValue: string | undefined) => {
        setValue(newValue || "");
        if (error) setError(null);
    };

    const handleSave = useCallback(async () => {
        try {
            const parsed = yaml.load(value);
            if (typeof parsed !== 'object' || parsed === null) {
                throw new Error("YAML must evaluate to an object");
            }
            await onSave(value, parsed);
        } catch (e: any) {
            console.error("YAML Parse Error", e);
            setError(e.message || "Invalid YAML");
            toast({
                variant: "destructive",
                title: "Invalid Configuration",
                description: e.message || "Please correct the YAML syntax errors."
            });
        }
    }, [value, onSave, toast]);

    return (
        <div className="flex flex-col h-full space-y-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <FileJson className="h-5 w-5 text-muted-foreground" />
                    <h3 className="font-medium">Stack Configuration (YAML)</h3>
                </div>
                {error && (
                    <div className="flex items-center gap-2 text-destructive text-sm bg-destructive/10 px-3 py-1 rounded animate-in fade-in">
                        <AlertCircle className="h-4 w-4" />
                        <span className="truncate max-w-[300px]" title={error}>{error}</span>
                    </div>
                )}
            </div>

            <div className="flex-1 border rounded-md overflow-hidden min-h-[400px]">
                <Editor
                    height="100%"
                    defaultLanguage="yaml"
                    value={value}
                    onChange={handleEditorChange}
                    theme="vs-dark"
                    options={{
                        minimap: { enabled: false },
                        scrollBeyondLastLine: false,
                        fontSize: 13,
                        automaticLayout: true,
                    }}
                />
            </div>

            <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={onCancel} disabled={isSaving}>
                    Cancel
                </Button>
                <Button onClick={handleSave} disabled={isSaving}>
                    {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    {isSaving ? "Deploying..." : "Deploy Stack"}
                </Button>
            </div>
        </div>
    );
}
