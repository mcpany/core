/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Button } from "@/components/ui/button";
import { ArrowRight, Sparkles } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export function WelcomeStep({ onStart }: { onStart: () => void }) {
  return (
    <Card className="w-full max-w-lg text-center border-none shadow-xl bg-background/60 backdrop-blur-xl">
      <CardHeader className="space-y-4 pb-2">
        <div className="mx-auto bg-primary/10 p-4 rounded-full w-fit mb-2">
            <Sparkles className="w-10 h-10 text-primary" />
        </div>
        <CardTitle className="text-3xl font-bold tracking-tight">Welcome to MCP Any</CardTitle>
        <CardDescription className="text-lg">
          Your universal gateway for AI tools. Let's get you set up with your first service.
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-6">
        <Button size="lg" className="w-full text-lg h-12 gap-2" onClick={onStart}>
          Get Started <ArrowRight className="w-5 h-5" />
        </Button>
      </CardContent>
    </Card>
  );
}
