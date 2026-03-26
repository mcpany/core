/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Suspense } from "react";
import { PlaygroundClientPro } from "@/components/playground/pro/playground-client-pro";
import { Loader2 } from "lucide-react";

/**
 * PlaygroundPage component.
 * @returns The rendered component.
 */
export default function PlaygroundPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-5rem)]">
      <h1 className="sr-only">Console</h1>
      <Suspense
        fallback={
        <div className="flex flex-1 items-center justify-center h-full w-full">
            <div className="flex flex-col items-center gap-4 text-muted-foreground">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                <p className="text-sm font-medium tracking-wide">Loading playground...</p>
            </div>
          </div>
        }
      >
        <PlaygroundClientPro />
      </Suspense>
    </div>
  );
}
