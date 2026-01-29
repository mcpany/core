/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { SERVICE_TEMPLATES, ServiceTemplate } from "@/lib/templates";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Search, Star } from "lucide-react";

interface ServiceTemplateSelectorProps {
  onSelect: (template: ServiceTemplate) => void;
}

/**
 * List of available categories for service templates.
 */
const CATEGORIES = ["All", "Web", "Productivity", "Database", "Dev Tools", "Cloud", "System", "Utility", "Other"];

/**
 * ServiceTemplateSelector component.
 * Allows users to browse and search for service templates.
 *
 * @param onSelect - Callback when a template is selected.
 */
export function ServiceTemplateSelector({ onSelect }: ServiceTemplateSelectorProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");

  const filteredTemplates = useMemo(() => {
    return SERVICE_TEMPLATES.filter((template) => {
      const matchesCategory = selectedCategory === "All" || template.category === selectedCategory;
      const matchesSearch = template.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            template.description.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesCategory && matchesSearch;
    }).sort((a, b) => {
        // Featured first
        if (a.featured && !b.featured) return -1;
        if (!a.featured && b.featured) return 1;
        return a.name.localeCompare(b.name);
    });
  }, [searchQuery, selectedCategory]);

  return (
    <div className="space-y-4 p-1">
      <div className="flex flex-col gap-4 sticky top-0 bg-background/95 backdrop-blur z-10 pb-2 border-b">
        <div className="relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
                placeholder="Search templates..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-8"
            />
        </div>

        <div className="flex overflow-x-auto pb-2 gap-2 scrollbar-thin scrollbar-thumb-muted scrollbar-track-transparent">
            {CATEGORIES.map(cat => (
                <button
                    key={cat}
                    onClick={() => setSelectedCategory(cat)}
                    className={cn(
                        "px-3 py-1.5 rounded-full text-xs font-medium transition-colors whitespace-nowrap border",
                        selectedCategory === cat
                            ? "bg-primary text-primary-foreground border-primary"
                            : "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground border-transparent"
                    )}
                >
                    {cat}
                </button>
            ))}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pb-4">
        {filteredTemplates.map((template) => {
          const Icon = template.icon;
          return (
            <div
              key={template.id}
              className={cn(
                  "flex flex-col items-start gap-3 p-4 border rounded-lg cursor-pointer transition-all duration-200 relative overflow-hidden h-full min-h-[140px]",
                  "hover:bg-muted/50 hover:border-primary/50 hover:shadow-sm",
                  "active:scale-[0.98]",
                  template.featured && "border-primary/20 bg-primary/5"
              )}
              onClick={() => onSelect(template)}
            >
              <div className="flex w-full justify-between items-start">
                  <div className={cn("p-2 rounded-md shrink-0", template.featured ? "bg-primary/20 text-primary" : "bg-primary/10 text-primary")}>
                    <Icon className="h-5 w-5" />
                  </div>
                  {template.featured && (
                      <Badge variant="secondary" className="bg-yellow-500/10 text-yellow-600 hover:bg-yellow-500/20 border-yellow-200 gap-1 text-[10px] px-1.5">
                          <Star className="h-3 w-3 fill-yellow-600" /> Featured
                      </Badge>
                  )}
              </div>

              <div className="space-y-1 flex-1">
                <h3 className="font-semibold leading-none tracking-tight">{template.name}</h3>
                <p className="text-sm text-muted-foreground leading-snug line-clamp-2">
                  {template.description}
                </p>
              </div>

              {template.category && (
                  <Badge variant="outline" className="mt-auto text-[10px] h-5 px-1.5 bg-background/50">
                      {template.category}
                  </Badge>
              )}
            </div>
          );
        })}

        {filteredTemplates.length === 0 && (
            <div className="col-span-full text-center py-12 text-muted-foreground flex flex-col items-center gap-2">
                <Search className="h-8 w-8 opacity-20" />
                <p>No templates found matching &quot;{searchQuery}&quot;.</p>
                <button
                    onClick={() => { setSearchQuery(""); setSelectedCategory("All"); }}
                    className="text-primary hover:underline text-sm"
                >
                    Clear filters
                </button>
            </div>
        )}
      </div>
    </div>
  );
}
