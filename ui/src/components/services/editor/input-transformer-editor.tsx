/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { InputTransformer, HttpParameterMapping, ParameterType } from "@proto/config/v1/call";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Info } from "lucide-react";
import { SmartTemplateEditor } from "./smart-template-editor";

interface InputTransformerEditorProps {
    transformer?: InputTransformer;
    onChange: (transformer: InputTransformer) => void;
    parameters?: HttpParameterMapping[];
}

/**
 * Editor for InputTransformer configuration.
 */
export function InputTransformerEditor({ transformer, onChange, parameters = [] }: InputTransformerEditorProps) {
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

    const variables = useMemo(() => {
        return parameters
            .map(p => p.schema?.name)
            .filter((n): n is string => !!n);
    }, [parameters]);

    const initialTestData = useMemo(() => {
        const data: Record<string, any> = {};
        parameters.forEach(p => {
            if (p.schema?.name) {
                if (p.schema.type === ParameterType.STRING) {
                    data[p.schema.name] = "example_value";
                } else if (p.schema.type === ParameterType.NUMBER || p.schema.type === ParameterType.INTEGER) {
                    data[p.schema.name] = 123;
                } else if (p.schema.type === ParameterType.BOOLEAN) {
                    data[p.schema.name] = true;
                } else {
                    data[p.schema.name] = null;
                }
            }
        });
        // If no parameters, provide a hint
        if (Object.keys(data).length === 0) {
            return "{\n  \"example_var\": \"value\"\n}";
        }
        return JSON.stringify(data, null, 2);
    }, [parameters]);

    return (
        <div className="space-y-4">
             {transformer?.webhook && (
                <Card className="bg-muted/50 mb-4">
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Info className="h-4 w-4" />
                            A webhook is configured for input transformation. The template below will be ignored if the webhook is active.
                        </div>
                    </CardContent>
                </Card>
            )}

            <div className="space-y-2">
                <SmartTemplateEditor
                    value={template}
                    onChange={handleTemplateChange}
                    variables={variables}
                    initialTestData={initialTestData}
                    label="Request Body Template"
                    description="Use Jinja2 syntax to construct the request body. Variables from input parameters are available."
                    placeholder='{"message": "{{ input_message }}"}'
                />
            </div>
        </div>
    );
}
