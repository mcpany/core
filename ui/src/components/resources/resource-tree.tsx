/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { TreeNode } from "@/lib/resource-tree";
import { Folder, FileText, ChevronRight, ChevronDown, Database, Server } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface ResourceTreeProps {
    data: TreeNode[];
    onSelect: (node: TreeNode) => void;
    selectedId?: string;
    level?: number;
}

/**
 * Recursive tree component for navigating resources.
 */
export function ResourceTree({ data, onSelect, selectedId, level = 0 }: ResourceTreeProps) {
    return (
        <ul className={cn("space-y-0.5", level > 0 && "ml-3")}>
            {data.map((node) => (
                <TreeNodeItem
                    key={node.id}
                    node={node}
                    onSelect={onSelect}
                    selectedId={selectedId}
                    level={level}
                />
            ))}
        </ul>
    );
}

function TreeNodeItem({ node, onSelect, selectedId, level }: { node: TreeNode, onSelect: (n: TreeNode) => void, selectedId?: string, level: number }) {
    const [isExpanded, setIsExpanded] = useState(true); // Default expanded for now
    const hasChildren = node.children && node.children.length > 0;
    const isSelected = selectedId === node.id || (node.type === "folder" && selectedId === node.fullPath);

    const handleToggle = (e: React.MouseEvent) => {
        e.stopPropagation();
        setIsExpanded(!isExpanded);
    };

    const handleClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        onSelect(node);
        if (node.type === "folder") {
            // Optional: Toggle expand on click too?
            // setIsExpanded(!isExpanded);
            // Better to keep selection and expansion separate usually, but for folders it makes sense to ensure expanded.
            if (!isExpanded) setIsExpanded(true);
        }
    };

    const getIcon = () => {
        if (node.type === "folder") {
            if (node.name.includes("://")) return <Server className="h-4 w-4 text-blue-500" />;
            if (node.name === "db" || node.fullPath.includes("postgres")) return <Database className="h-4 w-4 text-orange-500" />;
            return <Folder className={cn("h-4 w-4", isSelected ? "text-primary" : "text-muted-foreground")} />;
        }
        return <FileText className="h-4 w-4 text-muted-foreground" />;
    };

    return (
        <li>
            <div
                className={cn(
                    "flex items-center gap-1.5 py-1 px-2 rounded-md cursor-pointer transition-colors text-sm hover:bg-accent/50",
                    isSelected && "bg-accent text-accent-foreground font-medium"
                )}
                onClick={handleClick}
                style={{ paddingLeft: `${level * 4 + 8}px` }}
            >
                {hasChildren ? (
                    <div
                        role="button"
                        className="p-0.5 hover:bg-muted rounded text-muted-foreground"
                        onClick={handleToggle}
                    >
                        {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                    </div>
                ) : (
                    <div className="w-4" /> // Spacer
                )}

                {getIcon()}
                <span className="truncate">{node.name}</span>
            </div>

            {hasChildren && isExpanded && (
                <ResourceTree
                    data={node.children!}
                    onSelect={onSelect}
                    selectedId={selectedId}
                    level={level + 1}
                />
            )}
        </li>
    );
}
