/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceExplorer } from "@/components/resources/resource-explorer";

export default function ResourcesPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8">
      <ResourceExplorer />
    </div>
  );
}
