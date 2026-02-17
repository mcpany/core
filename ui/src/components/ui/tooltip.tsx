/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import * as TooltipPrimitive from "@radix-ui/react-tooltip"

import { cn } from "@/lib/utils"

/**
 * Wraps the application to provide tooltip functionality.
 *
 * @param props - The component props.
 * @returns The rendered tooltip provider.
 */
const TooltipProvider = TooltipPrimitive.Provider

/**
 * A popup that displays information related to an element when the element receives keyboard focus or the mouse hovers over it.
 *
 * @param props - The component props.
 * @returns The rendered tooltip root.
 */
const Tooltip = TooltipPrimitive.Root

/**
 * The element that triggers the tooltip.
 *
 * @param props - The component props.
 * @returns The rendered tooltip trigger.
 */
const TooltipTrigger = TooltipPrimitive.Trigger

/**
 * The content to display within the tooltip.
 *
 * @param props - The component props.
 * @param props.className - Additional class names to apply.
 * @param props.sideOffset - The offset from the trigger.
 * @returns The rendered tooltip content.
 */
const TooltipContent = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
  <TooltipPrimitive.Content
    ref={ref}
    sideOffset={sideOffset}
    className={cn(
      "z-50 overflow-hidden rounded-md border bg-popover px-3 py-1.5 text-sm text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
      className
    )}
    {...props}
  />
))
TooltipContent.displayName = TooltipPrimitive.Content.displayName

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider }
