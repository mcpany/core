/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { StackEditor } from "@/components/stacks/stack-editor";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RefreshCcw, Activity, PlayCircle, StopCircle, Trash2, Box } from "lucide-react";
import { use } from "react";

// Placeholder for StackStatus if we want a separate component
function StackStatus({ stackId }: { stackId: string }) {
    // Mock data for runtime status
    const services = [
        { name: "weather-service", status: "Running", uptime: "2d 4h", cpu: "0.5%", mem: "128MB" },
        { name: "local-files", status: "Stopped", uptime: "-", cpu: "0%", mem: "0MB" },
    ];

    return (
        <div className="space-y-4">
             <div className="flex items-center gap-4">
                 <Card className="flex-1">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Total Services</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{services.length}</div>
                    </CardContent>
                 </Card>
                 <Card className="flex-1">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Running</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-green-500">
                            {services.filter(s => s.status === "Running").length}
                        </div>
                    </CardContent>
                 </Card>
                 <Card className="flex-1">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Errors</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-muted-foreground">0</div>
                    </CardContent>
                 </Card>
             </div>

             <Card>
                <CardHeader>
                    <CardTitle className="text-lg">Runtime Status</CardTitle>
                    <CardDescription>Live status of services defined in this stack.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <div className="grid grid-cols-6 gap-4 p-4 border-b font-medium text-sm bg-muted/50">
                            <div className="col-span-2">Service Name</div>
                            <div>Status</div>
                            <div>Uptime</div>
                            <div>CPU</div>
                            <div className="text-right">Actions</div>
                        </div>
                        {services.map((svc) => (
                            <div key={svc.name} className="grid grid-cols-6 gap-4 p-4 items-center text-sm border-b last:border-0 hover:bg-muted/10 transition-colors">
                                <div className="col-span-2 font-mono flex items-center gap-2">
                                    <Box className="h-4 w-4 text-muted-foreground" />
                                    {svc.name}
                                </div>
                                <div>
                                    <Badge variant={svc.status === "Running" ? "default" : "secondary"} className={svc.status === "Running" ? "bg-green-500 hover:bg-green-600" : ""}>
                                        {svc.status}
                                    </Badge>
                                </div>
                                <div className="text-muted-foreground">{svc.uptime}</div>
                                <div className="text-muted-foreground font-mono text-xs">{svc.cpu} / {svc.mem}</div>
                                <div className="flex justify-end gap-2">
                                    {svc.status === "Running" ? (
                                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" title="Stop">
                                            <StopCircle className="h-4 w-4" />
                                        </Button>
                                    ) : (
                                        <Button variant="ghost" size="icon" className="h-8 w-8 text-green-500" title="Start">
                                            <PlayCircle className="h-4 w-4" />
                                        </Button>
                                    )}
                                    <Button variant="ghost" size="icon" className="h-8 w-8" title="Logs">
                                        <Activity className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
                </CardContent>
             </Card>
        </div>
    );
}

export default function StackDetailPage({ params }: { params: Promise<{ stackId: string }> }) {
    const resolvedParams = use(params);
    const [activeTab, setActiveTab] = useState("editor");

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div className="flex flex-col gap-1">
                     <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        {resolvedParams.stackId}
                        <Badge variant="outline" className="text-xs font-normal">Stack</Badge>
                     </h2>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm">
                        <RefreshCcw className="mr-2 h-4 w-4" /> Refresh
                    </Button>
                    {activeTab === "editor" && (
                         <Button size="sm" onClick={() => {
                             // This button duplicates the Save inside Editor,
                             // maybe just let Editor handle it or use a global context/ref
                         }} className="hidden">
                            Deploy Stack
                        </Button>
                    )}
                </div>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col space-y-4">
                <TabsList>
                    <TabsTrigger value="status">Overview & Status</TabsTrigger>
                    <TabsTrigger value="editor">Editor</TabsTrigger>
                </TabsList>

                <TabsContent value="status" className="flex-1">
                     <StackStatus stackId={resolvedParams.stackId} />
                </TabsContent>

                <TabsContent value="editor" className="flex-1 flex flex-col h-full min-h-0">
                    <StackEditor stackId={resolvedParams.stackId} />
                </TabsContent>
            </Tabs>
        </div>
    );
}
