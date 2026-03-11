/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { CheckCircle2 } from 'lucide-react';
import SyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';
import vs2015 from 'react-syntax-highlighter/dist/esm/styles/hljs/vs2015';

SyntaxHighlighter.registerLanguage('json', json);

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
                 <h3 className="font-medium">Spec Preview</h3>
                 <div className="rounded-md overflow-hidden border">
                     <SyntaxHighlighter language="json" style={vs2015} showLineNumbers customStyle={{ margin: 0, maxHeight: '300px' }}>
                        {JSON.stringify(config, null, 2)}
                     </SyntaxHighlighter>
                 </div>
            </div>

            <Button className="w-full" size="lg" onClick={() => onComplete(config)}>
                Finish & Save to Local Marketplace
            </Button>
        </div>
    );
}
