/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { SERVICE_TEMPLATES, ServiceTemplate } from "@/lib/templates";
import { cn } from "@/lib/utils";

interface ServiceTemplateSelectorProps {
  onSelect: (template: ServiceTemplate) => void;
}

/**
 * ServiceTemplateSelector.
 *
 * @param { onSelect - The { onSelect.
 */
export function ServiceTemplateSelector({ onSelect }: ServiceTemplateSelectorProps) {
  return (
    <div className="grid grid-cols-1 gap-4 p-1">
      {SERVICE_TEMPLATES.map((template) => {
        const Icon = template.icon;
        return (
          <div
            key={template.id}
            className={cn(
                "flex items-start gap-4 p-4 border rounded-lg cursor-pointer transition-all duration-200",
                "hover:bg-muted/50 hover:border-primary/50 hover:shadow-sm",
                "active:scale-[0.98]"
            )}
            onClick={() => onSelect(template)}
          >
            <div className="mt-1 p-2 bg-primary/10 rounded-md text-primary shrink-0">
              <Icon className="h-5 w-5" />
            </div>
            <div className="space-y-1">
              <h3 className="font-semibold leading-none tracking-tight">{template.name}</h3>
              <p className="text-sm text-muted-foreground leading-snug">
                {template.description}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}
