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
 * @param props - The component props.
 * @param props.children - The child components.
 * @param props.props - Additional props passed to the NextThemesProvider.
 * @returns {JSX.Element} The rendered theme provider.
 */
export function ThemeProvider({ children, ...props }: React.ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}
