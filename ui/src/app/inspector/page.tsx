/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Inspector } from "@/components/inspector/inspector"

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  return (
    <div className="h-[calc(100vh-4rem)] overflow-hidden bg-background">
      <Inspector />
    </div>
  )
}
