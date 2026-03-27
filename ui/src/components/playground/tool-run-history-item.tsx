import { useState } from "react";
import { ChevronDown, ChevronRight, Clock, CheckCircle2, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";

interface ToolRunHistoryItemProps {
    log: {
        timestamp: string;
        arguments: string;
        result: string;
        error: string;
        durationMs: number;
    };
}

export function ToolRunHistoryItem({ log }: ToolRunHistoryItemProps) {
    const [expanded, setExpanded] = useState(false);

    const isError = !!log.error;
    const time = new Date(log.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    const date = new Date(log.timestamp).toLocaleDateString();

    let parsedArgs: any = log.arguments;
    let parsedResult: any = log.result;

    try {
        if (typeof log.arguments === 'string') parsedArgs = JSON.parse(log.arguments);
    } catch (e) {}

    try {
        if (typeof log.result === 'string') parsedResult = JSON.parse(log.result);
    } catch (e) {}

    return (
        <div className="flex flex-col border-b last:border-0 hover:bg-muted/30 transition-colors">
            <button
                onClick={() => setExpanded(!expanded)}
                className="flex items-center justify-between p-3 w-full text-left"
            >
                <div className="flex items-center gap-3">
                    <div className="flex-shrink-0">
                        {isError ? (
                            <XCircle className="h-4 w-4 text-destructive" />
                        ) : (
                            <CheckCircle2 className="h-4 w-4 text-green-500" />
                        )}
                    </div>
                    <div className="flex flex-col">
                        <span className="text-sm font-medium">{time} <span className="text-muted-foreground text-xs font-normal">({date})</span></span>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-1 text-xs text-muted-foreground font-mono">
                        <Clock className="h-3 w-3" />
                        {log.durationMs}ms
                    </div>
                    {expanded ? <ChevronDown className="h-4 w-4 text-muted-foreground" /> : <ChevronRight className="h-4 w-4 text-muted-foreground" />}
                </div>
            </button>

            {expanded && (
                <div className="p-4 pt-0 bg-muted/10 border-t border-white/5 space-y-4 animate-in slide-in-from-top-2 duration-200">
                    <div className="space-y-1.5">
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Arguments</h4>
                        <div className="bg-background rounded-md border p-2 overflow-x-auto">
                            <pre className="text-[10px] font-mono whitespace-pre-wrap">
                                {typeof parsedArgs === 'object' ? JSON.stringify(parsedArgs, null, 2) : parsedArgs}
                            </pre>
                        </div>
                    </div>
                    <div className="space-y-1.5">
                        <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            {isError ? "Error" : "Result"}
                        </h4>
                        <div className={cn("bg-background rounded-md border p-2 overflow-x-auto", isError && "border-destructive/30 bg-destructive/5")}>
                            {isError ? (
                                <pre className="text-[10px] font-mono whitespace-pre-wrap text-destructive">
                                    {log.error}
                                </pre>
                            ) : (
                                <pre className="text-[10px] font-mono whitespace-pre-wrap">
                                    {typeof parsedResult === 'object' ? JSON.stringify(parsedResult, null, 2) : parsedResult}
                                </pre>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
