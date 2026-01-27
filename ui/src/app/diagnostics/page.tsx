/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SystemHealth } from "@/components/diagnostics/system-health";
import { ServiceDoctor } from "@/components/diagnostics/service-doctor";

/**
 * DiagnosticsPage component.
 * @returns The rendered component.
 */
export default function DiagnosticsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">System Diagnostics</h2>
            <p className="text-muted-foreground">Monitor system health, connectivity, and environment status.</p>
        </div>
      </div>

      <div className="flex-1 overflow-auto rounded-md border bg-muted/10 p-4">
           <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
               <div className="space-y-6">
                   <h3 className="text-xl font-semibold">Global Health</h3>
                   <SystemHealth />
               </div>
               <div className="space-y-6">
                   <h3 className="text-xl font-semibold">Service Doctor</h3>
                   <ServiceDoctor />
               </div>
           </div>
      </div>
    </div>
  );
}
