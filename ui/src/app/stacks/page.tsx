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
    <div className="h-full p-8 pt-6">
      <StackList />
    </div>
  );
}
