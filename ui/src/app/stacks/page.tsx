/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { StackList } from "@/components/stacks/stack-list";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Stacks</h2>
            <p className="text-muted-foreground">Manage service collections and deployments.</p>
        </div>
      </div>
      <StackList />
    </div>
  );
}
