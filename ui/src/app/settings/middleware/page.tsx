/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";

interface Middleware {
    name: string;
    priority: number;
    disabled: boolean;
}

export default function MiddlewarePage() {
    const [middleware, setMiddleware] = useState<Middleware[]>([]);

    useEffect(() => {
        async function fetchMiddleware() {
            const res = await fetch("/api/settings/middleware");
            if (res.ok) {
                setMiddleware(await res.json());
            }
        }
        fetchMiddleware();
    }, []);

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <h2 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h2>
            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle>Active Middleware</CardTitle>
                    <CardDescription>Visual management of the request processing pipeline.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Priority</TableHead>
                                <TableHead>Name</TableHead>
                                <TableHead>Status</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {middleware.sort((a,b) => a.priority - b.priority).map((mw) => (
                                <TableRow key={mw.name}>
                                    <TableCell>{mw.priority}</TableCell>
                                    <TableCell className="font-medium">{mw.name}</TableCell>
                                    <TableCell>
                                        <div className="flex items-center space-x-2">
                                            <Switch checked={!mw.disabled} />
                                            <span>{!mw.disabled ? "Active" : "Disabled"}</span>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
