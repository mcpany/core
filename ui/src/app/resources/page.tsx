/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceExplorer } from "@/components/resources/resource-explorer";

export default function ResourcesPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 gap-4">
      <div className="flex items-center justify-between shrink-0">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
      </div>
      <div className="flex-1 min-h-0">
        <ResourceExplorer />
      </div>
    </div>
  );
}
