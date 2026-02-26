/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { indexClient, IndexedTool, IndexStats } from "@/lib/index-client";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Search, Download, Database, Activity, AlertCircle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";

/**
 * IndexPage component.
 * Provides a UI for searching the Lazy-MCP Tool Index.
 */
export default function IndexPage() {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<IndexedTool[]>([]);
    const [total, setTotal] = useState(0);
    const [stats, setStats] = useState<IndexStats | null>(null);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    // Debounce search
    useEffect(() => {
        const timer = setTimeout(() => {
            handleSearch();
        }, 300);
        return () => clearTimeout(timer);
    }, [query]);

    useEffect(() => {
        fetchStats();
    }, []);

    const fetchStats = async () => {
        try {
            const data = await indexClient.getStats();
            setStats(data);
        } catch (e) {
            console.error("Failed to fetch stats", e);
        }
    };

    const handleSearch = async () => {
        setLoading(true);
        try {
            const res = await indexClient.search(query);
            setResults(res.tools);
            setTotal(res.total);
            fetchStats(); // Refresh stats after search
        } catch (e) {
            console.error("Search failed", e);
            toast({ title: "Search Failed", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    const handleInstall = (tool: IndexedTool) => {
        // Logic to trigger installation (e.g. open install dialog)
        // For now, just a toast
        toast({ title: "Installation Started", description: `Installing ${tool.name}... (Simulated)` });
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Tool Index</h2>
                    <p className="text-muted-foreground">
                        Discover and lazy-load tools from the global catalog.
                    </p>
                </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Indexed</CardTitle>
                        <Database className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.totalTools || 0}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Searches</CardTitle>
                        <Search className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.totalSearches || 0}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Cache Hits</CardTitle>
                        <Activity className="h-4 w-4 text-green-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.hits || 0}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Cache Misses</CardTitle>
                        <AlertCircle className="h-4 w-4 text-amber-500" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.misses || 0}</div>
                    </CardContent>
                </Card>
            </div>

            <div className="flex items-center space-x-2">
                <div className="relative flex-1">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search tools (e.g. 'weather', 'finance')..."
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        className="pl-8"
                    />
                </div>
            </div>

            <div className="rounded-md border flex-1 overflow-auto bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Category</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead className="text-right">Action</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {results.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                    {loading ? "Searching..." : "No tools found in the index."}
                                </TableCell>
                            </TableRow>
                        ) : (
                            results.map((tool) => (
                                <TableRow key={tool.name}>
                                    <TableCell className="font-medium">
                                        <div className="flex flex-col">
                                            <span>{tool.name}</span>
                                            <div className="flex gap-1 mt-1">
                                                {tool.tags.map(tag => (
                                                    <Badge key={tag} variant="outline" className="text-[10px] px-1 py-0">{tag}</Badge>
                                                ))}
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell>{tool.category}</TableCell>
                                    <TableCell className="max-w-[400px] truncate" title={tool.description}>{tool.description}</TableCell>
                                    <TableCell>
                                        {tool.installed ? (
                                            <Badge variant="default" className="bg-green-500">Installed</Badge>
                                        ) : (
                                            <Badge variant="secondary">Lazy</Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        {!tool.installed && (
                                            <Button size="sm" variant="outline" onClick={() => handleInstall(tool)}>
                                                <Download className="mr-2 h-4 w-4" /> Install
                                            </Button>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
