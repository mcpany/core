/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

export function LazyMcpDashboard() {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<{name: string, description: string}[]>([]);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const handleSearch = async () => {
        if (!query.trim()) {
            setResults([]);
            return;
        }

        setLoading(true);
        try {
            const data = await apiClient.listTools(undefined, query);
            setResults(data.tools || []);
        } catch (e) {
            console.error("Failed to search tools", e);
            toast({
                title: "Search Failed",
                description: String(e),
                variant: "destructive"
            });
            setResults([]);
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
                        disabled={loading}
                    />
                    <Button onClick={handleSearch} disabled={loading || !query.trim()}>
                        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Search"}
                    </Button>
                </div>
                {results.length > 0 && (
                    <div className="space-y-2">
                        <div className="text-sm font-semibold">Search Results (LazyMCP Filtered)</div>
                        {results.map((r, i) => (
                            <div key={i} className="flex justify-between items-center p-2 border rounded text-sm bg-muted/20">
                                <div>
                                    <div className="font-medium text-primary">{r.name}</div>
                                    <div className="text-xs text-muted-foreground line-clamp-1">{r.description}</div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
                {results.length === 0 && !loading && query.trim() && (
                    <div className="text-sm text-muted-foreground p-4 text-center border rounded">
                        No tools found matching your intent.
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
