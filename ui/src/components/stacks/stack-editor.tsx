/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { SourceEditor } from "@/components/services/editor/source-editor";
import { Loader2, Save } from "lucide-react";

interface StackEditorProps {
    initialValue: string;
    onSave: (value: string) => Promise<void>;
    isSaving?: boolean;
}

export function StackEditor({ initialValue, onSave, isSaving }: StackEditorProps) {
    const [value, setValue] = useState(initialValue);

    // Update local state if initialValue changes (e.g. loaded from API)
    useEffect(() => {
        setValue(initialValue);
    }, [initialValue]);

    return (
        <div className="flex flex-col gap-4 h-full">
            <div className="flex justify-end bg-background/50 backdrop-blur-sm p-2 sticky top-0 z-10 border-b">
                <Button onClick={() => onSave(value)} disabled={isSaving}>
                    {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    Save & Deploy
                </Button>
            </div>
            <div className="flex-1 border rounded-md overflow-hidden min-h-[500px]">
                <SourceEditor value={value} onChange={(v) => setValue(v || "")} />
            </div>
        </div>
    );
}
