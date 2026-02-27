/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { TreeNode } from "@/lib/resource-tree";
import { ChevronRight, Home } from "lucide-react";
import { Button } from "@/components/ui/button";

interface ResourceBreadcrumbProps {
    path: TreeNode[];
    onNavigate: (node: TreeNode | null) => void; // null for root
}

/**
 * Breadcrumb navigation for resource explorer.
 */
export function ResourceBreadcrumb({ path, onNavigate }: ResourceBreadcrumbProps) {
    return (
        <nav className="flex items-center text-sm text-muted-foreground overflow-hidden">
            <Button
                variant="ghost"
                size="sm"
                className="h-6 px-2 text-muted-foreground hover:text-foreground"
                onClick={() => onNavigate(null)}
            >
                <Home className="h-4 w-4" />
            </Button>

            {path.map((node, index) => (
                <div key={node.id} className="flex items-center">
                    <ChevronRight className="h-4 w-4 text-muted-foreground/50 mx-0.5" />
                    <Button
                        variant="ghost"
                        size="sm"
                        className={`h-6 px-2 ${index === path.length - 1 ? "font-medium text-foreground" : "text-muted-foreground hover:text-foreground"}`}
                        onClick={() => onNavigate(node)}
                    >
                        {node.name}
                    </Button>
                </div>
            ))}
        </nav>
    );
}
