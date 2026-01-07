
"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient, ToolDefinition } from "@/lib/client";
import { GlassCard } from "@/components/layout/glass-card";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { StatusBadge } from "@/components/layout/status-badge";
import { Button } from "@/components/ui/button";
import { Play } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";

export default function ToolsPage() {
    const [tools, setTools] = useState<ToolDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedTool, setSelectedTool] = useState<ToolDefinition | null>(null);
    const [executionArgs, setExecutionArgs] = useState("{}");
    const [executionResult, setExecutionResult] = useState<string>("");

    useEffect(() => {
        apiClient.listTools().then(res => {
            // Adjust based on actual API response structure (array vs object wrapper)
            setTools(Array.isArray(res) ? res : (res.tools || []));
            setLoading(false);
        }).catch(err => {
            console.error("Failed to load tools", err);
            setLoading(false);
        });
    }, []);

    const handleExecute = async () => {
        if (!selectedTool) return;
        try {
            const args = JSON.parse(executionArgs);
            const res = await apiClient.executeTool({
                tool_name: selectedTool.name,
                arguments: args
            });
            setExecutionResult(JSON.stringify(res, null, 2));
        } catch (e: any) {
            setExecutionResult(`Error: ${e.message}`);
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
            </div>
            <GlassCard>
                <CardHeader>
                    <CardTitle>Available Tools</CardTitle>
                    <CardDescription>Tools exposed by connected MCP servers.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Description</TableHead>
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={3} className="text-center">Loading...</TableCell>
                                </TableRow>
                            ) : tools.map((tool) => (
                                <TableRow key={tool.name}>
                                    <TableCell className="font-medium">{tool.name}</TableCell>
                                    <TableCell className="max-w-md truncate" title={tool.description}>{tool.description}</TableCell>
                                    <TableCell>
                                        <Button size="sm" variant="outline" onClick={() => {
                                            setSelectedTool(tool);
                                            setExecutionArgs("{}"); // Reset args or try to generate schema template
                                            setExecutionResult("");
                                        }}>
                                            <Play className="mr-2 h-3 w-3" /> Test
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </GlassCard>

            <Dialog open={!!selectedTool} onOpenChange={(open) => !open && setSelectedTool(null)}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>Execute Tool: {selectedTool?.name}</DialogTitle>
                        <DialogDescription>
                            {selectedTool?.description}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label>Arguments (JSON)</Label>
                            <Textarea
                                value={executionArgs}
                                onChange={(e) => setExecutionArgs(e.target.value)}
                                className="font-mono"
                                rows={5}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label>Result</Label>
                            <div className="rounded-md bg-muted p-4 font-mono text-xs whitespace-pre-wrap h-[200px] overflow-auto">
                                {executionResult || "Waiting for execution..."}
                            </div>
                        </div>
                    </div>
                    <div className="flex justify-end">
                        <Button onClick={handleExecute}>Execute</Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
