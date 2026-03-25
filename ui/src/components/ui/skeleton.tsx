/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { cn } from "@/lib/utils"

/**
 * Summary: Skeleton component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.className: The name of the class.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-muted", className)}
      {...props}
    />
  )
}

export { Skeleton }
