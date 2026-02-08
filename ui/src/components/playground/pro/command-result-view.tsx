/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Badge } from "@/components/ui/badge";
import { CheckCircle, XCircle, Clock } from "lucide-react";

interface CommandResult {
    stdout?: string;
    stderr?: string;
    return_code?: number;
    command?: string;
    start_time?: string;
    end_time?: string;
    duration?: string;
    status?: string;
}

export function CommandResultView({ result }: { result: unknown }) {
    const data = result as CommandResult;
    // Default to success if return_code is missing but status is success, or if return_code is 0
    const isSuccess = data.return_code === 0 || (data.return_code === undefined && data.status === 'success');

    let durationDisplay = data.duration;
    if (!durationDisplay && data.start_time && data.end_time) {
        try {
            const start = new Date(data.start_time).getTime();
            const end = new Date(data.end_time).getTime();
            const diff = end - start;
            durationDisplay = diff < 1000 ? `${diff}ms` : `${(diff / 1000).toFixed(2)}s`;
        } catch (e) {
            // ignore date parsing errors
        }
    }

    return (
        <div className="flex flex-col gap-0 border rounded-md overflow-hidden bg-card w-full">
            {/* Header / Metadata */}
            <div className="flex items-center justify-between p-2 bg-muted/50 border-b text-xs">
                <div className="flex items-center gap-2">
                    <Badge variant={isSuccess ? "outline" : "destructive"} className="gap-1 h-5 font-normal">
                        {isSuccess ? <CheckCircle className="h-3 w-3 text-green-500" /> : <XCircle className="h-3 w-3" />}
                        {data.return_code !== undefined ? `Exit: ${data.return_code}` : (data.status || "Unknown")}
                    </Badge>
                    {durationDisplay && (
                         <span className="flex items-center gap-1 text-muted-foreground ml-2">
                            <Clock className="h-3 w-3" /> {durationDisplay}
                        </span>
                    )}
                </div>
                 {data.command && (
                    <div className="font-mono text-muted-foreground truncate max-w-[50%] text-right" title={data.command}>
                        $ {data.command}
                    </div>
                )}
            </div>

            {/* Terminal Output */}
            <div className="bg-[#1e1e1e] text-gray-300 p-3 font-mono text-xs overflow-x-auto min-h-[100px] max-h-[500px] overflow-y-auto">
                {data.stdout ? (
                    <div className="whitespace-pre-wrap">
                        {data.stdout}
                    </div>
                ) : null}

                {data.stderr ? (
                    <div className="whitespace-pre-wrap text-red-400 mt-2 pt-2 border-t border-red-900/30">
                        <span className="opacity-50 select-none text-[10px] uppercase mb-1 block">Stderr output:</span>
                        {data.stderr}
                    </div>
                ) : null}

                {!data.stdout && !data.stderr && (
                    <div className="text-muted-foreground italic opacity-50">No output returned.</div>
                )}
            </div>
        </div>
    );
}
