/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { scoreMatch } from "@/lib/search";

export function LazyMcpDashboard() {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<{name: string, score: number}[]>([]);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const handleSearch = async () => {
        if (!query.trim()) {
            setResults([]);
            return;
        }

        setLoading(true);
        try {
            const data = await apiClient.listTools();
            const tools = Array.isArray(data) ? data : (data.tools || []);

            const scoredTools = tools.map((tool: any) => {
                const nameScore = scoreMatch(query, tool.name || "");
                const descScore = scoreMatch(query, tool.description || "");
                // Weighted score favoring name over description
                const finalScore = Math.max(nameScore, descScore * 0.8);
                return { name: tool.name, score: finalScore };
            });

            // Threshold and sort
            const filteredAndSorted = scoredTools
                .filter(t => t.score >= 0.85)
                .sort((a, b) => b.score - a.score);

            setResults(filteredAndSorted);
        } catch (e: any) {
            console.error("Failed to search tools", e);
            toast({
                title: "Search failed",
                description: e.message || "Failed to search tools.",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Card className="w-full mt-4">
            <CardHeader>
                <CardTitle className="text-sm font-medium">Lazy-MCP Tool Search Dashboard</CardTitle>
                <CardDescription>Manage the on-demand tool index and monitor similarity search hits.</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="flex gap-2 mb-4">
                    <Input
                        placeholder="Search tool index by intent..."
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                    />
                    <Button onClick={handleSearch} disabled={loading}>
                        {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        Search
                    </Button>
                </div>
                {results.length > 0 && (
                    <div className="space-y-2">
                        <div className="text-sm font-semibold">Search Results (Threshold: 0.85)</div>
                        {results.map((r, i) => (
                            <div key={i} className="flex justify-between items-center p-2 border rounded text-sm">
                                <span>{r.name}</span>
                                <span className={`font-mono ${r.score >= 0.85 ? 'text-green-500' : 'text-amber-500'}`}>
                                    {r.score.toFixed(2)}
                                </span>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
