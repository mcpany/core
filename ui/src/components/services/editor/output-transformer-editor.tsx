/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { OutputTransformer, OutputTransformer_OutputFormat } from "@proto/config/v1/call";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { KeyValueEditor } from "./key-value-editor";
import { Separator } from "@/components/ui/separator";
import { SmartTemplateEditor } from "./smart-template-editor";

interface OutputTransformerEditorProps {
    transformer?: OutputTransformer;
    onChange: (transformer: OutputTransformer) => void;
}

/**
 * Editor for OutputTransformer configuration.
 */
export function OutputTransformerEditor({ transformer, onChange }: OutputTransformerEditorProps) {
    const [localTransformer, setLocalTransformer] = useState<OutputTransformer>({
        format: OutputTransformer_OutputFormat.JSON,
        extractionRules: {},
        template: "",
        jqQuery: "",
        ...transformer
    });

    useEffect(() => {
        setLocalTransformer({
            format: OutputTransformer_OutputFormat.JSON,
            extractionRules: {},
            template: "",
            jqQuery: "",
            ...transformer
        });
    }, [transformer]);

    const updateTransformer = (updates: Partial<OutputTransformer>) => {
        const newVal = { ...localTransformer, ...updates };
        setLocalTransformer(newVal);
        onChange(newVal);
    };

    const variables = useMemo(() => {
        return Object.keys(localTransformer.extractionRules || {});
    }, [localTransformer.extractionRules]);

    const defaultTestData = useMemo(() => {
        const data: Record<string, string> = {};
        variables.forEach(v => {
            data[v] = `example_${v}`;
        });
        return JSON.stringify(data, null, 2);
    }, [variables]);

    return (
        <div className="space-y-6">
            <div className="space-y-2">
                <Label htmlFor="output-format">Output Format</Label>
                <Select
                    value={localTransformer.format.toString()}
                    onValueChange={(val) => updateTransformer({ format: parseInt(val) })}
                >
                    <SelectTrigger id="output-format">
                        <SelectValue placeholder="Format" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value={OutputTransformer_OutputFormat.JSON.toString()}>JSON (JSONPath)</SelectItem>
                        <SelectItem value={OutputTransformer_OutputFormat.XML.toString()}>XML (XPath)</SelectItem>
                        <SelectItem value={OutputTransformer_OutputFormat.TEXT.toString()}>Text (Regex)</SelectItem>
                        <SelectItem value={OutputTransformer_OutputFormat.JQ.toString()}>JQ Query</SelectItem>
                        <SelectItem value={OutputTransformer_OutputFormat.RAW_BYTES.toString()}>Raw Bytes</SelectItem>
                    </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                    Select how the upstream response should be parsed.
                </p>
            </div>

            {localTransformer.format === OutputTransformer_OutputFormat.JQ && (
                <div className="space-y-2">
                    <Label htmlFor="jq-query">JQ Query</Label>
                    <Input
                        id="jq-query"
                        value={localTransformer.jqQuery}
                        onChange={(e) => updateTransformer({ jqQuery: e.target.value })}
                        placeholder=".items[] | .name"
                        className="font-mono"
                    />
                     <p className="text-xs text-muted-foreground">
                        Enter a JQ query to transform the JSON response.
                    </p>
                </div>
            )}

            {(localTransformer.format === OutputTransformer_OutputFormat.JSON ||
              localTransformer.format === OutputTransformer_OutputFormat.XML ||
              localTransformer.format === OutputTransformer_OutputFormat.TEXT) && (
                <div className="space-y-2">
                    <Label>Extraction Rules</Label>
                    <p className="text-xs text-muted-foreground mb-2">
                        Map field names to extraction expressions (JSONPath, XPath, or Regex).
                    </p>
                    <KeyValueEditor
                        initialValues={localTransformer.extractionRules}
                        onChange={(rules) => updateTransformer({ extractionRules: rules })}
                        keyPlaceholder="Field Name"
                        valuePlaceholder="Expression ($.store.book[0].title)"
                    />
                </div>
            )}

            <Separator />

            <div className="space-y-2">
                <SmartTemplateEditor
                    label="Result Template (Optional)"
                    description="Use Jinja2 syntax to format the final output using extracted fields. If empty, the structured result is returned."
                    value={localTransformer.template}
                    onChange={(val) => updateTransformer({ template: val })}
                    placeholder="Weather in {{ location }} is {{ temperature }}."
                    variables={variables}
                    testData={defaultTestData}
                    className="h-[300px]"
                />
            </div>
        </div>
    );
}
