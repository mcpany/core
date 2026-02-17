/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { cn } from "@/lib/utils"

/**
 * Used to show a placeholder while content is loading.
 *
 * @param props - The component props.
 * @param props.className - Additional class names to apply.
 * @returns The rendered skeleton component.
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
