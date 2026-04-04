/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

/**
 * ThemeProvider executes the ThemeProvider logic.
 *
 * Summary: Executes the ThemeProvider logic.
 *
 * @param { children - The { children parameter.
 * @param ...props } - The ...props } parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function ThemeProvider({ children, ...props }: React.ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}
