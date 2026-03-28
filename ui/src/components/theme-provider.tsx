/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

/**
 * Intent: Document ThemeProvider
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
 * Provides theme context to the application using next-themes.
 *
 * @param props - The properties for the ThemeProvider, including children and theme settings.
 * @returns The ThemeProvider component.
 */
export function ThemeProvider({ children, ...props }: React.ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}
