/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Suspense } from "react";
import { Outlet } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { ThemeProvider } from "@/components/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import { GlobalSearch } from "@/components/global-search";
import { ConnectClientButton } from "@/components/connect-client-button";
import { Separator } from "@/components/ui/separator";
import { UserProvider } from "@/components/user-context";
import { KeyboardShortcutsProvider } from "@/contexts/keyboard-shortcuts-context";
import { ServiceHealthProvider } from "@/contexts/service-health-context";
import { SystemStatusBanner } from "@/components/system-status-banner";
import { ErrorBoundary } from "@/components/ui/error-boundary";

/**
 * Summary: PageFallback component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const PageFallback = () => (
  <div className="flex items-center justify-center h-full min-h-[200px]">
    <Loader2 className="animate-spin h-8 w-8 text-muted-foreground" />
  </div>
);

/**
 * Layout component that wraps all main application routes with the
 * sidebar, header, and context providers.  Uses React Router's
 * <Outlet /> to render the matched child route.
 */
export function Layout() {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <TooltipProvider>
        <UserProvider>
          <ServiceHealthProvider>
            <KeyboardShortcutsProvider>
              <SidebarProvider>
                <AppSidebar />
                <SidebarInset>
                  <header className="flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 backdrop-blur px-4 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
                    <SidebarTrigger className="-ml-1" />
                    <Separator orientation="vertical" className="mr-2 h-4" />
                    <div className="flex-1 flex items-center justify-between">
                      <div className="font-medium text-sm">MCP Any</div>
                      <div className="flex items-center gap-2">
                        <ConnectClientButton />
                        <GlobalSearch />
                        <ThemeToggle />
                      </div>
                    </div>
                  </header>
                  <SystemStatusBanner />
                  <main className="flex-1 overflow-auto p-4 md:p-6 lg:p-8">
                    <ErrorBoundary>
                      <Suspense fallback={<PageFallback />}>
                        <Outlet />
                      </Suspense>
                    </ErrorBoundary>
                  </main>
                </SidebarInset>
              </SidebarProvider>
            </KeyboardShortcutsProvider>
          </ServiceHealthProvider>
        </UserProvider>
        <Toaster />
      </TooltipProvider>
    </ThemeProvider>
  );
}
