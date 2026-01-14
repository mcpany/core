/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "./ui/badge";

/**
 * Props for the ServicePropertyCard component.
 */
interface ServicePropertyCardProps {
    /** The title of the card. */
    title: string;
    /** The data to display as key-value pairs. */
    data: Record<string, string | number | boolean>;
}

/**
 * A card component that displays a list of properties.
 */
export function ServicePropertyCard({ title, data }: ServicePropertyCardProps) {
    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-xl">{title}</CardTitle>
            </CardHeader>
            <CardContent>
                <dl className="space-y-2">
                    {Object.entries(data).map(([key, value]) => (
                        <div key={key} className="flex justify-between items-start">
                            <dt className="text-muted-foreground">{key}</dt>
                            <dd className="text-right font-mono text-sm">
                                {typeof value === 'boolean' ? (
                                    <Badge variant={value ? "default" : "secondary"}>{value ? "Enabled" : "Disabled"}</Badge>
                                ) : (
                                    String(value)
                                )}
                            </dd>
                        </div>
                    ))}
                </dl>
            </CardContent>
        </Card>
    )
}
