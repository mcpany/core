/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { StackEditor } from "@/components/stacks/stack-editor";
import { use } from "react";

/**
 * StackEditorPage component.
 * @param props - The component props.
 * @param props.params - The params property.
 * @returns The rendered component.
 */
export default function StackEditorPage({ params }: { params: Promise<{ stackId: string }> }) {
  // Unwrap params using `use` hook (Next.js 15 pattern, or standard await if async component)
  // Since this is "use client", we can't make the component async.
  // We use `use` to unwrap the promise if it is one.
  const resolvedParams = use(params);

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">
            {resolvedParams.stackId === 'new' ? 'New Stack' : 'Edit Stack'}
        </h1>
      </div>
      <div className="flex-1 min-h-0">
        <StackEditor stackId={resolvedParams.stackId} />
      </div>
    </div>
  );
}
