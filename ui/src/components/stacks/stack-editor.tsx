/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { Save, X, AlertTriangle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import yaml from "js-yaml";

interface StackEditorProps {
    initialContent: string;
    onSave: (content: string) => Promise<void>;
    onCancel: () => void;
}

export function StackEditor({ initialContent, onSave, onCancel }: StackEditorProps) {
    const [content, setContent] = useState(initialContent);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    const handleSave = async () => {
        setSaving(true);
        setError(null);
        try {
            // Validate YAML
            try {
                yaml.load(content);
            } catch (e: any) {
                throw new Error(`Invalid YAML: ${e.message}`);
            }

            await onSave(content);
        } catch (e: any) {
            console.error("Failed to save stack", e);
            setError(e.message);
            toast({
                title: "Error",
                description: e.message || "Failed to save stack configuration.",
                variant: "destructive"
            });
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-full bg-background border rounded-lg overflow-hidden shadow-sm">
            <div className="flex items-center justify-between p-4 border-b bg-muted/20">
                <div className="flex items-center gap-2">
                    <h2 className="text-lg font-semibold">Stack Configuration</h2>
                    {error && (
                         <div className="text-xs text-destructive flex items-center gap-1 bg-destructive/10 px-2 py-1 rounded">
                             <AlertTriangle className="h-3 w-3" />
                             {error}
                         </div>
                    )}
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="ghost" onClick={onCancel} disabled={saving}>
                        <X className="mr-2 h-4 w-4" /> Cancel
                    </Button>
                    <Button onClick={handleSave} disabled={saving}>
                        <Save className="mr-2 h-4 w-4" /> {saving ? "Saving..." : "Save Changes"}
                    </Button>
                </div>
            </div>
            <div className="flex-1 min-h-0 relative">
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
                        automaticLayout: true
                    }}
                />
            </div>
        </div>
    );
}
