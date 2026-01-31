/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { Save, RefreshCw, FileText, AlertTriangle, Download, Columns, PanelLeftClose, PanelLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import jsyaml from "js-yaml";
import { stackManager } from "@/lib/stack-manager";

// New Components
import { ServicePalette } from "@/components/stacks/service-palette";
import { StackVisualizer } from "@/components/stacks/stack-visualizer";
import { ConfigEditor } from "./config-editor";

interface StackEditorProps {
    stackId: string;
}

/**
 * StackEditor.
 *
 * @param { stackId - The { stackId.
 */
export function StackEditor({ stackId }: StackEditorProps) {
    const [content, setContent] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [isValid, setIsValid] = useState(true);
    const [validationError, setValidationError] = useState<string | null>(null);
    const [showPalette, setShowPalette] = useState(true);
    const [showVisualizer, setShowVisualizer] = useState(true);

    // Initial load
    useEffect(() => {
        loadConfig();
    }, [stackId]);

    const loadConfig = async () => {
        setIsLoading(true);
        try {
            const stack = await stackManager.getStack(stackId);
            // Transform services array to map for YAML Editor
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const servicesMap: Record<string, any> = {};
            if (stack?.services) {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                stack.services.forEach((s: any) => {
                    servicesMap[s.name] = s;
                });
            }

            const configObj = {
                name: stackId,
                services: servicesMap
            };

            const yaml = jsyaml.dump(configObj);
            setContent(yaml);
        } catch (error) {
            console.error("DEBUG: loadConfig failed:", error);
            toast.error("Failed to load stack configuration");
        } finally {
            setIsLoading(false);
        }
    };

    const handleContentChange = (newVal: string | undefined) => {
        const value = newVal || "";
        setContent(value);
        validateYaml(value);
    };

    const validateYaml = (value: string) => {
        try {
            jsyaml.load(value);
            setIsValid(true);
            setValidationError(null);
        } catch (e: unknown) {
            setIsValid(false);
            if (e instanceof Error) {
                setValidationError(e.message);
            } else {
                setValidationError("Unknown validation error");
            }
        }
    };

    const handleSave = async () => {
        if (!isValid) {
            toast.error("Cannot save invalid configuration");
            return;
        }

        setIsSaving(true);
        try {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const configObj = jsyaml.load(content) as any;

            // Transform services map to array for StackManager
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            let servicesArray: any[] = [];
            if (configObj.services) {
                if (Array.isArray(configObj.services)) {
                     servicesArray = configObj.services;
                } else {
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    Object.entries(configObj.services).forEach(([key, val]: [string, any]) => {
                        servicesArray.push({ ...val, name: key });
                    });
                }
            }

            await stackManager.saveStack(stackId, servicesArray);
            toast.success("Configuration saved successfully");
        } catch (error) {
            console.error(error);
            toast.error("Failed to save configuration");
        } finally {
            setIsSaving(false);
        }
    };

    const handleDownload = () => {
        const blob = new Blob([content], { type: 'text/yaml' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${stackId}-config.yaml`;
        a.click();
    };

    const handleTemplateInsert = (snippet: string) => {
        let newContent = content;

        // Better insertion logic
        const servicesRegex = /^services:\s*$/m;
        const match = newContent.match(servicesRegex);

        if (match) {
            // Found services block.
            if (!newContent.endsWith("\n")) newContent += "\n";
            newContent += snippet;
        } else {
            // No services block found. Append services: block
            if (!newContent.endsWith("\n") && newContent.length > 0) newContent += "\n";
            newContent += "services:\n" + snippet;
        }

        setContent(newContent);
        validateYaml(newContent);
        toast.success("Service template added!");
    };

    return (
        <Card className="flex flex-col h-[650px] border-muted/50 shadow-sm overflow-hidden">
            <CardHeader className="py-2 px-4 border-b flex flex-row items-center justify-between bg-muted/10 shrink-0 h-14">
                <div className="flex items-center gap-2">
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setShowPalette(!showPalette)}>
                        {showPalette ? <PanelLeftClose className="h-4 w-4" /> : <PanelLeft className="h-4 w-4" />}
                    </Button>
                    <FileText className="h-4 w-4 text-muted-foreground ml-2" />
                    <span className="font-medium text-sm">config.yaml</span>
                    {isValid ? (
                         <Badge variant="outline" className="ml-2 bg-green-500/10 text-green-500 border-green-500/20 text-[10px] h-5">
                             Valid YAML
                         </Badge>
                    ) : (
                        <Badge variant="destructive" className="ml-2 text-[10px] h-5">
                             Invalid YAML
                        </Badge>
                    )}
                </div>
                <div className="flex items-center gap-2">
                     <Button variant="ghost" size="sm" onClick={() => setShowVisualizer(!showVisualizer)} title="Toggle Preview">
                        <Columns className="h-4 w-4 mr-1" /> {showVisualizer ? "Hide Preview" : "Show Preview"}
                     </Button>
                     <div className="h-4 w-px bg-border mx-1" />
                     <Button variant="ghost" size="sm" onClick={loadConfig} disabled={isLoading} title="Reset to last saved">
                        <RefreshCw className={`h-3 w-3 mr-1 ${isLoading ? 'animate-spin' : ''}`} /> Reset
                    </Button>
                     <Button variant="ghost" size="sm" onClick={handleDownload} title="Download Config">
                        <Download className="h-3 w-3 mr-1" /> Export
                    </Button>
                    <Button size="sm" onClick={handleSave} disabled={isSaving || !isValid || isLoading}>
                        {isSaving ? <RefreshCw className="h-3 w-3 mr-1 animate-spin" /> : <Save className="h-3 w-3 mr-1" />}
                        Save Changes
                    </Button>
                </div>
            </CardHeader>

            <CardContent className="p-0 flex-1 relative flex overflow-hidden">
                {/* Left Panel: Palette */}
                <div
                    className={`transition-all duration-300 ease-in-out border-r overflow-hidden ${showPalette ? "w-[280px]" : "w-0 border-r-0"}`}
                >
                    <ServicePalette onTemplateSelect={handleTemplateInsert} />
                </div>

                {/* Center Panel: Editor */}
                <div className="flex-1 relative flex flex-col bg-background overflow-hidden min-w-0">
                     <ConfigEditor
                        value={content}
                        onChange={handleContentChange}
                    />
                     {validationError && (
                        <div className="absolute bottom-0 left-0 right-0 py-2 px-4 bg-red-900/90 border-t border-red-500/50 text-red-200 text-xs font-mono z-10 flex items-center">
                            <AlertTriangle className="h-3 w-3 mr-2 text-red-400" />
                            {validationError}
                        </div>
                    )}
                </div>

                {/* Right Panel: Visualizer */}
                <div
                    className={`transition-all duration-300 ease-in-out border-l bg-muted/5 overflow-hidden ${showVisualizer ? "w-[280px]" : "w-0 border-l-0"}`}
                >
                    <StackVisualizer yamlContent={content} />
                </div>
            </CardContent>
        </Card>
    );
}
