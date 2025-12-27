import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { LayoutDashboard, Server, Wrench, FileText, MessageSquare, Settings, Activity, Webhook, Layers } from "lucide-react";
import { Toaster } from "@/components/ui/toaster";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });

export const metadata: Metadata = {
  title: "MCP Any",
  description: "Enterprise MCP Server Management",
};

const SidebarItem = ({ href, icon: Icon, label, active }: { href: string; icon: any; label: string; active?: boolean }) => (
  <Link
    href={href}
    className={cn(
      "flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md transition-colors",
      active
        ? "bg-primary/10 text-primary"
        : "text-muted-foreground hover:bg-muted hover:text-foreground"
    )}
  >
    <Icon className="w-4 h-4" />
    {label}
  </Link>
);

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={cn(inter.variable, "min-h-screen font-sans antialiased bg-slate-50")}>
        <div className="flex h-screen overflow-hidden">
          {/* Sidebar */}
          <aside className="w-64 border-r bg-background/80 backdrop-blur-xl hidden md:flex flex-col">
            <div className="p-6 h-14 flex items-center border-b border-transparent">
               <div className="flex items-center gap-2 font-bold text-lg text-primary">
                 <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-primary-foreground">
                    M
                 </div>
                 MCP Any
               </div>
            </div>

            <div className="flex-1 overflow-y-auto py-6 px-4 space-y-6">
              <div>
                <h4 className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground/70">Platform</h4>
                <div className="space-y-1">
                  <SidebarItem href="/" icon={LayoutDashboard} label="Dashboard" />
                  <SidebarItem href="/services" icon={Server} label="Services" />
                </div>
              </div>

              <div>
                <h4 className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground/70">Resources</h4>
                <div className="space-y-1">
                  <SidebarItem href="/tools" icon={Wrench} label="Tools" />
                  <SidebarItem href="/resources" icon={FileText} label="Resources" />
                  <SidebarItem href="/prompts" icon={MessageSquare} label="Prompts" />
                </div>
              </div>

               <div>
                <h4 className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground/70">Advanced</h4>
                <div className="space-y-1">
                  <SidebarItem href="/middleware" icon={Layers} label="Middleware" />
                  <SidebarItem href="/webhooks" icon={Webhook} label="Webhooks" />
                  <SidebarItem href="/logs" icon={Activity} label="Logs" />
                </div>
              </div>
            </div>

            <div className="p-4 border-t mt-auto">
               <SidebarItem href="/settings" icon={Settings} label="Settings" />
            </div>
          </aside>

          {/* Main Content */}
          <main className="flex-1 flex flex-col h-full overflow-hidden relative">
            {/* Header (Mobile only basically, or breadcrumbs) */}
             <header className="h-14 border-b bg-background/50 backdrop-blur-sm flex items-center px-6 md:hidden">
               <span className="font-semibold">MCP Any</span>
             </header>

            <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8">
               {children}
            </div>
          </main>
        </div>
        <Toaster />
      </body>
    </html>
  );
}
