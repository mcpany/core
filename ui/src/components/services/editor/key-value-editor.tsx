/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, X } from "lucide-react";

interface KeyValueEditorProps {
    initialValues?: Record<string, string>;
    onChange: (values: Record<string, string>) => void;
    keyPlaceholder?: string;
    valuePlaceholder?: string;
}

interface KeyValuePair {
    key: string;
    value: string;
}

/**
 * KeyValueEditor allows editing a map of strings.
 */
export function KeyValueEditor({ initialValues, onChange, keyPlaceholder = "Key", valuePlaceholder = "Value" }: KeyValueEditorProps) {
    const [pairs, setPairs] = useState<KeyValuePair[]>([]);

    useEffect(() => {
        if (initialValues) {
            setPairs(Object.entries(initialValues).map(([key, value]) => ({ key, value })));
        } else {
            setPairs([]);
        }
    }, [initialValues]);

    const updateParent = (currentPairs: KeyValuePair[]) => {
        const newValues: Record<string, string> = {};
        currentPairs.forEach(p => {
            if (p.key) {
                newValues[p.key] = p.value;
            }
        });
        onChange(newValues);
    };

    const addPair = () => {
        const newPairs = [...pairs, { key: "", value: "" }];
        setPairs(newPairs);
        // Do not update parent on add, wait for input
    };

    const removePair = (index: number) => {
        const newPairs = pairs.filter((_, i) => i !== index);
        setPairs(newPairs);
        updateParent(newPairs);
    };

    const updatePair = (index: number, field: keyof KeyValuePair, value: string) => {
        const newPairs = pairs.map((p, i) => {
            if (i === index) {
                return { ...p, [field]: value };
            }
            return p;
        });
        setPairs(newPairs);
        updateParent(newPairs);
    };

    return (
        <div className="space-y-2">
            {pairs.map((pair, index) => (
                <div key={index} className="flex items-center gap-2">
                    <Input
                        placeholder={keyPlaceholder}
                        value={pair.key}
                        onChange={(e) => updatePair(index, "key", e.target.value)}
                        className="flex-1"
                    />
                    <Input
                        placeholder={valuePlaceholder}
                        value={pair.value}
                        onChange={(e) => updatePair(index, "value", e.target.value)}
                        className="flex-1"
                    />
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => removePair(index)}
                        className="text-muted-foreground hover:text-destructive"
                    >
                        <X className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addPair}
                className="w-full"
            >
                <Plus className="mr-2 h-3 w-3" /> Add Item
            </Button>
        </div>
    );
}
