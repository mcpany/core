/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import nunjucks from "nunjucks";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Copy, AlertCircle, Play, CheckCircle2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface SmartTemplateEditorProps {
    value: string;
    onChange: (value: string) => void;
    variables?: string[];
    placeholder?: string;
    label?: string;
    helperText?: React.ReactNode;
}

/**
 * SmartTemplateEditor provides a rich editing experience for Jinja2 templates.
 * It includes variable auto-completion (via copy-paste) and live preview capabilities.
 */
export function SmartTemplateEditor({
    value,
    onChange,
    variables = [],
    placeholder,
    label = "Template",
    helperText
}: SmartTemplateEditorProps) {
    const [testData, setTestData] = useState("{}");
    const [previewResult, setPreviewResult] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState("edit");
    const { toast } = useToast();

    // Configure nunjucks
    useEffect(() => {
        nunjucks.configure({ autoescape: true });
    }, []);

    // Initialize test data with variables if empty
    useEffect(() => {
        if (testData === "{}" && variables.length > 0) {
            const initialData: Record<string, string> = {};
            variables.forEach(v => {
                initialData[v] = `value_of_${v}`;
            });
            setTestData(JSON.stringify(initialData, null, 2));
        }
    }, [variables.length]); // Run only when variables populate initially

    const renderPreview = useMemo(() => {
        try {
            const data = JSON.parse(testData);
            const result = nunjucks.renderString(value, data);
            setError(null);
            return result;
        } catch (e: any) {
            // Don't set error state immediately on every keystroke to avoid flickering UI
            // but return error string for preview
            return `Error: ${e.message}`;
        }
    }, [value, testData]);

    const copyVariable = (variable: string) => {
        const textToCopy = `{{ ${variable} }}`;
        navigator.clipboard.writeText(textToCopy);
        toast({
            title: "Copied to clipboard",
            description: `${textToCopy} copied. Paste it into the template.`,
            duration: 2000,
        });
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <Label htmlFor="smart-template-editor">{label}</Label>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Left Column: Editor */}
                <div className="space-y-2">
                    {variables.length > 0 && (
                        <div className="flex flex-wrap gap-2 mb-2 p-2 bg-muted/30 rounded-md border">
                            <span className="text-xs text-muted-foreground w-full flex items-center gap-1">
                                Available Variables (Click to copy):
                            </span>
                            {variables.map(v => (
                                <Badge
                                    key={v}
                                    variant="outline"
                                    className="cursor-pointer hover:bg-primary/10 hover:text-primary transition-colors flex items-center gap-1"
                                    onClick={() => copyVariable(v)}
                                    title={`Copy {{ ${v} }}`}
                                >
                                    {v}
                                    <Copy className="h-3 w-3 opacity-50" />
                                </Badge>
                            ))}
                        </div>
                    )}

                    <Textarea
                        id="smart-template-editor"
                        value={value}
                        onChange={(e) => onChange(e.target.value)}
                        placeholder={placeholder || "Enter Jinja2 template..."}
                        className="font-mono text-sm min-h-[300px]"
                    />

                    {helperText && (
                        <div className="text-xs text-muted-foreground mt-1">
                            {helperText}
                        </div>
                    )}
                </div>

                {/* Right Column: Preview & Test Data */}
                <div className="space-y-4 flex flex-col h-full">
                    <Card className="flex-1 flex flex-col overflow-hidden">
                        <div className="border-b bg-muted/40 p-2">
                            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                                <TabsList className="w-full grid grid-cols-2">
                                    <TabsTrigger value="edit">Test Data (JSON)</TabsTrigger>
                                    <TabsTrigger value="preview">Live Preview</TabsTrigger>
                                </TabsList>
                            </Tabs>
                        </div>

                        <CardContent className="flex-1 p-0 relative min-h-[300px]">
                            {activeTab === "edit" ? (
                                <div className="h-full flex flex-col">
                                    <Textarea
                                        value={testData}
                                        onChange={(e) => setTestData(e.target.value)}
                                        className="flex-1 border-0 rounded-none resize-none font-mono text-xs focus-visible:ring-0 p-4"
                                        placeholder='{"variable": "value"}'
                                    />
                                    {(() => {
                                        try {
                                            JSON.parse(testData);
                                            return (
                                                <div className="bg-green-500/10 text-green-600 text-xs py-1 px-4 flex items-center gap-2 border-t border-green-500/20">
                                                    <CheckCircle2 className="h-3 w-3" /> Valid JSON
                                                </div>
                                            );
                                        } catch (e: any) {
                                            return (
                                                <div className="bg-destructive/10 text-destructive text-xs py-1 px-4 flex items-center gap-2 border-t border-destructive/20">
                                                    <AlertCircle className="h-3 w-3" /> Invalid JSON
                                                </div>
                                            );
                                        }
                                    })()}
                                </div>
                            ) : (
                                <div className="h-full flex flex-col">
                                    <div className="flex-1 p-4 overflow-auto bg-muted/10 font-mono text-sm whitespace-pre-wrap break-all">
                                        {renderPreview}
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
