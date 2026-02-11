/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, CardHeader, CardContent } from "@/components/ui/card";

/**
 * WidgetSkeleton component.
 * Renders a placeholder card with a loading animation.
 * Used as a fallback while heavy widgets are loading.
 */
export function WidgetSkeleton() {
    return (
        <Card className="h-full border border-muted bg-card/50 backdrop-blur-sm shadow-sm animate-pulse">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <div className="h-5 w-1/3 bg-muted rounded" />
                <div className="h-4 w-4 bg-muted rounded-full" />
            </CardHeader>
            <CardContent>
                <div className="space-y-2 mt-2">
                    <div className="h-4 w-full bg-muted/50 rounded" />
                    <div className="h-4 w-5/6 bg-muted/50 rounded" />
                    <div className="h-4 w-4/6 bg-muted/50 rounded" />
                </div>
            </CardContent>
        </Card>
    );
}
