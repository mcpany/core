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
 * Intent: Document NavItem
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * A navigation item for the sidebar or menu.
 *
 * @param props - The component props.
 * @param props.href - The URL to link to.
 * @param props.icon - The icon component to display.
 * @param props.title - The title of the navigation item.
 * @param props.isActive - Whether the item is currently active.
 * @returns {JSX.Element} The rendered navigation item.
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
