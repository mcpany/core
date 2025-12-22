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

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

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
    <div className={cn("pb-12 bg-[#2c3e50] text-white transition-all duration-300", collapsed ? "w-16" : "w-64", className)}>
      <div className="space-y-4 py-4">
        <div className="px-4 py-2 flex items-center justify-between">
          {!collapsed && (
             <div className="flex items-center gap-2 font-bold text-xl truncate">
                <div className="bg-blue-500 rounded-full p-1">
                    <LayoutGrid className="h-5 w-5 text-white" />
                </div>
                <span>MCP Any</span>
             </div>
          )}
           {collapsed && (
             <div className="mx-auto bg-blue-500 rounded-full p-1">
                <LayoutGrid className="h-5 w-5 text-white" />
             </div>
           )}
           {/* Mobile toggle or specialized toggle could go here, for now just a simple logic if needed */}
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
                    "flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-white/10 transition-colors",
                    isActive ? "bg-blue-600 text-white" : "transparent text-gray-300",
                    collapsed ? "justify-center" : "justify-start"
                  )}
                  title={item.title}
                >
                  <item.icon className={cn("h-5 w-5", collapsed ? "mr-0" : "mr-2")} />
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
                className="w-full text-gray-400 hover:bg-white/10 hover:text-white"
                onClick={() => setCollapsed(!collapsed)}
            >
                <Menu className="h-5 w-5" />
            </Button>
       </div>
    </div>
  );
}
}
