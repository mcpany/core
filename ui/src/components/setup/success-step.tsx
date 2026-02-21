/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Button } from "@/components/ui/button";
import { CheckCircle2, LayoutDashboard, Plus } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import Link from "next/link";

export function SuccessStep() {
  return (
    <Card className="w-full max-w-lg text-center border-none shadow-xl bg-background/60 backdrop-blur-xl">
      <CardHeader className="space-y-4 pb-2">
        <div className="mx-auto bg-green-500/10 p-4 rounded-full w-fit mb-2">
            <CheckCircle2 className="w-12 h-12 text-green-500" />
        </div>
        <CardTitle className="text-3xl font-bold tracking-tight">You are All Set!</CardTitle>
        <CardDescription className="text-lg">
          Your service has been successfully connected and is ready to use.
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-6 space-y-3">
        <Link href="/">
            <Button size="lg" className="w-full text-lg h-12 gap-2">
            <LayoutDashboard className="w-5 h-5" /> Go to Dashboard
            </Button>
        </Link>
        <Link href="/setup">
             <Button variant="outline" className="w-full gap-2" onClick={() => window.location.reload()}>
                <Plus className="w-4 h-4" /> Connect Another Service
            </Button>
        </Link>
      </CardContent>
    </Card>
  );
}
