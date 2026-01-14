/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { cn } from "@/lib/utils"
import { LucideIcon } from "lucide-react"
import Link from "next/link"

/**
 * Props for the NavItem component.
 */
interface NavItemProps {
  /** The URL to link to. */
  href: string
  /** The icon component to display. */
  icon: LucideIcon
  /** The title of the navigation item. */
  title: string
  /** Whether the item is currently active. */
  isActive?: boolean
}

/**
 * A navigation item for the sidebar or menu.
 */
export function NavItem({ href, icon: Icon, title, isActive }: NavItemProps) {
  return (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all hover:text-primary",
        isActive
          ? "bg-muted text-primary"
          : "text-muted-foreground"
      )}
    >
      <Icon className="h-4 w-4" />
      {title}
    </Link>
  )
}
