/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { PlusCircle, Terminal, Lock, Network, ArrowRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

const actions = [
    {
        title: "Register Service",
        href: "/upstream-services",
        icon: PlusCircle,
        description: "Connect a new API or database",
        color: "text-blue-500",
        bgColor: "bg-blue-500/10"
    },
    {
        title: "Playground",
        href: "/playground",
        icon: Terminal,
        description: "Test tools interactively",
        color: "text-green-500"
    },
    {
        title: "Manage Secrets",
        href: "/secrets",
        icon: Lock,
        description: "Secure your credentials",
        color: "text-amber-500"
    },
    {
        title: "Network Map",
        href: "/network",
        icon: Network,
        description: "View system topology",
        color: "text-purple-500"
    }
];

/**
 * A widget that provides quick access links to common actions and pages.
 *
 * @returns The rendered widget component.
 */
export function QuickActionsWidget() {
    return (
        <Card className="h-full flex flex-col backdrop-blur-xl bg-background/60 border border-white/20 shadow-sm hover:shadow-md transition-all duration-300">
            <CardHeader className="pb-3">
                <CardTitle className="text-lg font-medium">Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="grid grid-cols-1 gap-3 flex-1">
                {actions.map((action) => (
                    <Link key={action.href} href={action.href} className="group block">
                        <div className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors">
                            <div className="flex items-center gap-3">
                                <div className={`p-2 rounded-full bg-background border ${action.color} bg-opacity-10`}>
                                    <action.icon className={`h-4 w-4 ${action.color}`} />
                                </div>
                                <div>
                                    <div className="font-medium text-sm group-hover:text-primary transition-colors">{action.title}</div>
                                    <div className="text-xs text-muted-foreground">{action.description}</div>
                                </div>
                            </div>
                            <ArrowRight className="h-4 w-4 text-muted-foreground opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all duration-200" />
                        </div>
                    </Link>
                ))}
            </CardContent>
        </Card>
    );
}
