/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Fragment } from "react";
import { ChevronRight, Home } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Represents a single item in the breadcrumb navigation.
 */
export interface BreadcrumbItem {
    /** The label to display for the breadcrumb. */
    label: string;
    /** The URL link for the breadcrumb. */
    href: string;
}

/**
 * Props for the Breadcrumbs component.
 */
interface BreadcrumbsProps {
    /** The list of breadcrumb items to display. */
    items: BreadcrumbItem[];
    /** Optional CSS class names. */
    className?: string;
}

/**
 * Breadcrumbs navigation component.
 */
export function Breadcrumbs({ items, className }: BreadcrumbsProps) {
    return (
        <nav aria-label="Breadcrumb" className={cn("w-full max-w-6xl mb-4", className)}>
            <ol className="flex items-center gap-1 text-sm text-muted-foreground">
                <li>
                    <Link href="/" className="flex items-center gap-1.5 font-semibold text-foreground/80 hover:text-foreground transition-colors">
                        <Home className="size-4" />
                        <span className="sr-only">Home</span>
                    </Link>
                </li>
                {items.map((item, index) => (
                    <Fragment key={item.href}>
                        <li className="flex items-center gap-1">
                             <ChevronRight className="size-4" />
                            <Link
                                href={item.href}
                                aria-current={index === items.length - 1 ? "page" : undefined}
                                className={cn(
                                    "font-medium transition-colors hover:text-foreground",
                                    index === items.length - 1 ? "text-foreground" : "text-foreground/80"
                                )}
                            >
                                {item.label}
                            </Link>
                        </li>
                    </Fragment>
                ))}
            </ol>
        </nav>
    );
}
