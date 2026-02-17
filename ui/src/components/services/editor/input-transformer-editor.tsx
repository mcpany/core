/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { InputTransformer } from "@proto/config/v1/call";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Info } from "lucide-react";

interface InputTransformerEditorProps {
    transformer?: InputTransformer;
    onChange: (transformer: InputTransformer) => void;
}

/**
 * Editor for InputTransformer configuration.
 *
*
 * Summary: Editor for InputTransformer configuration.
 */
export function InputTransformerEditor({ transformer, onChange }: InputTransformerEditorProps) {
    const [template, setTemplate] = useState(transformer?.template || "");

    useEffect(() => {
        setTemplate(transformer?.template || "");
    }, [transformer]);

    const handleTemplateChange = (value: string) => {
        setTemplate(value);
        onChange({
            ...transformer,
            template: value,
        });
    };

    return (
        <div className="space-y-4">
            <div className="space-y-2">
                <Label htmlFor="input-template">Request Body Template</Label>
                <div className="text-xs text-muted-foreground flex items-center gap-2 mb-2">
                    <Info className="h-4 w-4" />
                    Use Jinja2 syntax to construct the request body. Variables from input are available.
                </div>
                <Textarea
                    id="input-template"
                    value={template}
                    onChange={(e) => handleTemplateChange(e.target.value)}
                    placeholder='{"message": "{{ input_message }}"}'
                    className="font-mono text-sm min-h-[200px]"
                />
            </div>
             {transformer?.webhook && (
                <Card className="bg-muted/50">
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Info className="h-4 w-4" />
                            A webhook is configured for input transformation. The template above will be ignored if the webhook is active.
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
