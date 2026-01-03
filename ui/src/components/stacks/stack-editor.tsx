/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { Save, RefreshCw, FileText, CheckCircle, AlertTriangle, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import jsyaml from "js-yaml";
import { apiClient } from "@/lib/client";

interface StackEditorProps {
    stackId: string;
}

export function StackEditor({ stackId }: StackEditorProps) {
    const [content, setContent] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [isValid, setIsValid] = useState(true);
    const [validationError, setValidationError] = useState<string | null>(null);
    const lineNumbersRef = React.useRef<HTMLDivElement>(null);
    const textareaRef = React.useRef<HTMLTextAreaElement>(null);

    // Mock initial load
    useEffect(() => {
        loadConfig();
    }, [stackId]);

    const handleScroll = (e: React.UIEvent<HTMLTextAreaElement>) => {
        if (lineNumbersRef.current) {
            lineNumbersRef.current.scrollTop = e.currentTarget.scrollTop;
        }
    };

    const loadConfig = async () => {
        setIsLoading(true);
        try {
            // Attempt to fetch from API, if fail use mock
            try {
                const config = await apiClient.getStackConfig(stackId);
                setContent(config);
            } catch (e) {
                // Fallback / Mock
                const mockConfig = `# Stack Configuration for ${stackId}
version: "1.0"
services:
  weather-service:
    image: mcp/weather:latest
    environment:
      - API_KEY=\${WEATHER_API_KEY}
  local-files:
    command: npx -y @modelcontextprotocol/server-filesystem /Users/me/Documents
`;
                setContent(mockConfig);
            }
        } catch (error) {
            toast.error("Failed to load stack configuration");
        } finally {
            setIsLoading(false);
        }
    };

    const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const newVal = e.target.value;
        setContent(newVal);
        validateYaml(newVal);
    };

    const validateYaml = (value: string) => {
        try {
            jsyaml.load(value);
            setIsValid(true);
            setValidationError(null);
        } catch (e: any) {
            setIsValid(false);
            setValidationError(e.message);
        }
    };

    const handleSave = async () => {
        if (!isValid) {
            toast.error("Cannot save invalid configuration");
            return;
        }

        setIsSaving(true);
        try {
            await apiClient.saveStackConfig(stackId, content);
            toast.success("Configuration saved successfully");
        } catch (error) {
            // If API not implemented yet, simulate success for UI demo
            console.warn("API save failed, simulating success for demo", error);
             setTimeout(() => {
                toast.success("Configuration saved (Simulated)");
            }, 500);
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

    return (
        <Card className="flex flex-col h-[600px] border-muted/50 shadow-sm">
            <CardHeader className="py-3 px-4 border-b flex flex-row items-center justify-between bg-muted/10">
                <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4 text-muted-foreground" />
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
            <CardContent className="p-0 flex-1 relative bg-[#1e1e1e] text-[#d4d4d4] font-mono text-sm overflow-hidden">
                <div className="absolute inset-0 flex">
                     {/* Line Numbers - Simple implementation */}
                    <div
                        ref={lineNumbersRef}
                        className="w-10 bg-[#1e1e1e] border-r border-[#333] text-right pr-2 pt-4 text-xs text-[#858585] select-none opacity-50 overflow-hidden"
                    >
                        {content.split('\n').map((_, i) => (
                            <div key={i} className="leading-6">{i + 1}</div>
                        ))}
                    </div>
                    {/* Editor Area */}
                    <Textarea
                        ref={textareaRef}
                        value={content}
                        onChange={handleContentChange}
                        onScroll={handleScroll}
                        className="flex-1 h-full w-full resize-none border-0 rounded-none bg-transparent text-inherit p-4 leading-6 focus-visible:ring-0 focus-visible:ring-offset-0 font-mono"
                        spellCheck={false}
                    />
                </div>
            </CardContent>
             {validationError && (
                <CardFooter className="py-2 px-4 bg-red-900/10 border-t border-red-900/20 text-red-500 text-xs font-mono">
                    <AlertTriangle className="h-3 w-3 mr-2" />
                    {validationError}
                </CardFooter>
            )}
        </Card>
    );
}
