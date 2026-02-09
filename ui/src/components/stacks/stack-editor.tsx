/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { Loader2, Save, Play, AlertCircle, CheckCircle2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface StackEditorProps {
    stackId: string;
    onClose: () => void;
    onSaved: () => void;
}

export function StackEditor({ stackId, onClose, onSaved }: StackEditorProps) {
    const [yaml, setYaml] = useState("");
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [applying, setApplying] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    // If stackId is "new", we start with a template
    const isNew = stackId === "new";

    useEffect(() => {
        if (isNew) {
            setYaml(`# New Stack Configuration
name: my-stack
description: A collection of services
services:
  - name: weather-service
    httpService:
      address: https://wttr.in
      tools:
        - name: get_weather
          description: Get weather
          call_id: weather_call
      calls:
        weather_call:
          endpoint_path: "/?format=j1"
          method: HTTP_METHOD_GET
`);
            return;
        }

        async function load() {
            setLoading(true);
            try {
                const config = await apiClient.getStackConfig(stackId);
                setYaml(config);
            } catch (e: any) {
                setError(e.message);
                toast({
                    title: "Failed to load stack",
                    description: e.message,
                    variant: "destructive"
                });
            } finally {
                setLoading(false);
            }
        }
        load();
    }, [stackId, isNew, toast]);

    const handleSave = async () => {
        setSaving(true);
        setError(null);
        try {
            // If new, we need to extract name from YAML to determine ID
            // For simplicity, we assume ID matches Name.
            // Parse YAML locally to get name? Or let backend handle it?
            // saveStackConfig takes an ID.
            // If isNew, we might need a prompt or regex to get the name.
            let targetId = stackId;
            if (isNew) {
                const match = yaml.match(/^name:\s*(.+)$/m);
                if (match) {
                    targetId = match[1].trim();
                } else {
                    throw new Error("YAML must contain a 'name' field.");
                }
            }

            await apiClient.saveStackConfig(targetId, yaml);
            toast({
                title: "Stack Saved",
                description: `Configuration for ${targetId} has been saved.`,
                action: <CheckCircle2 className="h-5 w-5 text-green-500" />
            });
            onSaved();
        } catch (e: any) {
            setError(e.message);
            toast({
                title: "Save Failed",
                description: e.message,
                variant: "destructive"
            });
        } finally {
            setSaving(false);
        }
    };

    const handleApply = async () => {
        setApplying(true);
        try {
            // Must save first if dirty? Or Apply takes current config?
            // applyStack endpoint uses the SAVED config. So we must save first.
            await handleSave();

            // Re-derive ID if it was new (handleSave might have updated parent, but here we need local logic)
            let targetId = stackId;
            if (isNew) {
                 const match = yaml.match(/^name:\s*(.+)$/m);
                 if (match) targetId = match[1].trim();
            }

            await apiClient.applyStack(targetId);
            toast({
                title: "Stack Applied",
                description: "All services in the stack have been registered/updated.",
            });
        } catch (e: any) {
             toast({
                title: "Apply Failed",
                description: e.message,
                variant: "destructive"
            });
        } finally {
            setApplying(false);
        }
    }

    if (loading) {
        return <div className="flex items-center justify-center h-64"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>;
    }

    return (
        <div className="flex flex-col h-full space-y-4">
            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}
            <div className="flex-1 min-h-[400px] border rounded-md overflow-hidden relative">
                <Textarea
                    value={yaml}
                    onChange={(e) => setYaml(e.target.value)}
                    className="w-full h-full font-mono text-sm p-4 resize-none border-0 focus-visible:ring-0 bg-muted/20"
                    placeholder="Enter stack configuration in YAML..."
                    spellCheck={false}
                />
            </div>
            <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={onClose}>Close</Button>
                <Button variant="secondary" onClick={handleSave} disabled={saving}>
                    {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    Save
                </Button>
                <Button onClick={handleApply} disabled={applying || saving}>
                    {applying ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                    Save & Apply
                </Button>
            </div>
        </div>
    );
}
