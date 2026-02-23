/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import nunjucks from "nunjucks";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Copy, AlertCircle, Play, RefreshCw, Variable } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";

interface SmartTemplateEditorProps {
    value: string;
    onChange: (value: string) => void;
    variables?: string[];
    testData?: string;
    placeholder?: string;
    label?: string;
    description?: string;
    className?: string;
}

/**
 * A smart editor for Jinja2 templates with live preview and variable picker.
 */
export function SmartTemplateEditor({
    value,
    onChange,
    variables = [],
    testData: initialTestData = "{}",
    placeholder = "Enter your template here...",
    label = "Template",
    description,
    className
}: SmartTemplateEditorProps) {
    const [testDataJson, setTestDataJson] = useState(initialTestData);
    const [previewResult, setPreviewResult] = useState<string>("");
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    // Configure nunjucks
    const env = useMemo(() => {
        return new nunjucks.Environment(null, { autoescape: false });
    }, []);

    // Effect to render preview when value or testData changes
    useEffect(() => {
        try {
            let data = {};
            try {
                data = JSON.parse(testDataJson);
            } catch {
                // Invalid JSON in test data, just stop rendering but don't error loudly yet
                // The JSON editor part will show its own validation if we had a smart one.
                // For now, let's just use empty object or last valid?
                // Actually, let's set a specific error for JSON
                setError("Invalid JSON in Test Data");
                return;
            }

            const result = env.renderString(value, data);
            setPreviewResult(result);
            setError(null);
        } catch (e: any) {
            setPreviewResult("");
            // Nunjucks error
            setError(`Template Error: ${e.message}`);
        }
    }, [value, testDataJson, env]);

    const copyVariable = (variable: string) => {
        const text = `{{ ${variable} }}`;
        navigator.clipboard.writeText(text);
        toast({
            title: "Copied",
            description: `${text} copied to clipboard`,
            duration: 1000,
        });
    };

    return (
        <div className={`grid grid-cols-1 lg:grid-cols-2 gap-4 h-full ${className}`}>
            <div className="flex flex-col space-y-4 h-full">
                <div className="space-y-2 flex-1 flex flex-col">
                    <div className="flex justify-between items-center">
                        <Label>{label}</Label>
                        {variables.length > 0 && (
                            <span className="text-xs text-muted-foreground">{variables.length} variables available</span>
                        )}
                    </div>

                    {description && (
                         <p className="text-xs text-muted-foreground">{description}</p>
                    )}

                    {variables.length > 0 && (
                        <ScrollArea className="h-12 w-full border rounded-md bg-muted/20 p-2">
                            <div className="flex flex-wrap gap-2">
                                {variables.map((v) => (
                                    <Badge
                                        key={v}
                                        variant="outline"
                                        className="cursor-pointer hover:bg-primary/10 transition-colors flex items-center gap-1"
                                        onClick={() => copyVariable(v)}
                                        title="Click to copy"
                                    >
                                        <Variable className="h-3 w-3 opacity-50" />
                                        {v}
                                    </Badge>
                                ))}
                            </div>
                        </ScrollArea>
                    )}

                    <Textarea
                        value={value}
                        onChange={(e) => onChange(e.target.value)}
                        placeholder={placeholder}
                        className="font-mono text-sm flex-1 min-h-[200px] resize-none"
                    />
                </div>
            </div>

            <div className="flex flex-col space-y-4 h-full">
                <div className="space-y-2 flex-1 flex flex-col">
                    <Label>Test Data (JSON)</Label>
                    <Textarea
                        value={testDataJson}
                        onChange={(e) => setTestDataJson(e.target.value)}
                        className="font-mono text-xs h-[150px] resize-none"
                        placeholder="{}"
                    />
                </div>

                <div className="space-y-2 flex-1 flex flex-col min-h-[150px]">
                    <div className="flex justify-between items-center">
                         <Label>Live Preview</Label>
                         {error ? (
                             <Badge variant="destructive" className="h-5 text-[10px]">Error</Badge>
                         ) : (
                             <Badge variant="outline" className="h-5 text-[10px] text-green-600 border-green-200 bg-green-50">Valid</Badge>
                         )}
                    </div>

                    <Card className={`flex-1 overflow-hidden ${error ? "border-destructive/50 bg-destructive/5" : "bg-muted/10"}`}>
                        <CardContent className="p-0 h-full flex flex-col">
                            {error ? (
                                <div className="p-4 text-xs text-destructive font-mono whitespace-pre-wrap break-words overflow-auto">
                                    <div className="flex items-center gap-2 mb-2 font-semibold">
                                        <AlertCircle className="h-4 w-4" />
                                        Rendering Failed
                                    </div>
                                    {error}
                                </div>
                            ) : (
                                <div className="p-4 h-full overflow-auto">
                                    <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                                        {previewResult || <span className="text-muted-foreground italic">Result will appear here...</span>}
                                    </pre>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
