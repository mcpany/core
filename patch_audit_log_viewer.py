import re
with open("ui/src/components/audit/audit-log-viewer.tsx", "r") as f:
    content = f.read()

replacement = """
            <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <DialogContent className="max-w-4xl max-h-[85vh] flex flex-col p-0 overflow-hidden bg-background/95 backdrop-blur-xl border-muted/50 shadow-2xl">
                    <DialogHeader className="px-6 py-4 border-b bg-muted/10 shrink-0">
                        <div className="flex items-center justify-between">
                            <div>
                                <DialogTitle className="text-xl flex items-center gap-2">
                                    {selectedLog?.toolName}
                                    {selectedLog?.error ? (
                                        <Badge variant="destructive" className="ml-2 font-normal text-xs uppercase tracking-wider">Failed</Badge>
                                    ) : (
                                        <Badge variant="outline" className="ml-2 font-normal text-xs text-green-500 border-green-500/50 uppercase tracking-wider">Success</Badge>
                                    )}
                                </DialogTitle>
                                <DialogDescription className="mt-1">
                                    Executed at {selectedLog && new Date(selectedLog.timestamp).toLocaleString()}
                                </DialogDescription>
                            </div>
                        </div>
                    </DialogHeader>
                    {selectedLog && (
                        <div className="flex-1 overflow-y-auto p-6 space-y-6">
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 rounded-xl bg-muted/20 border border-border/50">
                                <div>
                                    <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground block mb-1">User</span>
                                    <span className="text-sm font-medium">{selectedLog.userId || "System"}</span>
                                </div>
                                <div>
                                    <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground block mb-1">Profile</span>
                                    <span className="text-sm font-medium">{selectedLog.profileId || "Default"}</span>
                                </div>
                                <div>
                                    <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground block mb-1">Duration</span>
                                    <span className="text-sm font-medium font-mono">{selectedLog.durationMs}ms</span>
                                </div>
                                <div>
                                    <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground block mb-1">Trace ID</span>
                                    <span className="text-sm font-mono text-muted-foreground truncate block" title={selectedLog.traceId || ""}>{selectedLog.traceId || "-"}</span>
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4 text-red-600 dark:text-red-400">
                                    <div className="flex items-start gap-3">
                                        <AlertTriangle className="h-5 w-5 mt-0.5 shrink-0" />
                                        <div>
                                            <h4 className="text-sm font-semibold mb-1">Execution Error</h4>
                                            <p className="text-sm whitespace-pre-wrap font-mono text-xs">{selectedLog.error}</p>
                                        </div>
                                    </div>
                                </div>
                            )}

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="space-y-3 flex flex-col">
                                    <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                                        <div className="h-px bg-border flex-1"></div>
                                        Request Payload
                                        <div className="h-px bg-border flex-1"></div>
                                    </h4>
                                    <div className="flex-1 rounded-xl overflow-hidden border border-border/50 bg-card shadow-sm min-h-[300px]">
                                        <RichResultViewer result={safeParse(selectedLog.arguments) || {}} />
                                    </div>
                                </div>

                                <div className="space-y-3 flex flex-col">
                                    <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                                        <div className="h-px bg-border flex-1"></div>
                                        Response Data
                                        <div className="h-px bg-border flex-1"></div>
                                    </h4>
                                    <div className="flex-1 rounded-xl overflow-hidden border border-border/50 bg-card shadow-sm min-h-[300px]">
                                        <RichResultViewer result={safeParse(selectedLog.result) || (selectedLog.error ? null : {})} />
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>
"""

new_content = re.sub(
    r'<Dialog open=\{!!selectedLog\} onOpenChange=\{\(open\) => !open && setSelectedLog\(null\)\}>.*?</Dialog>',
    replacement.strip(),
    content,
    flags=re.DOTALL
)

with open("ui/src/components/audit/audit-log-viewer.tsx", "w") as f:
    f.write(new_content)
