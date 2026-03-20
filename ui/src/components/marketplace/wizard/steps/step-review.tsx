/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { CheckCircle2, Code } from 'lucide-react';
import * as yaml from 'js-yaml';

/**
 * StepReview.
 *
 * @param { onComplete - The { onComplete.
 */
export function StepReview({ onComplete }: { onComplete: (config: any) => void }) {
    const { state } = useWizard();
    const { config } = state;

    return (
        <div className="space-y-6">
            <div className="bg-green-500/10 text-green-500 p-4 rounded-lg flex items-center gap-3">
                <CheckCircle2 className="h-6 w-6" />
                <div className="font-medium">Configuration Ready</div>
            </div>

            <div className="space-y-2">
                 <div className="flex items-center gap-2 mb-2">
                     <Code className="h-4 w-4 text-muted-foreground" />
                     <h3 className="font-medium">Spec Preview</h3>
                 </div>
                 <div className="rounded-md overflow-hidden border bg-[#1e1e1e]">
                     <ScrollArea className="max-h-[300px]">
                         <pre className="p-4 text-xs font-mono text-[#ce9178] whitespace-pre-wrap break-all">
                             {yaml.dump(config, { indent: 2, skipInvalid: true })}
                         </pre>
                     </ScrollArea>
                 </div>
            </div>

            <Button className="w-full" size="lg" onClick={() => onComplete(config)}>
                Finish & Save to Local Marketplace
            </Button>
        </div>
    );
}
