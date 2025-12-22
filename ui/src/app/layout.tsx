/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type {Metadata} from 'next';
import './globals.css';
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { ThemeProvider } from "@/components/theme-provider"
import { ThemeToggle } from "@/components/theme-toggle"
import { Sidebar } from "@/components/sidebar"

export const metadata: Metadata = {
  title: 'MCPAny Manager',
  description: 'A server management UI for the MCP Any server.',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Inter&display=swap" rel="stylesheet" />
        <link href="https://fonts.googleapis.com/css2?family=Roboto+Mono&display=swap" rel="stylesheet" />
      </head>
      <body className="font-body antialiased bg-gray-50 dark:bg-gray-900">
        <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
          <TooltipProvider>
            <div className="flex min-h-screen">
                <Sidebar className="flex-shrink-0" />
                <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                    <header className="flex items-center justify-between h-14 px-4 border-b bg-background/95 backdrop-blur z-40">
                         {/* Breadcrumbs or Title could go here */}
                         <div className="flex items-center font-medium">
                            {/* Placeholder for Breadcrumbs */}
                            Dashboard
                         </div>
                        <div className="flex items-center gap-2">
                             <ThemeToggle />
                        </div>
                    </header>
                    <main className="flex-1 overflow-auto p-4 md:p-6 lg:p-8">
                        {children}
                    </main>
                </div>
            </div>
            <Toaster />
          </TooltipProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
