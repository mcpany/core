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
import { GlobalSearch } from "@/components/global-search"
import { Separator } from "@/components/ui/separator"
import { UserProvider } from "@/components/user-context"
import { KeyboardShortcutsProvider } from "@/contexts/keyboard-shortcuts-context"

export const metadata: Metadata = {
  title: 'MCPAny Manager',
  description: 'A server management UI for the MCP Any server.',
};

/**
 * Root layout component for the application.
 * Wraps the application with necessary providers and the sidebar layout.
 * @param props.children - The child components to render.
 * @returns The root layout structure.
 */
export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Inter&display=swap" rel="stylesheet" />
        <link href="https://fonts.googleapis.com/css2?family=Roboto+Mono&display=swap" rel="stylesheet" />
      </head>
      <body className="font-body antialiased">
        <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
          <TooltipProvider>
            <UserProvider>
              <KeyboardShortcutsProvider>
                <SidebarProvider>
                  <AppSidebar />
                  <SidebarInset>
                  <header className="flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 backdrop-blur px-4 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
                    <SidebarTrigger className="-ml-1" />
                    <Separator orientation="vertical" className="mr-2 h-4" />
                     <div className="flex-1 flex items-center justify-between">
                         <div className="font-medium text-sm">
                             MCP Any
                         </div>
                         <div className="flex items-center gap-2">
                             <GlobalSearch />
                             <ThemeToggle />
                         </div>
                     </div>
                  </header>
                  <main className="flex-1 overflow-auto p-4 md:p-6 lg:p-8">
                    {children}
                  </main>
                  </SidebarInset>
                </SidebarProvider>
              </KeyboardShortcutsProvider>
            </UserProvider>
            <Toaster />
          </TooltipProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
