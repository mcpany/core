/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import {
    Cpu,
    Server,
    Users,
    Zap,
    Link as LinkIcon,
    Layers,
    Webhook,
    Database,
    MessageSquare,
    Activity,
    CheckCircle2,
    XCircle,
    AlertTriangle,
    Info
} from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export function NetworkLegend() {
    return (
        <Card className="w-full h-full border-none shadow-none bg-transparent">
            <CardHeader className="p-4 pb-2">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <Info className="h-4 w-4" /> Legend
                </CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-0 space-y-6">
                <div className="space-y-3">
                    <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Node Types</h4>
                    <div className="grid grid-cols-2 gap-2 text-xs">
                        <div className="flex items-center gap-2">
                            <Cpu className="h-4 w-4 text-blue-500 dark:text-blue-400" />
                            <span>Core System</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Server className="h-4 w-4 text-indigo-500 dark:text-indigo-400" />
                            <span>Service</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Users className="h-4 w-4 text-green-500 dark:text-green-400" />
                            <span>Client</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Zap className="h-4 w-4 text-amber-500 dark:text-amber-400" />
                            <span>Tool</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Layers className="h-4 w-4 text-orange-500 dark:text-orange-400" />
                            <span>Middleware</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Webhook className="h-4 w-4 text-pink-500 dark:text-pink-400" />
                            <span>Webhook</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Database className="h-4 w-4 text-cyan-500 dark:text-cyan-400" />
                            <span>Resource</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <MessageSquare className="h-4 w-4 text-purple-500 dark:text-purple-400" />
                            <span>Prompt</span>
                        </div>
                         <div className="flex items-center gap-2">
                            <LinkIcon className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                            <span>API Call</span>
                        </div>
                    </div>
                </div>

                <div className="space-y-3">
                    <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Status</h4>
                    <div className="space-y-2 text-xs">
                        <div className="flex items-center gap-2">
                            <CheckCircle2 className="h-4 w-4 text-green-500" />
                            <span>Active / Healthy</span>
                        </div>
                        <div className="flex items-center gap-2">
                             <XCircle className="h-4 w-4 text-slate-400" />
                            <span>Inactive / Disabled</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <AlertTriangle className="h-4 w-4 text-destructive" />
                            <span>Error / Failing</span>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
