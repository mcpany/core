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
            .filter((name): name is string => !!name);
    }, [parameters]);

    const defaultTestData = useMemo(() => {
        const data: Record<string, any> = {};
        parameters.forEach(p => {
            if (!p.schema?.name) return;
            switch (p.schema.type) {
                case ParameterType.STRING:
                    data[p.schema.name] = "example_string";
                    break;
                case ParameterType.NUMBER:
                case ParameterType.INTEGER:
                    data[p.schema.name] = 123;
                    break;
                case ParameterType.BOOLEAN:
                    data[p.schema.name] = true;
                    break;
                default:
                    data[p.schema.name] = "value";
            }
        });
        return JSON.stringify(data, null, 2);
    }, [parameters]);

    return (
        <div className="space-y-4">
            <div className="space-y-2">
                <SmartTemplateEditor
                    label="Request Body Template"
                    description="Use Jinja2 syntax to construct the request body. Variables from input are available."
                    value={template}
                    onChange={handleTemplateChange}
                    variables={variables}
                    testData={defaultTestData}
                    placeholder='{"message": "{{ input_message }}"}'
                    className="h-[500px]"
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
