/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
// @ts-ignore
import nunjucks from "nunjucks/browser/nunjucks";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { AlertCircle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface SmartTemplateEditorProps {
    value: string;
    onChange: (value: string) => void;
    variables?: string[];
    initialTestData?: string;
    placeholder?: string;
    label?: string;
    description?: string;
}

/**
 * SmartTemplateEditor component.
 * Provides a template editor with live preview using Nunjucks and optional test data.
 * @param props - The component props.
 * @returns The rendered component.
 */
export function SmartTemplateEditor({
    value,
    onChange,
    variables = [],
    initialTestData = "{}",
    placeholder,
    label = "Template",
    description
}: SmartTemplateEditorProps) {
    const [testData, setTestData] = useState(initialTestData);
    const [preview, setPreview] = useState("");
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    // Update test data if initial changes significantly (optional, but good for resetting)
    // Actually, keeping local state is better so user edits aren't lost if props change unexpectedly.
    // However, if we switch tools, we might want to reset.
    // Let's rely on key changes in parent or just initial mount.

    useEffect(() => {
        let active = true;
        try {
            const data = JSON.parse(testData);
            if (!active) return;

            setError(null);
            try {
                // Configure nunjucks to be more like Jinja2 (autoescape false for this use case usually?)
                // Actually, for API requests, we might want autoescape off or on depending on content type.
                // Defaulting to autoescape: true is safer for HTML, but for JSON bodies it might escape quotes.
                // Let's use autoescape: false as we are generating raw text/json usually.
                const env = new nunjucks.Environment(null, { autoescape: false });
                const res = env.renderString(value, data);
                if (active) setPreview(res);
            } catch (e: any) {
                // Template error
                if (active) {
                    setPreview("");
                    setError(`Template Error: ${e.message}`);
                }
            }
        } catch (e: any) {
            // JSON error
            if (active) setError(`Invalid JSON Test Data: ${e.message}`);
        }
        return () => { active = false; };
    }, [value, testData]);

    const copyVariable = (variable: string) => {
        const text = `{{ ${variable} }}`;
        navigator.clipboard.writeText(text);
        toast({
            title: "Copied",
            description: `${text} copied to clipboard.`
        });
    };

    return (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 h-full min-h-[400px]">
            <div className="flex flex-col gap-4 h-full">
                <div className="space-y-2 flex-1 flex flex-col h-full">
                    <div className="flex items-center justify-between">
                        <Label htmlFor="template-editor">{label}</Label>
                        {variables.length > 0 && (
                            <div className="flex gap-1 flex-wrap justify-end max-w-[70%]">
                                {variables.map(v => (
                                    <Badge
                                        key={v}
                                        variant="secondary"
                                        className="cursor-pointer hover:bg-secondary/80 active:scale-95 transition-all text-[10px]"
                                        onClick={() => copyVariable(v)}
                                        title="Click to copy"
                                    >
                                        {v}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                    {description && <p className="text-xs text-muted-foreground">{description}</p>}
                    <Textarea
                        id="template-editor"
                        value={value}
                        onChange={(e) => onChange(e.target.value)}
                        placeholder={placeholder}
                        className="font-mono text-sm flex-1 min-h-[200px] resize-none"
                    />
                </div>
            </div>

            <div className="flex flex-col gap-4 h-full">
                 <div className="space-y-2 h-1/2 flex flex-col">
                    <Label htmlFor="test-data-editor">Test Data (JSON)</Label>
                    <Textarea
                        id="test-data-editor"
                        value={testData}
                        onChange={(e) => setTestData(e.target.value)}
                        className="font-mono text-xs flex-1 resize-none bg-muted/30"
                        placeholder="{}"
                    />
                </div>

                <div className="space-y-2 h-1/2 flex flex-col">
                    <Label>Live Preview</Label>
                    <Card className="flex-1 bg-muted/50 overflow-hidden flex flex-col">
                        <CardContent className="p-3 flex-1 overflow-auto">
                            {error ? (
                                <Alert variant="destructive" className="py-2">
                                    <AlertCircle className="h-4 w-4" />
                                    <AlertTitle>Error</AlertTitle>
                                    <AlertDescription className="text-xs">{error}</AlertDescription>
                                </Alert>
                            ) : (
                                <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                                    {preview}
                                </pre>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
