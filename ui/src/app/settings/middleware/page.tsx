
"use client";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, ArrowRight, Shield, Zap, FileJson } from "lucide-react";
import { Badge } from "@/components/ui/badge";

export default function MiddlewarePage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Middleware</h2>
        <Button><Plus className="mr-2 h-4 w-4" /> Add Pipeline Step</Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50 border-muted/20">
        <CardHeader>
          <CardTitle>Request Processing Pipeline</CardTitle>
          <CardDescription>Visual management of middleware applied to requests.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="relative flex flex-col items-center justify-center space-y-8 py-10">
            {/* Pipeline Step 1 */}
            <div className="flex items-center w-full max-w-2xl bg-card border rounded-lg p-4 shadow-sm z-10 relative">
                <div className="h-10 w-10 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center mr-4">
                    <Shield className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                </div>
                <div className="flex-1">
                    <h3 className="font-semibold">Authentication</h3>
                    <p className="text-sm text-muted-foreground">Validates API keys and JWT tokens.</p>
                </div>
                <Badge variant="outline" className="ml-2">System</Badge>
            </div>

            <ArrowRight className="h-6 w-6 text-muted-foreground rotate-90" />

            {/* Pipeline Step 2 */}
            <div className="flex items-center w-full max-w-2xl bg-card border rounded-lg p-4 shadow-sm z-10 relative">
                 <div className="h-10 w-10 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center mr-4">
                    <Zap className="h-5 w-5 text-orange-600 dark:text-orange-400" />
                </div>
                <div className="flex-1">
                    <h3 className="font-semibold">Rate Limiting</h3>
                    <p className="text-sm text-muted-foreground">Global limit: 100 req/sec.</p>
                </div>
                 <Badge className="ml-2">Active</Badge>
            </div>

            <ArrowRight className="h-6 w-6 text-muted-foreground rotate-90" />

            {/* Pipeline Step 3 */}
            <div className="flex items-center w-full max-w-2xl bg-card border rounded-lg p-4 shadow-sm z-10 relative">
                 <div className="h-10 w-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mr-4">
                    <FileJson className="h-5 w-5 text-green-600 dark:text-green-400" />
                </div>
                <div className="flex-1">
                    <h3 className="font-semibold">Input Validation</h3>
                    <p className="text-sm text-muted-foreground">Checks request body against schemas.</p>
                </div>
                 <Badge className="ml-2">Active</Badge>
            </div>

            {/* Pipeline Visualization Line */}
             <div className="absolute top-10 bottom-10 left-1/2 w-0.5 bg-border -z-0 hidden md:block" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
