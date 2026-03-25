/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
import * as React from "react"
import * as TooltipPrimitive from "@radix-ui/react-tooltip"

import { cn } from "@/lib/utils"

/**
 * Summary: TooltipProvider component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const TooltipProvider = TooltipPrimitive.Provider

/**
 * Summary: Tooltip component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const Tooltip = TooltipPrimitive.Root

/**
 * Summary: TooltipTrigger component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const TooltipTrigger = TooltipPrimitive.Trigger

/**
 * Summary: TooltipContent component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
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
