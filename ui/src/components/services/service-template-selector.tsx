/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { SERVICE_TEMPLATES, ServiceTemplate } from "@/lib/templates";
import { SERVICE_REGISTRY } from "@/lib/service-registry";
import { marketplaceService, CommunityServer } from "@/lib/marketplace-service";
import { mapRegistryToTemplate, mapCommunityToTemplate } from "@/lib/catalog-mapper";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Search, Star, ShieldCheck, Globe, Loader2 } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface ServiceTemplateSelectorProps {
  onSelect: (template: ServiceTemplate) => void;
}

/**
 * List of available service template categories.
 */
const CATEGORIES = ["All", "Web", "Productivity", "Database", "Dev Tools", "Cloud", "System", "Utility", "Other"];

/**
 * ServiceTemplateSelector component.
 * Allows users to browse and search for service templates from Official, Registry, and Community sources.
 *
 * @param onSelect - Callback when a template is selected.
 */
export function ServiceTemplateSelector({ onSelect }: ServiceTemplateSelectorProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");
  const [sourceFilter, setSourceFilter] = useState("all");

  const [communityTemplates, setCommunityTemplates] = useState<ServiceTemplate[]>([]);
  const [isLoadingCommunity, setIsLoadingCommunity] = useState(true);

  // Load Community Servers
  useEffect(() => {
      const fetchCommunity = async () => {
          try {
              const servers = await marketplaceService.fetchCommunityServers();
              const templates = servers.map(mapCommunityToTemplate);
              setCommunityTemplates(templates);
          } catch (e) {
              console.error("Failed to fetch community servers", e);
          } finally {
              setIsLoadingCommunity(false);
          }
      };
      fetchCommunity();
  }, []);

  const combinedCatalog = useMemo(() => {
      // 1. Official Templates (Static)
      // Mark them as "official" source for UI
      const official = SERVICE_TEMPLATES.map(t => ({ ...t, source: "official" }));

      // 2. Registry Items (Static Structured)
      const registry = SERVICE_REGISTRY.map(mapRegistryToTemplate);

      // 3. Community (Dynamic)
      // Already mapped in state

      // Merge and Deduplicate
      // Priority: Official > Registry > Community
      const map = new Map<string, ServiceTemplate>();

      // Add Official
      official.forEach(t => map.set(t.name.toLowerCase(), t as ServiceTemplate));

      // Add Registry (if not exists)
      registry.forEach(t => {
          if (!map.has(t.name.toLowerCase())) {
              map.set(t.name.toLowerCase(), t);
          }
      });

      // Add Community (if not exists)
      communityTemplates.forEach(t => {
          // Fuzzy match or exact match? Exact for now.
          if (!map.has(t.name.toLowerCase())) {
              map.set(t.name.toLowerCase(), t);
          }
      });

      return Array.from(map.values());
  }, [communityTemplates]);

  const filteredTemplates = useMemo(() => {
    return combinedCatalog.filter((template) => {
      // Source Filter
      const source = template.source || "official";
      if (sourceFilter !== "all") {
          if (sourceFilter === "official" && source !== "official") return false;
          if (sourceFilter === "verified" && source !== "verified") return false;
          if (sourceFilter === "community" && source !== "community") return false;
      }

      // Category Filter
      // Map community categories to standard ones if possible, or just string match
      const templateCat = template.category || "Other";
      const matchesCategory = selectedCategory === "All" || templateCat === selectedCategory || (selectedCategory === "Other" && !CATEGORIES.includes(templateCat) && templateCat !== "All");

      // Search Filter
      const matchesSearch = template.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            template.description.toLowerCase().includes(searchQuery.toLowerCase());

      return matchesCategory && matchesSearch;
    }).sort((a, b) => {
        // Featured first
        if (a.featured && !b.featured) return -1;
        if (!a.featured && b.featured) return 1;

        // Official first, then Verified, then Community
        const score = (t: ServiceTemplate) => t.source === 'official' ? 3 : t.source === 'verified' ? 2 : 1;
        const scoreA = score(a);
        const scoreB = score(b);
        if (scoreA !== scoreB) return scoreB - scoreA;

        return a.name.localeCompare(b.name);
    });
  }, [combinedCatalog, searchQuery, selectedCategory, sourceFilter]);

  return (
    <div className="space-y-4 p-1">
      <div className="flex flex-col gap-4 sticky top-0 bg-background/95 backdrop-blur z-10 pb-2 border-b">
        <div className="flex gap-2">
            <div className="relative flex-1">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search tools (e.g. Postgres, Slack)..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-8"
                />
            </div>
            <Tabs value={sourceFilter} onValueChange={setSourceFilter} className="w-[400px]">
                <TabsList className="grid w-full grid-cols-4">
                    <TabsTrigger value="all">All</TabsTrigger>
                    <TabsTrigger value="official">Official</TabsTrigger>
                    <TabsTrigger value="verified">Verified</TabsTrigger>
                    <TabsTrigger value="community">Community</TabsTrigger>
                </TabsList>
            </Tabs>
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
          const source = template.source || "official";

          return (
            <div
              key={template.id}
              className={cn(
                  "flex flex-col items-start gap-3 p-4 border rounded-lg cursor-pointer transition-all duration-200 relative overflow-hidden h-full min-h-[140px]",
                  "hover:bg-muted/50 hover:border-primary/50 hover:shadow-sm",
                  "active:scale-[0.98]",
                  template.featured && "border-primary/20 bg-primary/5",
                  source === "community" && "border-dashed"
              )}
              onClick={() => onSelect(template)}
            >
              <div className="flex w-full justify-between items-start">
                  <div className={cn("p-2 rounded-md shrink-0", template.featured ? "bg-primary/20 text-primary" : "bg-primary/10 text-primary")}>
                    <Icon className="h-5 w-5" />
                  </div>
                  <div className="flex gap-1">
                    {source === "official" && (
                        <Badge variant="secondary" className="bg-primary/10 text-primary hover:bg-primary/20 border-primary/20 gap-1 text-[10px] px-1.5">
                            <Star className="h-3 w-3 fill-primary" /> Official
                        </Badge>
                    )}
                    {source === "verified" && (
                        <Badge variant="secondary" className="bg-green-500/10 text-green-600 hover:bg-green-500/20 border-green-200 gap-1 text-[10px] px-1.5">
                            <ShieldCheck className="h-3 w-3" /> Verified
                        </Badge>
                    )}
                    {source === "community" && (
                        <Badge variant="outline" className="text-muted-foreground border-dashed gap-1 text-[10px] px-1.5">
                            <Globe className="h-3 w-3" /> Community
                        </Badge>
                    )}
                  </div>
              </div>

              <div className="space-y-1 flex-1">
                <h3 className="font-semibold leading-none tracking-tight flex items-center gap-2">
                    {template.name}
                </h3>
                <p className="text-sm text-muted-foreground leading-snug line-clamp-2">
                  {template.description}
                </p>
              </div>

              <div className="flex justify-between w-full items-end mt-auto">
                  {template.category && (
                      <Badge variant="outline" className="text-[10px] h-5 px-1.5 bg-background/50">
                          {template.category}
                      </Badge>
                  )}
                  {template.url && (
                      <span className="text-[10px] text-muted-foreground flex items-center gap-1 opacity-50">
                          <Globe className="h-3 w-3" />
                          Source
                      </span>
                  )}
              </div>
            </div>
          );
        })}

        {filteredTemplates.length === 0 && (
            <div className="col-span-full text-center py-12 text-muted-foreground flex flex-col items-center gap-2">
                {isLoadingCommunity ? (
                    <>
                        <Loader2 className="h-8 w-8 animate-spin" />
                        <p>Loading community catalog...</p>
                    </>
                ) : (
                    <>
                        <Search className="h-8 w-8 opacity-20" />
                        <p>No templates found matching &quot;{searchQuery}&quot;.</p>
                        <button
                            onClick={() => { setSearchQuery(""); setSelectedCategory("All"); setSourceFilter("all"); }}
                            className="text-primary hover:underline text-sm"
                        >
                            Clear filters
                        </button>
                    </>
                )}
            </div>
        )}
      </div>
    </div>
  );
}
