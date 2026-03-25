/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
import * as React from "react"
import * as PopoverPrimitive from "@radix-ui/react-popover"

import { cn } from "@/lib/utils"

/**
 * Summary: Popover component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.align: The align property.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const Popover = PopoverPrimitive.Root

/**
 * Summary: PopoverTrigger component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.align: The align property.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PopoverTrigger = PopoverPrimitive.Trigger

/**
 * Summary: PopoverContent component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *   - props.align: The align property.
 *   - props.sideOffset: The unique identifier for seOffset.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PopoverContent = React.forwardRef<
  React.ElementRef<typeof PopoverPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>
>(({ className, align = "center", sideOffset = 4, ...props }, ref) => (
  <PopoverPrimitive.Portal>
    <PopoverPrimitive.Content
      ref={ref}
      align={align}
      sideOffset={sideOffset}
      className={cn(
        "z-50 w-72 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
        className
      )}
      {...props}
    />
  </PopoverPrimitive.Portal>
))
PopoverContent.displayName = PopoverPrimitive.Content.displayName

export { Popover, PopoverTrigger, PopoverContent }
