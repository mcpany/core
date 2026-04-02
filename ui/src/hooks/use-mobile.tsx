/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import * as React from "react"

const MOBILE_BREAKPOINT = 768

/**
 * Summary: Hook to dynamically detect if the current viewport is mobile-sized.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - boolean: True if the viewport is < 768px width, false otherwise.
 *
 * Errors:
 *   - N/A: Only runs on the client.
 *
 * Side Effects:
 *   - Attaches a matchMedia event listener to track window resize events.
 */
export function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState<boolean | undefined>(undefined)

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`)
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    }
    mql.addEventListener("change", onChange)
    setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    return () => mql.removeEventListener("change", onChange)
  }, [])

  return !!isMobile
}
