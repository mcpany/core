/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { SERVICE_TEMPLATES } from "@/lib/templates";
import { ArrowRight, BookOpen, ExternalLink, Plus } from "lucide-react";
import Link from "next/link";

interface ServiceOnboardingProps {
    onSelectTemplate: (templateId: string) => void;
}

/**
 * ServiceOnboarding component.
 * displayed when no services are registered.
 *
 * @param props - The component props.
 * @param props.onSelectTemplate - Callback when a template is selected.
 * @returns The rendered component.
 */
export function ServiceOnboarding({ onSelectTemplate }: ServiceOnboardingProps) {
    // Select popular templates for the quick start grid
    const featuredIds = ["github", "postgres", "filesystem", "web-search", "slack", "google-maps"];
    const featuredTemplates = SERVICE_TEMPLATES.filter(t => featuredIds.includes(t.id));

    return (
        <div className="flex flex-col items-center justify-center py-12 px-4 animate-in fade-in zoom-in-95 duration-500">
            <div className="text-center max-w-2xl mx-auto mb-10 space-y-4">
                <div className="bg-primary/10 w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-6 rotate-3 hover:rotate-6 transition-transform">
                    <Plus className="w-8 h-8 text-primary" />
                </div>
                <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
                    Connect your first service
                </h1>
                <p className="text-lg text-muted-foreground">
                    MCP Any acts as a universal adapter. Connect your existing tools, databases, and APIs to make them available to your AI agents instantly.
                </p>
                <div className="flex items-center justify-center gap-4 pt-2">
                     <Button onClick={() => onSelectTemplate("empty")} size="lg" className="gap-2">
                        Connect Manually <ArrowRight className="w-4 h-4" />
                     </Button>
                     <Button variant="outline" size="lg" asChild>
                        <Link href="/docs" className="gap-2">
                            <BookOpen className="w-4 h-4" /> Read Documentation
                        </Link>
                     </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-w-5xl w-full">
                {featuredTemplates.map(template => {
                    const Icon = template.icon;
                    return (
                        <Card
                            key={template.id}
                            className="group hover:border-primary/50 transition-all hover:shadow-md cursor-pointer relative overflow-hidden"
                            onClick={() => onSelectTemplate(template.id)}
                        >
                            <div className="absolute top-0 right-0 p-4 opacity-0 group-hover:opacity-100 transition-opacity">
                                <ArrowRight className="w-4 h-4 text-primary" />
                            </div>
                            <CardHeader className="pb-2">
                                <div className="w-10 h-10 rounded-lg bg-primary/5 text-primary flex items-center justify-center mb-2 group-hover:bg-primary/10 transition-colors">
                                    <Icon className="w-5 h-5" />
                                </div>
                                <CardTitle className="text-base">{template.name}</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <CardDescription className="line-clamp-2">
                                    {template.description}
                                </CardDescription>
                            </CardContent>
                            <CardFooter className="pt-0 text-xs text-muted-foreground">
                                <span className="bg-muted px-2 py-1 rounded-full group-hover:bg-primary/10 group-hover:text-primary transition-colors">
                                    {template.category}
                                </span>
                            </CardFooter>
                        </Card>
                    );
                })}
            </div>

            <div className="mt-12 text-center text-sm text-muted-foreground">
                <p>
                    Looking for something else? We support <Link href="/docs/cli" className="underline hover:text-primary underline-offset-4">Command Line Interface</Link> and <Link href="/docs/custom" className="underline hover:text-primary underline-offset-4">Custom Integrations</Link>.
                </p>
            </div>
        </div>
    );
}
