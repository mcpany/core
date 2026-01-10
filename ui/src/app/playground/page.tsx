/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Suspense } from "react";
import { PlaygroundClientPro } from "@/components/playground/pro/playground-client-pro";
import { Loader2 } from "lucide-react";

export default function PlaygroundPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-5rem)]">
      <Suspense
        fallback={
          <div className="flex items-center justify-center h-full">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <span className="ml-2 text-muted-foreground">Loading playground...</span>
          </div>
        }
      >
        <PlaygroundClientPro />
      </Suspense>
    </div>
  );
}
