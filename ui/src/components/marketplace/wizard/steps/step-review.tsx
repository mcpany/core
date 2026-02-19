/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Button } from '@/components/ui/button';
import { CheckCircle2 } from 'lucide-react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface StepReviewProps {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onComplete: (config: any) => void;
}

/**
 * StepReview component.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function StepReview({ onComplete }: StepReviewProps) {
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
                     <SyntaxHighlighter language="json" style={oneDark} showLineNumbers customStyle={{ margin: 0, maxHeight: '300px' }}>
                        {JSON.stringify(config, null, 2)}
                     </SyntaxHighlighter>
                 </div>
            </div>

            <Button className="w-full" size="lg" onClick={() => onComplete(config)}>
                Deploy Service
            </Button>
        </div>
    );
}
