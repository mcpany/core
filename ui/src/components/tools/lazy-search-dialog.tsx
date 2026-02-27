/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Loader2, Search, Wrench } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface SearchResult {
  tool: {
    name: string;
    description: string;
    inputSchema: any;
  };
  score: number;
}

export function LazySearchDialog({
  trigger,
  onSelect,
}: {
  trigger?: React.ReactNode;
  onSelect?: (toolName: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleSearch = async () => {
    if (!query.trim()) return;
    setLoading(true);
    setResults([]);
    try {
      const res = await apiClient.searchTools(query);
      setResults(res);
    } catch (e: any) {
      toast({
        variant: "destructive",
        title: "Search failed",
        description: e.message || "Failed to search tools",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSearch();
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button variant="outline" className="w-full justify-start text-left font-normal text-muted-foreground">
            <Search className="mr-2 h-4 w-4" />
            Semantic Tool Search...
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Semantic Tool Search</DialogTitle>
          <DialogDescription>
            Find tools by describing what you want to do.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="flex gap-2">
            <Input
              placeholder="e.g. 'I need to check the weather in London'"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              autoFocus
            />
            <Button onClick={handleSearch} disabled={loading}>
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Search"}
            </Button>
          </div>
          <ScrollArea className="h-[300px] rounded-md border p-4">
            {results.length === 0 && !loading && (
              <div className="text-center text-muted-foreground text-sm mt-10">
                {query ? "No matching tools found." : "Enter a query to start searching."}
              </div>
            )}
            <div className="flex flex-col gap-3">
              {results.map((item, i) => (
                <div
                  key={i}
                  className="flex items-start justify-between p-3 rounded-lg border hover:bg-muted/50 transition-colors cursor-pointer"
                  onClick={() => {
                    if (onSelect) {
                      onSelect(item.tool.name);
                      setOpen(false);
                    }
                  }}
                >
                  <div className="space-y-1">
                    <div className="font-semibold flex items-center gap-2">
                      <Wrench className="h-4 w-4 text-primary" />
                      {item.tool.name}
                    </div>
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {item.tool.description || "No description provided."}
                    </p>
                  </div>
                  <Badge variant={item.score > 0.8 ? "default" : "secondary"}>
                    {(item.score * 100).toFixed(0)}%
                  </Badge>
                </div>
              ))}
            </div>
          </ScrollArea>
        </div>
      </DialogContent>
    </Dialog>
  );
}
