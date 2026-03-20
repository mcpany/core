"use client";

import { MainLayout } from "@/components/layout/main-layout";
import { SystemHealth } from "@/components/diagnostics/system-health";
import { DiscoveryStatus } from "@/components/diagnostics/discovery-status";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ShieldCheck, Activity, Search } from "lucide-react";

export default function DiagnosticsPage() {
  return (
    <MainLayout>
      <div className="flex flex-col h-full bg-muted/10">
        <header className="flex h-14 items-center gap-4 border-b bg-background/80 backdrop-blur-xl px-6 lg:h-[60px]">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-primary" />
            <h1 className="text-lg font-semibold tracking-tight">Security & Diagnostics</h1>
          </div>
        </header>

        <main className="flex-1 overflow-auto p-6 md:p-8">
          <div className="max-w-6xl mx-auto space-y-8">
             <div className="flex flex-col gap-1.5">
               <h2 className="text-3xl font-bold tracking-tight">System Analysis</h2>
               <p className="text-muted-foreground text-sm">
                 Monitor system integrity, network exposure, and registry configuration for your MCP Any node.
               </p>
             </div>

             <Tabs defaultValue="health" className="w-full">
                <TabsList className="grid w-full md:w-[400px] grid-cols-2 mb-8 bg-muted/50 p-1">
                   <TabsTrigger value="health" className="flex items-center gap-2 rounded-sm text-xs font-medium uppercase tracking-wider data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all">
                      <Activity className="h-3.5 w-3.5" />
                      Connectivity & Security
                   </TabsTrigger>
                   <TabsTrigger value="discovery" className="flex items-center gap-2 rounded-sm text-xs font-medium uppercase tracking-wider data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all">
                      <Search className="h-3.5 w-3.5" />
                      Registry Sources
                   </TabsTrigger>
                </TabsList>

                <TabsContent value="health" className="mt-0 outline-none animate-in fade-in-50 zoom-in-[0.98] duration-300">
                   <SystemHealth />
                </TabsContent>

                <TabsContent value="discovery" className="mt-0 outline-none animate-in fade-in-50 zoom-in-[0.98] duration-300">
                   <DiscoveryStatus />
                </TabsContent>
             </Tabs>
          </div>
        </main>
      </div>
    </MainLayout>
  );
}
