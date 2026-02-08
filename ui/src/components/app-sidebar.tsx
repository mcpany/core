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
  Key,
  Database,
  User,
  Users,
  ChevronsUpDown,
  LogOut,
  Layers,
  ShoppingBag,
  ShieldCheck,
  Zap,
  ClipboardCheck,
  Bug
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
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { useUser } from "@/components/user-context"


const platformItems = [
  {
    title: "Dashboard",
    url: "/",
    icon: LayoutDashboard,
  },
  {
    title: "Network Graph",
    url: "/network",
    icon: Network,
  },
  {
    title: "Diagnostics",
    url: "/diagnostics",
    icon: Activity,
  },
  {
    title: "Live Logs",
    url: "/logs",
    icon: Terminal,
  },
  {
    title: "Audit Logs",
    url: "/audit",
    icon: ClipboardCheck,
  },
  {
    title: "Traces",
    url: "/traces",
    icon: Activity,
  },
  {
    title: "Alerts",
    url: "/alerts",
    icon: Activity,
  },
  {
    title: "Stacks",
    url: "/stacks",
    icon: Layers,
  },
  {
    title: "Analytics",
    url: "/stats",
    icon: Activity,
  },
  {
    title: "Marketplace",
    url: "/marketplace",
    icon: ShoppingBag,
  },
]

const devItems = [
  {
    title: "Playground",
    url: "/playground",
    icon: Bot,
  },
  {
    title: "Schema Validation",
    url: "/playground/schema",
    icon: ShieldCheck,
  },
  {
    title: "Tools",
    url: "/tools",
    icon: Wrench,
  },
  {
    title: "Resources",
    url: "/resources",
    icon: Database,
  },
  {
    title: "Prompts",
    url: "/prompts",
    icon: FileText,
  },
  {
    title: "Skills",
    url: "/skills",
    icon: Zap,
  },
  {
    title: "Config Validator",
    url: "/config-validator",
    icon: ShieldCheck,
  },
]

const configItems = [

  {
    title: "Upstream Services",
    url: "/upstream-services",
    icon: Server,
  },
  {
    title: "Profiles",
    url: "/profiles",
    icon: User,
  },
  {
    title: "Users",

    url: "/users",
    icon: Users,
  },
  {
    title: "Credentials",
    url: "/credentials",
    icon: ShieldCheck,
  },
  {
    title: "Secrets Vault",
    url: "/secrets",
    icon: Key,
  },
  {
    title: "Settings",
    url: "/settings",
    icon: Settings,
  },
]

/**
 * The main application sidebar.
 * Displays navigation links and user profile menu.
 *
 * @returns {JSX.Element} The rendered sidebar component.
 */
export function AppSidebar() {
  const pathname = usePathname()
  const { user, login } = useUser()

  const isAdmin = user?.role === 'admin';

  // Filter items based on role
  // Regular users see Dashboard, Network Graph, Analytics, Marketplace for Platform?
  // User said: "Regular user, probably will not see... Live Logs/Traces"
  const filteredPlatformItems = platformItems.filter(item => {
    if (!isAdmin) {
        return !['Live Logs', 'Traces'].includes(item.title);
    }
    return true;
  });

  // User said: "probably will not see 'Configuration' section"
  // So we hide the whole config group if not admin?
  // "Regular user can only see and manage settings belong to their own copy of profile"
  // Maybe we keep Settings but hide Services, Users, Secrets?
  const filteredConfigItems = configItems.filter(item => {
      if (!isAdmin) {
          // Keep Settings, hide others?
          return item.title === 'Settings';
      }
      return true;
  });

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
          <div className="flex items-center gap-2 p-2">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <Network className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
                <span className="truncate font-semibold">MCP Any</span>
                <span className="truncate text-xs">Admin Console</span>
              </div>
          </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Platform</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {filteredPlatformItems.map((item) => (
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

        <SidebarGroup>
          <SidebarGroupLabel>Development</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {devItems.map((item) => (
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

        {/* Only show Configuration group if there are items to show */}
        {filteredConfigItems.length > 0 && (
            <SidebarGroup>
            <SidebarGroupLabel>Configuration</SidebarGroupLabel>
            <SidebarGroupContent>
                <SidebarMenu>
                {filteredConfigItems.map((item) => (
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
        )}
      </SidebarContent>

      <SidebarFooter className="mt-auto p-0 border-t">
        <SidebarMenu className="gap-0">
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="rounded-none h-14 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                >
                  <Avatar className="h-8 w-8 rounded-lg">
                    <AvatarImage src={user?.avatar} alt={user?.name} />
                    <AvatarFallback className="rounded-lg">AD</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
                    <span className="truncate font-semibold">{user?.name}</span>
                    <span className="truncate text-xs">{user?.email}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4 group-data-[collapsible=icon]:hidden" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                side="bottom"
                align="end"
                sideOffset={4}
              >
                <DropdownMenuLabel className="p-0 font-normal">
                  <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                    <Avatar className="h-8 w-8 rounded-lg">
                      <AvatarImage src={user?.avatar} alt={user?.name} />
                      <AvatarFallback className="rounded-lg">AD</AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-semibold">{user?.name}</span>
                      <span className="truncate text-xs">{user?.email}</span>
                    </div>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => login(user?.role === 'admin' ? 'viewer' : 'admin')}>
                   <User className="mr-2 h-4 w-4" />
                   Switch Role (Demo)
                </DropdownMenuItem>
                <DropdownMenuItem>
                   <Settings className="mr-2 h-4 w-4" />
                   Preferences
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <LogOut className="mr-2 h-4 w-4" />
                  Log out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
