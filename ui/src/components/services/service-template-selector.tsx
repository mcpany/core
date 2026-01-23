/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { SERVICE_TEMPLATES, ServiceTemplate } from "@/lib/templates";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Search } from "lucide-react";

interface ServiceTemplateSelectorProps {
  onSelect: (template: ServiceTemplate) => void;
}

/**
 * ServiceTemplateSelector.
 *
 * @param { onSelect - The { onSelect.
 */
export function ServiceTemplateSelector({ onSelect }: ServiceTemplateSelectorProps) {
  const [selectedCategory, setSelectedCategory] = useState<string>("All");
  const [searchQuery, setSearchQuery] = useState<string>("");

  const categories = useMemo(() => {
    const cats = Array.from(new Set(SERVICE_TEMPLATES.map(t => t.category || "Other")));
    return ["All", ...cats.sort()];
  }, []);

  const filteredTemplates = useMemo(() => {
    return SERVICE_TEMPLATES.filter(template => {
      const matchesCategory = selectedCategory === "All" || (template.category || "Other") === selectedCategory;
      const matchesSearch = template.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            template.description.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesCategory && matchesSearch;
    });
  }, [selectedCategory, searchQuery]);

  return (
    <div className="space-y-4">
      {/* Search and Filters */}
      <div className="space-y-3">
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search templates..."
            className="pl-8"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <div className="flex flex-wrap gap-2">
          {categories.map(category => (
            <Button
              key={category}
              variant={selectedCategory === category ? "default" : "secondary"}
              size="sm"
              onClick={() => setSelectedCategory(category)}
              className="text-xs h-7"
            >
              {category}
            </Button>
          ))}
        </div>
      </div>

      {/* Templates Grid */}
      <div className="grid grid-cols-1 gap-4 p-1 max-h-[60vh] overflow-y-auto">
        {filteredTemplates.length > 0 ? (
          filteredTemplates.map((template) => {
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
                <div className="space-y-1 flex-1">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold leading-none tracking-tight">{template.name}</h3>
                    {template.category && (
                      <Badge variant="outline" className="text-[10px] px-1 h-5 text-muted-foreground">
                        {template.category}
                      </Badge>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground leading-snug">
                    {template.description}
                  </p>
                </div>
              </div>
            );
          })
        ) : (
          <div className="text-center py-8 text-muted-foreground">
            No templates found matching your criteria.
          </div>
        )}
      </div>
    </div>
  );
}
