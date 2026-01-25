/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { RegisterServiceDialog } from "@/components/register-service-dialog";
import Link from "next/link";
import { Plus, ShoppingBag, Terminal, FileText, Lock, Zap, Server, LucideIcon } from "lucide-react";
import React from "react";

interface QuickActionButtonProps {
    icon: LucideIcon;
    label: string;
    href?: string;
    iconColorClass?: string;
    iconBgClass?: string;
    onClick?: () => void;
    asChild?: boolean;
    children?: React.ReactNode;
}

function QuickActionButton({ icon: Icon, label, href, iconColorClass = "text-primary", iconBgClass = "bg-primary/10 group-hover:bg-primary/20", asChild, children }: QuickActionButtonProps) {
    const content = (
        <>
            <div className={`${iconBgClass} p-2 rounded-full transition-colors`}>
                <Icon className={`h-5 w-5 ${iconColorClass}`} />
            </div>
            <span className="font-medium text-xs">{label}</span>
        </>
    );

    const buttonClass = "h-full flex flex-col items-center justify-center gap-2 p-2 hover:border-primary/50 hover:bg-primary/5 transition-all group whitespace-normal text-center h-24 w-full";

    if (href) {
        return (
            <Button variant="outline" className={buttonClass} asChild>
                <Link href={href}>
                    {content}
                </Link>
            </Button>
        );
    }

    if (children) {
        // If used as a trigger for a dialog, the dialog usually expects to clone/wrap the element.
        // But Button accepts children.
        return (
             <Button variant="outline" className={buttonClass}>
                {content}
            </Button>
        )
    }

    return (
        <Button variant="outline" className={buttonClass}>
            {content}
        </Button>
    );
}


/**
 * QuickActionsWidget component.
 * Displays a grid of quick access buttons for common tasks.
 * @returns The rendered component.
 */
export function QuickActionsWidget() {
    return (
        <Card className="h-full flex flex-col shadow-sm">
            <CardHeader className="pb-3 pt-4 px-4">
                <CardTitle className="text-base font-medium flex items-center gap-2">
                    <Zap className="h-4 w-4 text-primary" />
                    Quick Actions
                </CardTitle>
            </CardHeader>
            <CardContent className="flex-1 grid grid-cols-2 sm:grid-cols-3 gap-2 px-4 pb-4">
                <RegisterServiceDialog
                    trigger={
                        <Button variant="outline" className="h-full flex flex-col items-center justify-center gap-2 p-2 hover:border-primary/50 hover:bg-primary/5 transition-all group whitespace-normal text-center h-24 w-full">
                            <div className="bg-primary/10 p-2 rounded-full group-hover:bg-primary/20 transition-colors">
                                <Plus className="h-5 w-5 text-primary" />
                            </div>
                            <span className="font-medium text-xs">Add Service</span>
                        </Button>
                    }
                />

                <QuickActionButton
                    icon={ShoppingBag}
                    label="Marketplace"
                    href="/marketplace"
                    iconColorClass="text-purple-600 dark:text-purple-400"
                    iconBgClass="bg-purple-500/10 group-hover:bg-purple-500/20"
                />

                <QuickActionButton
                    icon={Terminal}
                    label="Playground"
                    href="/playground"
                    iconColorClass="text-orange-600 dark:text-orange-400"
                    iconBgClass="bg-orange-500/10 group-hover:bg-orange-500/20"
                />

                <QuickActionButton
                    icon={Server}
                    label="All Services"
                    href="/services"
                    iconColorClass="text-blue-600 dark:text-blue-400"
                    iconBgClass="bg-blue-500/10 group-hover:bg-blue-500/20"
                />

                <QuickActionButton
                    icon={FileText}
                    label="System Logs"
                    href="/logs"
                    iconColorClass="text-slate-600 dark:text-slate-400"
                    iconBgClass="bg-slate-500/10 group-hover:bg-slate-500/20"
                />

                <QuickActionButton
                    icon={Lock}
                    label="Secrets"
                    href="/secrets"
                    iconColorClass="text-red-600 dark:text-red-400"
                    iconBgClass="bg-red-500/10 group-hover:bg-red-500/20"
                />
            </CardContent>
        </Card>
    );
}
