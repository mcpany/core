const fs = require('fs');
const file = 'ui/src/components/playground/tool-runner.tsx';
let content = fs.readFileSync(file, 'utf8');

if (!content.includes('ToolRunHistoryItem')) {
    const importStatment = `import { ToolRunHistoryItem } from "./tool-run-history-item";\n`;
    content = content.replace('import { SchemaViewer }', importStatment + 'import { SchemaViewer }');

    const oldHistorySection = `<div className="rounded-md border divide-y">
                            {recentStats.chartData.length === 0 && (
                                <div className="text-xs text-muted-foreground p-4 text-center">
                                    No recent executions.
                                </div>
                            )}
                            {[...recentStats.chartData].reverse().slice(0, 10).map((h, i) => (
                                <div key={i} className="flex items-center justify-between text-sm p-3 hover:bg-muted/50 transition-colors">
                                    <div className="flex items-center gap-3">
                                        <div className={cn("h-2.5 w-2.5 rounded-full", h.status === "success" ? "bg-green-500" : "bg-destructive")} />
                                        <span className="font-medium">{h.time}</span>
                                    </div>
                                    <span className="text-muted-foreground font-mono text-xs">{h.latency}ms</span>
                                </div>
                            ))}
                        </div>`;

    const newHistorySection = `<div className="rounded-xl border border-white/10 shadow-sm backdrop-blur-md bg-card/50 overflow-hidden transition-all divide-y divide-border/50">
                            {auditLogs.length === 0 && (
                                <div className="flex flex-col items-center justify-center p-8 text-muted-foreground space-y-3 opacity-70">
                                    <HistoryIcon className="h-8 w-8 stroke-[1]" />
                                    <p className="text-sm">No execution history found.</p>
                                </div>
                            )}
                            {auditLogs.slice(0, 20).map((log, i) => (
                                <ToolRunHistoryItem key={i} log={log as any} />
                            ))}
                        </div>`;

    content = content.replace(oldHistorySection, newHistorySection);

    // Also patch the label to say Execution History so the test passes
    content = content.replace('<HistoryIcon className="h-4 w-4" /> Recent Timeline', '<HistoryIcon className="h-4 w-4" /> Execution History');

    fs.writeFileSync(file, content);
    console.log("Patched tool-runner.tsx");
}
