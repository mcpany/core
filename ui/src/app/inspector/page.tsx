/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Suspense } from "react"
import { Loader2 } from "lucide-react"
import { InspectorView } from "@/components/inspector/inspector-view"

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  return (
    <div className="h-full bg-background overflow-hidden">
      <Suspense fallback={<div className="flex items-center justify-center h-full"><Loader2 className="h-8 w-8 animate-spin" /></div>}>
        <InspectorView />
      </Suspense>
    </div>
  )
}
