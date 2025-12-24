/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { NavItem } from "./nav-item"
import { LayoutDashboard, Server, Wrench, Database, MessageSquare, Settings, Activity, GitBranch, Search } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"

export function Sidebar() {
  const triggerCommandMenu = () => {
    const event = new KeyboardEvent('keydown', {
      key: 'k',
      metaKey: true,
      bubbles: true
    });
    document.dispatchEvent(event);
  };

  return (
    <div className="hidden border-r bg-muted/40 md:block min-h-screen w-[220px] lg:w-[280px]">
      <div className="flex h-full max-h-screen flex-col gap-2">
        <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6">
          <Link href="/" className="flex items-center gap-2 font-semibold">
            <Activity className="h-6 w-6 text-primary" />
            <span className="">MCP Any</span>
          </Link>
        </div>
        <div className="px-4 py-2">
           <Button
            variant="outline"
            className="w-full justify-start text-muted-foreground bg-background/50 h-9 px-2"
            onClick={triggerCommandMenu}
           >
              <Search className="mr-2 h-4 w-4" />
              <span className="text-sm">Search...</span>
              <kbd className="pointer-events-none ml-auto inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
                <span className="text-xs">⌘</span>K
              </kbd>
           </Button>
        </div>
        <div className="flex-1">
          <nav className="grid items-start px-2 text-sm font-medium lg:px-4">
            <NavItem href="/" icon={LayoutDashboard} title="Dashboard" isActive={true} />
            <NavItem href="/services" icon={Server} title="Services" />
            <NavItem href="/tools" icon={Wrench} title="Tools" />
            <NavItem href="/resources" icon={Database} title="Resources" />
            <NavItem href="/prompts" icon={MessageSquare} title="Prompts" />
            <div className="my-2 border-t" />
             <NavItem href="/middleware" icon={GitBranch} title="Middleware" />
            <NavItem href="/settings" icon={Settings} title="Settings" />
          </nav>
        </div>
        <div className="mt-auto p-4">
            <div className="text-xs text-muted-foreground text-center">
                v1.0.0
            </div>
        </div>
      </div>
    </div>
  )
}
