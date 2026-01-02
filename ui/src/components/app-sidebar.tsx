/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import {
  LayoutDashboard,
  Server,
  Network,
  Terminal,
  FileText,
  Wrench,
  Bot,
  Settings,
  Activity,
} from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

// Menu items.
const items = [
  {
    title: "Dashboard",
    url: "/",
    icon: LayoutDashboard,
  },
  {
    title: "Services",
    url: "/services",
    icon: Server,
  },
  {
    title: "Network",
    url: "/network",
    icon: Network,
  },
  {
    title: "Logs",
    url: "/logs",
    icon: Terminal,
  },
  {
    title: "Playground",
    url: "/playground",
    icon: Bot,
  },
  {
    title: "Prompts",
    url: "/prompts",
    icon: Terminal,
  },
  {
    title: "Resources",
    url: "/resources",
    icon: FileText,
  },
  {
    title: "Tools",
    url: "/tools",
    icon: Wrench,
  },
  {
    title: "Stats",
    url: "/stats",
    icon: Activity,
  },
  {
    title: "Settings",
    url: "/settings",
    icon: Settings,
  },
]

/**
 * The main application sidebar component.
 *
 * Renders the navigation menu with links to different sections of the application.
 * Highlights the active link based on the current pathname.
 */
export function AppSidebar() {
  const pathname = usePathname()

  return (
    <Sidebar collapsible="icon">
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Application</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild isActive={pathname === item.url} tooltip={item.title}>
                    <Link href={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}
