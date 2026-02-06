/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

/**
 * Provides theme context to the application using next-themes.
 * It wraps the application and handles light/dark mode switching.
 *
 * @param props - The component props (ThemeActionProviderProps).
 * @param props.children - The child components to wrap.
 * @param props.attribute - The HTML attribute to modify (e.g., "class").
 * @param props.defaultTheme - The default theme (e.g., "system").
 * @param props.enableSystem - Whether to enable system preference detection.
 * @returns The ThemeProvider component.
 */
export function ThemeProvider({ children, ...props }: React.ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}
