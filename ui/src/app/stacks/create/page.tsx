/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { StackEditor } from "@/components/stacks/stack-editor";

/**
 * StackCreatePage component.
 * @returns The rendered component.
 */
export default function StackCreatePage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <StackEditor isNew={true} />
    </div>
  );
}
