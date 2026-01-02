/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutGrid,
  Layers,
  Box,
  Settings,
  Menu,
  Home
} from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";

type SidebarProps = React.HTMLAttributes<HTMLDivElement>;

export function Sidebar({ className }: SidebarProps) {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);

  const navItems = [
    {
      title: "Home",
      href: "/dashboard", // Assuming we might move Home there, or logic to redirect "/"
      icon: Home,
      matches: ["/dashboard", "/"] // Matches "/" too for now
    },
    {
      title: "Stacks",
      href: "/stacks",
      icon: Layers,
      matches: ["/stacks"]
    },
    {
      title: "Services",
      href: "/services",
      icon: Box,
      matches: ["/services"]
    },
    {
      title: "Settings",
      href: "/settings",
      icon: Settings,
      matches: ["/settings"]
    },
  ];

  return (
    <div className={cn("pb-12 bg-sidebar-background text-sidebar-foreground transition-all duration-300 border-r border-border/10", collapsed ? "w-16" : "w-64", className)}>
      <div className="space-y-4 py-4">
        <div className="px-4 py-3 flex items-center justify-between">
          {!collapsed && (
             <div className="flex items-center gap-3 font-semibold text-lg tracking-tight">
                <div className="bg-primary/20 p-2 rounded-lg">
                    <LayoutGrid className="h-5 w-5 text-primary" />
                </div>
                <span>MCP Any</span>
             </div>
          )}
           {collapsed && (
             <div className="mx-auto bg-primary/20 p-2 rounded-lg">
                <LayoutGrid className="h-5 w-5 text-primary" />
             </div>
           )}
        </div>
        <div className="px-3 py-2">
          <div className="space-y-1">
            {navItems.map((item) => {
               const isActive = item.matches.some(m => pathname === m || (m !== '/' && pathname.startsWith(m)));
               return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "flex items-center rounded-lg px-3 py-2.5 text-sm font-medium transition-all group relative",
                    isActive ? "text-primary bg-primary/10" : "text-muted-foreground hover:text-foreground hover:bg-white/5",
                    collapsed ? "justify-center" : "justify-start"
                  )}
                  title={item.title}
                >
                  {isActive && !collapsed && <div className="absolute left-0 top-1/2 -translate-y-1/2 h-8 w-1 bg-primary rounded-r-md" />}
                  <item.icon className={cn("h-5 w-5", collapsed ? "mr-0" : "mr-3")} />
                  {!collapsed && <span>{item.title}</span>}
                </Link>
              );
            })}
          </div>
        </div>
      </div>
       <div className="absolute bottom-4 left-0 right-0 px-3">
            <Button
                variant="ghost"
                size="icon"
                className="w-full text-muted-foreground hover:text-foreground hover:bg-white/5"
                onClick={() => setCollapsed(!collapsed)}
            >
                <Menu className="h-5 w-5" />
            </Button>
       </div>
    </div>
  );
}
