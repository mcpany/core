/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { InputTransformer, HttpParameterMapping } from "@proto/config/v1/call";
import { Card, CardContent } from "@/components/ui/card";
import { Info } from "lucide-react";
import { SmartTemplateEditor } from "./smart-template-editor";

interface InputTransformerEditorProps {
    transformer?: InputTransformer;
    parameters?: HttpParameterMapping[];
    onChange: (transformer: InputTransformer) => void;
}

/**
 * Editor for InputTransformer configuration.
 */
export function InputTransformerEditor({ transformer, parameters = [], onChange }: InputTransformerEditorProps) {
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

    const variableNames = useMemo(() => {
        return parameters.map(p => p.schema?.name || "").filter(n => n !== "");
    }, [parameters]);

    return (
        <div className="space-y-4">
            <SmartTemplateEditor
                label="Request Body Template"
                value={template}
                onChange={handleTemplateChange}
                variables={variableNames}
                placeholder='{"message": "{{ input_message }}"}'
                helperText={
                    <span className="flex items-center gap-1">
                        <Info className="h-3 w-3 inline" /> Use Jinja2 syntax to construct the request body. Variables from input are available.
                    </span>
                }
            />

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
