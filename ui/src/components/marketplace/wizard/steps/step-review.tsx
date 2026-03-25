/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { CheckCircle2, Wrench, Book, Database, Shield, Server, ArrowRight } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';

/**
 * StepReview.
 *
 * @param { onComplete - The { onComplete.
 */
export function StepReview({ onComplete }: { onComplete: (config: any) => void }) {
    const { state } = useWizard();
    const { config } = state;

    // Helper to extract nested arrays from the config regardless of service type
    const extractArray = (key: string) => {
        if (!config) return [];
        for (const [k, v] of Object.entries(config)) {
            if (v && typeof v === 'object' && key in v) {
                return (v as any)[key] || [];
            }
        }
        return (config as any)[key] || [];
    };

    const tools = extractArray('tools');
    const prompts = extractArray('prompts');
    const resources = extractArray('resources');

    // Determine Auth type
    const hasAuth = !!(config as any)?.upstreamAuth;
    const authType = hasAuth ? Object.keys((config as any).upstreamAuth || {})[0] : 'None';

    // Determine Service Type
    let serviceType = 'Unknown';
    if (config?.httpService) serviceType = 'HTTP';
    else if (config?.grpcService) serviceType = 'gRPC';
    else if (config?.commandLineService) serviceType = 'CLI';
    else if (config?.mcpService) serviceType = 'MCP Upstream';

    return (
        <div className="flex flex-col h-full max-h-[60vh]">
            <ScrollArea className="flex-1 pr-4">
                <div className="space-y-6 pb-6">
                    <div className="bg-green-500/10 border border-green-500/20 text-green-600 dark:text-green-400 p-4 rounded-lg flex items-center gap-3 backdrop-blur-sm shadow-sm">
                        <CheckCircle2 className="h-6 w-6 shrink-0" />
                        <div>
                            <div className="font-semibold">Configuration Ready</div>
                            <div className="text-sm opacity-90">Review your service setup before installing.</div>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <Card className="shadow-sm border-muted/60 bg-gradient-to-br from-background to-muted/10">
                            <CardHeader className="pb-2">
                                <CardTitle className="text-md flex items-center gap-2">
                                    <Server className="h-4 w-4 text-primary" />
                                    General Information
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="text-sm space-y-2">
                                <div className="flex justify-between border-b pb-1">
                                    <span className="text-muted-foreground">Name</span>
                                    <span className="font-medium truncate max-w-[200px]">{config?.name || 'Unnamed'}</span>
                                </div>
                                <div className="flex justify-between border-b pb-1">
                                    <span className="text-muted-foreground">Type</span>
                                    <Badge variant="outline" className="bg-background">{serviceType}</Badge>
                                </div>
                                <div className="flex justify-between border-b pb-1">
                                    <span className="text-muted-foreground">Version</span>
                                    <span>{config?.version || '1.0.0'}</span>
                                </div>
                                <div className="flex justify-between pt-1">
                                    <span className="text-muted-foreground flex items-center gap-1">
                                        <Shield className="h-3 w-3" /> Auth
                                    </span>
                                    <Badge variant={hasAuth ? 'default' : 'secondary'} className="capitalize">{authType}</Badge>
                                </div>
                            </CardContent>
                        </Card>

                        <Card className="shadow-sm border-muted/60 bg-gradient-to-bl from-background to-muted/10">
                            <CardHeader className="pb-2">
                                <CardTitle className="text-md flex items-center gap-2">
                                    <CheckCircle2 className="h-4 w-4 text-primary" />
                                    Capabilities Discovered
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <div className="p-1.5 bg-blue-500/10 rounded-md text-blue-500"><Wrench className="h-4 w-4" /></div>
                                        Tools
                                    </div>
                                    <span className="font-mono font-semibold bg-muted px-2 py-0.5 rounded-md">{tools.length}</span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <div className="p-1.5 bg-purple-500/10 rounded-md text-purple-500"><Book className="h-4 w-4" /></div>
                                        Prompts
                                    </div>
                                    <span className="font-mono font-semibold bg-muted px-2 py-0.5 rounded-md">{prompts.length}</span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <div className="p-1.5 bg-amber-500/10 rounded-md text-amber-500"><Database className="h-4 w-4" /></div>
                                        Resources
                                    </div>
                                    <span className="font-mono font-semibold bg-muted px-2 py-0.5 rounded-md">{resources.length}</span>
                                </div>
                            </CardContent>
                        </Card>
                    </div>

                    <Accordion type="single" collapsible className="w-full border rounded-lg bg-card/50 backdrop-blur-sm">
                        <AccordionItem value="json-dump" className="border-b-0">
                            <AccordionTrigger className="px-4 py-3 hover:bg-muted/30 transition-colors">
                                <span className="text-sm font-medium">View Raw JSON Specification</span>
                            </AccordionTrigger>
                            <AccordionContent className="pt-0 pb-4 px-4">
                                <div className="rounded-md overflow-hidden border bg-[#1e1e1e] shadow-inner mt-2">
                                    <ScrollArea className="h-[250px]">
                                        <pre className="p-4 text-xs font-mono text-gray-300 whitespace-pre-wrap break-all selection:bg-primary/30">
                                            {JSON.stringify(config, null, 2)}
                                        </pre>
                                    </ScrollArea>
                                </div>
                            </AccordionContent>
                        </AccordionItem>
                    </Accordion>
                </div>
            </ScrollArea>

            <div className="pt-4 border-t mt-auto shrink-0">
                <Button
                    className="w-full group shadow-md hover:shadow-lg transition-all"
                    size="lg"
                    onClick={() => onComplete(config)}
                >
                    Finish & Save to Local Marketplace
                    <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
                </Button>
            </div>
        </div>
    );
}
