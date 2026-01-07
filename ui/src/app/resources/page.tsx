
"use client";

import { useState, useEffect } from "react";
import { apiClient, ResourceDefinition } from "@/lib/client";
import { GlassCard } from "@/components/layout/glass-card";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { StatusBadge } from "@/components/layout/status-badge";
import { Button } from "@/components/ui/button";
import { Copy, ExternalLink } from "lucide-react";

export default function ResourcesPage() {
    const [resources, setResources] = useState<ResourceDefinition[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        apiClient.listResources().then(res => {
            setResources(Array.isArray(res) ? res : (res.resources || []));
            setLoading(false);
        }).catch(err => {
            console.error("Failed to load resources", err);
            setLoading(false);
        });
    }, []);

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
            </div>
            <GlassCard>
                <CardHeader>
                    <CardTitle>Resources</CardTitle>
                    <CardDescription>Data resources available to LLMs.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>URI</TableHead>
                                <TableHead>MIME Type</TableHead>
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={4} className="text-center">Loading...</TableCell>
                                </TableRow>
                            ) : resources.map((res) => (
                                <TableRow key={res.uri}>
                                    <TableCell className="font-medium">{res.name}</TableCell>
                                    <TableCell className="font-mono text-xs">{res.uri}</TableCell>
                                    <TableCell>{res.mimeType || "text/plain"}</TableCell>
                                    <TableCell>
                                        <Button size="icon" variant="ghost" title="Copy URI" onClick={() => navigator.clipboard.writeText(res.uri)}>
                                            <Copy className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </GlassCard>
        </div>
    );
}
