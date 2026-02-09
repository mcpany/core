/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, use } from "react";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

interface PageProps {
    params: Promise<{ stackId: string }>;
}

export default function StackDetailPage(props: PageProps) {
    const params = use(props.params);
    const [initialData, setInitialData] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const isNew = params.stackId === "new";

    useEffect(() => {
        if (!isNew) {
            // Fetch initial data to populate name/desc
            // We can reuse getCollection to get metadata
            apiClient.getCollection(params.stackId)
                .then(data => {
                    setInitialData(data);
                })
                .catch(console.error)
                .finally(() => setLoading(false));
        } else {
            setLoading(false);
        }
    }, [isNew, params.stackId]);

    if (loading) {
        return (
            <div className="h-[calc(100vh-4rem)] flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="h-[calc(100vh-4rem)] p-8 pt-6">
            <StackEditor
                stackId={isNew ? undefined : params.stackId}
                initialData={initialData}
            />
        </div>
    );
}
