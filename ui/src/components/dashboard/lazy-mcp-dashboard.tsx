/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

/**
 * Summary: Renders the Lazy-MCP Tool Search Dashboard component for managing the on-demand tool index.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.JSX.Element: The rendered React component containing the search input and results.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - Updates local component state (query and results) on user interaction.
 */
export function LazyMcpDashboard() {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<{name: string, score: number}[]>([]);

    const handleSearch = () => {
        // Mock search execution
        if (query) {
            setResults([
                { name: "fs:read", score: 0.95 },
                { name: "fs:write", score: 0.88 },
                { name: "db:query", score: 0.65 }
            ]);
        } else {
            setResults([]);
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
                    <Button onClick={handleSearch}>Search</Button>
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
