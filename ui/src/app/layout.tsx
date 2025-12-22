/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type {Metadata} from 'next';
import './globals.css';
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { GlobalSearch } from "@/components/global-search";

export const metadata: Metadata = {
  title: 'MCPAny Manager',
  description: 'A server management UI for the MCP Any server.',
};

/**
 * Root layout component for the application.
 *
 * @param props - The component props.
 * @param props.children - The child components to render.
 * @returns The root HTML structure.
 */
export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Inter&display=swap" rel="stylesheet" />
      </head>
      <body className="font-body antialiased">
        <TooltipProvider>
          {children}
          <GlobalSearch />
          <Toaster />
        </TooltipProvider>
      </body>
    </html>
  );
}
