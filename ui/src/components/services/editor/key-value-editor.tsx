/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



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
    id: string;
    key: string;
    value: string;
}

/**
 * Intent: Document KeyValueEditor
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * KeyValueEditor allows editing a map of strings.
 */
export function KeyValueEditor({ initialValues, onChange, keyPlaceholder = "Key", valuePlaceholder = "Value" }: KeyValueEditorProps) {
    const [pairs, setPairs] = useState<KeyValuePair[]>([]);

    useEffect(() => {
        if (initialValues) {
            setPairs(Object.entries(initialValues).map(([key, value]) => ({ id: crypto.randomUUID(), key, value })));
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
        const newPairs = [...pairs, { id: crypto.randomUUID(), key: "", value: "" }];
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
            {/* ⚡ BOLT: [Render Optimization] Use stable IDs instead of array index for list keys to prevent React state/focus loss and unnecessary remounts.
                Randomized Selection from Top 5 High-Impact Targets (Render Category) */}
            {pairs.map((pair, index) => (
                <div key={pair.id} className="flex items-center gap-2">
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
