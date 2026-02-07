/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

/**
 * Provides theme context to the application using next-themes.
 *
 * Summary: Wraps the application or component tree with the theme provider to enable light/dark mode switching.
 *
 * @param props - The properties for the ThemeProvider (children, defaultTheme, etc.).
 * @returns The wrapped component tree with theme context.
 */
export function ThemeProvider({ children, ...props }: React.ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}
