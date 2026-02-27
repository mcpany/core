/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Search, Loader2, Zap, Server } from "lucide-react";
import { ToolSearchResult, discoveryApi } from "@/lib/api-discovery";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

/**
 * Renders the search console for testing tool discovery.
 *
 * @returns The rendered SearchConsole component.
 */
export function SearchConsole() {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<ToolSearchResult[]>([]);
    const [loading, setLoading] = useState(false);
    const [hasSearched, setHasSearched] = useState(false);

    const handleSearch = async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        if (!query.trim()) return;

        setLoading(true);
        setHasSearched(true);
        try {
            const data = await discoveryApi.searchTools(query);
            setResults(data);
        } catch (error) {
            console.error("Search failed", error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Card className="w-full">
            <CardHeader>
                <CardTitle>Tool Search Simulator</CardTitle>
                <CardDescription>
                    Test the semantic search engine used by agents to discover tools on-demand.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSearch} className="flex space-x-2 mb-6">
                    <div className="relative flex-1">
                        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            type="search"
                            placeholder="Describe what you want to do (e.g., 'get weather in tokyo')"
                            className="pl-9"
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                        />
                    </div>
                    <Button type="submit" disabled={loading}>
                        {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                        Search
                    </Button>
                </form>

                <div className="rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[200px]">Tool</TableHead>
                                <TableHead className="w-[150px]">Service</TableHead>
                                <TableHead className="w-[100px]">Relevance</TableHead>
                                <TableHead>Description</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                                        Searching index...
                                    </TableCell>
                                </TableRow>
                            ) : results.length > 0 ? (
                                results.map((result, idx) => (
                                    <TableRow key={idx}>
                                        <TableCell className="font-medium flex items-center gap-2">
                                            <Zap className="h-3 w-3 text-amber-500" />
                                            {result.toolName}
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="outline" className="flex w-fit items-center gap-1 font-normal">
                                                <Server className="h-3 w-3" /> {result.serviceName}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <div className="h-2 w-16 bg-muted rounded-full overflow-hidden">
                                                    <div
                                                        className="h-full bg-primary"
                                                        style={{ width: `${result.relevance * 100}%` }}
                                                    />
                                                </div>
                                                <span className="text-xs text-muted-foreground">
                                                    {(result.relevance * 100).toFixed(0)}%
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-muted-foreground text-sm truncate max-w-md">
                                            {result.description}
                                        </TableCell>
                                    </TableRow>
                                ))
                            ) : (
                                <TableRow>
                                    <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                                        {hasSearched ? "No matching tools found." : "Enter a query to test discovery."}
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            </CardContent>
        </Card>
    );
}
