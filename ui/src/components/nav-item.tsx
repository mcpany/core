/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { cn } from "@/lib/utils"
import { LucideIcon } from "lucide-react"
import { Link } from 'react-router-dom'

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
 * NavItem executes the NavItem logic.
 *
 * Summary: Executes the NavItem logic.
 *
 * @param { href - The { href parameter.
 * @param icon - The icon parameter.
 * @param title - The title parameter.
 * @param isActive } - The isActive } parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function NavItem({ href, icon: Icon, title, isActive }: NavItemProps) {
  return (
    <Link
      to={href}
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
