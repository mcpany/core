/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import yaml from "js-yaml";
import { Loader2, Save, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/client";

interface StackEditorProps {
    initialContent?: string;
    stackId?: string;
    isNew?: boolean;
}

/**
 * StackEditor component allows users to edit stack configurations using a YAML editor.
 * @param props The component props.
 * @param props.initialContent The initial YAML content.
 * @param props.stackId The ID of the stack.
 * @param props.isNew Whether this is a new stack.
 * @returns The rendered component.
 */
const DEFAULT_TEMPLATE = `name: my-new-stack
description: A collection of services
services:
  - name: weather
    mcp_service:
      http_connection:
        http_address: http://example.com
`;

export function StackEditor({ initialContent, stackId, isNew }: StackEditorProps) {
    const [content, setContent] = useState(initialContent || DEFAULT_TEMPLATE);
    const [saving, setSaving] = useState(false);
    const { toast } = useToast();
    const router = useRouter();

    const handleSave = async () => {
        setSaving(true);
        try {
            // Validate YAML locally first
            let parsed: any;
            try {
                parsed = yaml.load(content);
            } catch (e) {
                throw new Error("Invalid YAML structure: " + String(e));
            }

            if (!parsed || typeof parsed !== 'object') {
                throw new Error("Invalid YAML structure");
            }

            const idToSave = parsed.name || stackId;
            if (!idToSave) {
                 throw new Error("Stack name is required in YAML");
            }

            if (isNew) {
                // If creating, we use saveStackYaml but potentially need to handle ID logic
                // The client method takes stackId, which maps to URL path
                // For new stack, we use the name from YAML as ID
                await apiClient.saveStackYaml(idToSave, content);
            } else {
                // For existing stack, we use the original ID unless we want to support rename (which changes ID)
                // If ID changes, we create a new one and maybe delete old?
                // For now, assume ID is immutable or we just save to the ID in URL
                if (stackId && idToSave !== stackId) {
                     if (!confirm("Changing the stack name will create a new stack. Continue?")) {
                         setSaving(false);
                         return;
                     }
                }
                await apiClient.saveStackYaml(idToSave, content);
            }

            toast({
                title: "Stack Saved",
                description: `Stack configuration has been saved successfully.`
            });

            router.push("/stacks");
            router.refresh(); // Refresh list
        } catch (e: any) {
            console.error("Save failed", e);
            toast({
                variant: "destructive",
                title: "Save Failed",
                description: e.message || "Could not save stack configuration."
            });
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-full gap-4">
            <div className="flex justify-between items-center bg-background/50 backdrop-blur-sm p-4 border rounded-lg shadow-sm">
                <div>
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        {isNew ? <><span className="text-primary">+</span> Create Stack</> : <>Edit Stack: <span className="font-mono text-muted-foreground">{stackId}</span></>}
                    </h2>
                    <p className="text-sm text-muted-foreground">Define your service collection using YAML.</p>
                </div>
                <div className="flex gap-2">
                     <Button variant="ghost" onClick={() => router.back()}>
                        <X className="mr-2 h-4 w-4" /> Cancel
                    </Button>
                    <Button onClick={handleSave} disabled={saving}>
                        {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        {isNew ? "Create Stack" : "Save Changes"}
                    </Button>
                </div>
            </div>

            <div className="flex-1 border rounded-lg overflow-hidden relative shadow-inner bg-[#1e1e1e]">
                <Editor
                    height="100%"
                    defaultLanguage="yaml"
                    value={content}
                    onChange={(value) => setContent(value || "")}
                    theme="vs-dark"
                    options={{
                        minimap: { enabled: false },
                        fontSize: 14,
                        scrollBeyondLastLine: false,
                        automaticLayout: true,
                        wordWrap: "on",
                    }}
                />
            </div>
             <div className="text-xs text-muted-foreground px-2 flex justify-between">
                 <span>
                     Required fields: <code>name</code>, <code>services</code>.
                 </span>
                 <span>
                     Use <code>Ctrl+S</code> / <code>Cmd+S</code> to save.
                 </span>
             </div>
        </div>
    );
}
