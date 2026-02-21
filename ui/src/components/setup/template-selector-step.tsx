/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ServiceTemplateSelector } from "@/components/services/service-template-selector";
import { ServiceTemplate } from "@/lib/templates";

export function TemplateSelectorStep({ onSelect }: { onSelect: (t: ServiceTemplate) => void }) {
  return (
    <div className="w-full max-w-4xl space-y-4">
        <div className="text-center space-y-2 mb-8">
            <h1 className="text-3xl font-bold tracking-tight">Choose a Starter Template</h1>
            <p className="text-muted-foreground text-lg">Select a service type to connect. You can add more later.</p>
        </div>

        {/* We reuse the existing selector which has search and categories */}
        <ServiceTemplateSelector onSelect={onSelect} />
    </div>
  );
}
