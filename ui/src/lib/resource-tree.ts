/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceDefinition } from "@/lib/client";

export interface TreeNode {
    id: string;
    name: string;
    fullPath: string; // The full path or URI prefix
    type: "folder" | "file";
    children?: TreeNode[];
    resource?: ResourceDefinition; // Only for files
}

/**
 * Builds a hierarchical tree from a list of resources.
 * Handles `file://` URIs by splitting paths.
 * Groups other schemes by their authority/path.
 *
 * @param resources List of flat resources
 * @returns Root TreeNode containing the hierarchy
 */
export function buildResourceTree(resources: ResourceDefinition[]): TreeNode[] {
    const rootChildren: TreeNode[] = [];

    // Helper to find or create a folder node in a list of siblings
    const findOrCreateFolder = (siblings: TreeNode[], name: string, fullPath: string): TreeNode => {
        let node = siblings.find(n => n.name === name && n.type === "folder");
        if (!node) {
            node = {
                id: fullPath,
                name: name,
                fullPath: fullPath,
                type: "folder",
                children: []
            };
            siblings.push(node);
            // Sort folders alphabetically
            siblings.sort((a, b) => {
                if (a.type !== b.type) return a.type === "folder" ? -1 : 1;
                return a.name.localeCompare(b.name);
            });
        }
        return node;
    };

    resources.forEach(res => {
        // Simple URI parsing
        // We handle file:// specially
        if (res.uri.startsWith("file://")) {
            // Remove scheme
            const path = res.uri.replace("file://", "");
            // Split by '/'
            // Filter empty parts (e.g. leading slash results in empty first element)
            const parts = path.split("/").filter(p => p);

            let currentLevel = rootChildren;
            let currentPath = "file://";

            // Traverse path segments
            for (let i = 0; i < parts.length; i++) {
                const part = parts[i];
                const isLast = i === parts.length - 1;

                if (isLast) {
                    // File node
                    currentLevel.push({
                        id: res.uri,
                        name: part, // Use filename from URI, or res.name? Usually matching.
                        fullPath: res.uri,
                        type: "file",
                        resource: res
                    });
                } else {
                    // Folder node
                    currentPath += "/" + part;
                    const folder = findOrCreateFolder(currentLevel, part, currentPath);
                    if (!folder.children) folder.children = [];
                    currentLevel = folder.children;
                }
            }
        } else {
            // Handle other schemes (e.g. postgres://db/users/schema)
            // We can try a generic split by '/' after the scheme
            const schemeMatch = res.uri.match(/^([a-z][a-z0-9+.-]*):\/\/(.*)/i);
            if (schemeMatch) {
                const scheme = schemeMatch[1];
                const rest = schemeMatch[2];
                const parts = rest.split("/").filter(p => p);

                // Group by Scheme first
                let currentLevel = rootChildren;
                let currentPath = scheme + "://";

                // Add scheme folder if not exists
                const schemeFolder = findOrCreateFolder(currentLevel, scheme + "://", currentPath);
                if (!schemeFolder.children) schemeFolder.children = [];
                currentLevel = schemeFolder.children;

                // Traverse segments
                for (let i = 0; i < parts.length; i++) {
                    const part = parts[i];
                    const isLast = i === parts.length - 1;

                    if (isLast) {
                        currentLevel.push({
                            id: res.uri,
                            name: part,
                            fullPath: res.uri,
                            type: "file",
                            resource: res
                        });
                    } else {
                        currentPath += (currentPath.endsWith("://") ? "" : "/") + part;
                        const folder = findOrCreateFolder(currentLevel, part, currentPath);
                        if (!folder.children) folder.children = [];
                        currentLevel = folder.children;
                    }
                }
            } else {
                // Fallback for opaque URIs -> Flat list at root
                rootChildren.push({
                    id: res.uri,
                    name: res.name || res.uri,
                    fullPath: res.uri,
                    type: "file",
                    resource: res
                });
            }
        }
    });

    // Final sort of root
    rootChildren.sort((a, b) => {
        if (a.type !== b.type) return a.type === "folder" ? -1 : 1;
        return a.name.localeCompare(b.name);
    });

    return rootChildren;
}

/**
 * Flattens a node and its children for searching or linear traversal if needed.
 */
export function flattenTree(nodes: TreeNode[]): TreeNode[] {
    let result: TreeNode[] = [];
    for (const node of nodes) {
        result.push(node);
        if (node.children) {
            result = result.concat(flattenTree(node.children));
        }
    }
    return result;
}
