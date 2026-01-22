/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Fragment } from "react";
import { ChevronRight, ChevronDown, Home } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

/**
 * Represents a single item in the breadcrumb navigation.
 */
export interface BreadcrumbItem {
    /** The label to display for the breadcrumb. */
    label: string;
    /** The URL link for the breadcrumb. */
    href: string;
    /** Optional list of sibling items for navigation. */
    siblings?: { label: string; href: string }[];
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
 *
 * @param className - The className.
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
                             <ChevronRight className="size-4 text-muted-foreground/50" />
                             <div className="flex items-center gap-0.5">
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
                                {item.siblings && item.siblings.length > 0 && (
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <button className="p-0.5 rounded-sm hover:bg-muted focus:outline-none transition-colors">
                                                <ChevronDown className="size-3 text-muted-foreground/70" />
                                                <span className="sr-only">More options</span>
                                            </button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="start" className="max-h-[300px] overflow-y-auto">
                                            {item.siblings.map((sibling) => (
                                                <DropdownMenuItem key={sibling.href} asChild>
                                                    <Link href={sibling.href} className={cn("cursor-pointer", sibling.href === item.href && "font-semibold bg-accent")}>
                                                        {sibling.label}
                                                    </Link>
                                                </DropdownMenuItem>
                                            ))}
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                )}
                            </div>
                        </li>
                    </Fragment>
                ))}
            </ol>
        </nav>
    );
}
