/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { CheckCircle2 } from 'lucide-react';

/**
 * Intent: Document StepReview
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
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
                 <h3 className="font-medium">Spec Preview</h3>
                 <div className="rounded-md overflow-hidden border">
                     <ScrollArea className="max-h-[300px]">
                         <pre className="p-4 text-xs font-mono bg-[#1e1e1e] text-gray-200 whitespace-pre-wrap break-all">
                             {JSON.stringify(config, null, 2)}
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
