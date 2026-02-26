/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { Plus, Trash2, Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export interface EnvVar {
    key: string;
    value: string;
}

interface EnvVarEditorProps {
    value: EnvVar[];
    onChange: (vars: EnvVar[]) => void;
}

/**
 * EnvVarEditor allows editing a list of environment variables.
 */
export function EnvVarEditor({ value, onChange }: EnvVarEditorProps) {
    const handleAdd = () => {
        onChange([...value, { key: "", value: "" }]);
    };

    const handleRemove = (index: number) => {
        const newValue = [...value];
        newValue.splice(index, 1);
        onChange(newValue);
    };

    const handleChange = (index: number, field: "key" | "value", newVal: string) => {
        const newValue = [...value];
        newValue[index] = { ...newValue[index], [field]: newVal };
        onChange(newValue);
    };

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <Label className="text-sm font-medium">Environment Variables</Label>
            </div>

            <div className="space-y-2">
                {value.map((item, index) => (
                    <div key={index} className="flex gap-2 items-center">
                        <Input
                            placeholder="KEY"
                            value={item.key}
                            onChange={(e) => handleChange(index, "key", e.target.value)}
                            className="flex-1 font-mono text-xs"
                        />
                        <div className="flex-1 relative">
                            <Input
                                placeholder="VALUE"
                                value={item.value}
                                onChange={(e) => handleChange(index, "value", e.target.value)}
                                className="font-mono text-xs pr-8"
                                type="password"
                            />
                        </div>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            onClick={() => handleRemove(index)}
                            className="h-9 w-9 shrink-0 text-muted-foreground hover:text-destructive"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                ))}
            </div>

            <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAdd}
                className="w-full text-xs border-dashed"
            >
                <Plus className="mr-2 h-3 w-3" /> Add Variable
            </Button>
        </div>
    );
}
