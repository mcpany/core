/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { InspectorView } from "@/components/inspector/inspector-view";

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  return (
    <div className="h-full w-full">
      <InspectorView />
    </div>
  );
}
