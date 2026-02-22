/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useSearchParams } from "next/navigation";
import { InstallWizard } from "@/components/marketplace/wizard/install-wizard";
import { Suspense } from "react";
import { Loader2 } from "lucide-react";

function InstallPageContent() {
    const searchParams = useSearchParams();
    const repo = searchParams.get("repo") || "";
    const name = searchParams.get("name") || "";
    const description = searchParams.get("description") || "";
    const templateId = searchParams.get("templateId") || undefined;

    return (
        <InstallWizard
            initialRepo={repo}
            initialName={name}
            initialDescription={description}
            templateId={templateId}
        />
    );
}

export default function InstallPage() {
    return (
        <div className="flex flex-col h-full p-8 overflow-y-auto">
            <div className="mb-8">
                <h1 className="text-3xl font-bold tracking-tight">Install Service</h1>
                <p className="text-muted-foreground mt-2">
                    Configure and deploy a new MCP service from a template or community repository.
                </p>
            </div>

            <Suspense fallback={
                <div className="flex items-center justify-center h-64">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }>
                <InstallPageContent />
            </Suspense>
        </div>
    );
}
