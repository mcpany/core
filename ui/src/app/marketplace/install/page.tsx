/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { InstallWizard } from "@/components/marketplace/wizard/install-wizard";
import { Suspense } from "react";
import { Loader2 } from "lucide-react";

export default function InstallPage() {
    return (
        <Suspense fallback={
            <div className="flex h-screen items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        }>
            <InstallWizard />
        </Suspense>
    );
}
