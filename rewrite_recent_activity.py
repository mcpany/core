import re

with open('ui/src/components/dashboard/recent-activity-widget.tsx', 'r') as f:
    content = f.read()

# Make the outer container use the new styles
content = content.replace(
    '<Card className="col-span-3 bg-background/80 backdrop-blur-md border-muted/50 shadow-sm overflow-hidden flex flex-col">',
    '<Card className="col-span-3 bg-background/80 backdrop-blur-md shadow-lg border-border/50 overflow-hidden flex flex-col">'
)

# Header Title change
content = content.replace(
    '<Activity className="h-4 w-4 text-primary" />\n                Recent Activity',
    '<Activity className="h-4 w-4 text-blue-500" />\n                <span className="font-sans font-semibold tracking-tight">Recent Activity</span>'
)

# Replace the inner div wrapper
content = content.replace(
    '<div className="p-4 space-y-3 bg-gradient-to-b from-transparent via-background to-background">',
    '<div className="p-4 space-y-5 relative before:absolute before:inset-0 before:ml-8 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-border/50 before:to-transparent">'
)

# Replace the trace div
old_trace_div = """                    <div
                        key={trace.id}
                        className={cn(
                            "group relative rounded-xl border p-4 cursor-pointer transition-all duration-300 ease-in-out font-sans",
                            isExpanded ? "bg-card shadow-md scale-[1.01]" : "bg-card/50 hover:bg-card hover:shadow-sm hover:scale-[1.01]",
                            isError && "border-destructive/20 bg-destructive/5 hover:bg-destructive/10"
                        )}
                        onClick={() => toggleExpand(trace.id)}
                    >
                        <div className="flex items-start gap-4">
                            {/* Status Icon with Timeline Node Styling */}
                            <div className="relative mt-0.5 flex-shrink-0 z-10">
                                <div className={cn(
                                    "flex items-center justify-center w-6 h-6 rounded-full border bg-background shadow-sm transition-transform duration-300",
                                    isError ? "border-destructive text-destructive" : isSuccess && !hasResponseDiff ? "border-emerald-500 text-emerald-500" : "border-blue-500 text-blue-500",
                                    isExpanded && "scale-110"
                                )}>
                                    {isError ? (
                                        <XCircle className="h-3 w-3" />
                                    ) : hasResponseDiff ? (
                                        <Activity className="h-3 w-3" />
                                    ) : (
                                        <CheckCircle2 className="h-3 w-3" />
                                    )}
                                </div>
                                {/* Pulsating Ring for Active/Recent */}
                                {isExpanded && (
                                    <div className={cn(
                                        "absolute inset-0 rounded-full animate-ping opacity-20",
                                        isSuccess ? "bg-emerald-500" : isError ? "bg-destructive" : "bg-amber-500"
                                    )} />
                                )}
                            </div>

                            {/* Main Content */}
                            <div className="flex-1 min-w-0 flex flex-col justify-center">"""

new_trace_div = """                    <div key={trace.id} className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
                        {/* Icon / Dot indicator */}
                        <div className={cn(
                            "flex items-center justify-center w-8 h-8 rounded-full border-2 bg-background shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 shadow-sm relative z-10 transition-transform duration-300",
                            isError ? "border-destructive/50 text-destructive" : hasResponseDiff ? "border-blue-500/50 text-blue-500" : "border-emerald-500/50 text-emerald-500",
                            isExpanded && "scale-110 shadow-md ring-2 ring-offset-1 ring-offset-background",
                            isExpanded && isError ? "ring-destructive/30" : isExpanded && hasResponseDiff ? "ring-blue-500/30" : isExpanded ? "ring-emerald-500/30" : ""
                        )}>
                            {isError ? (
                                <XCircle className="h-4 w-4" />
                            ) : hasResponseDiff ? (
                                <Activity className="h-4 w-4" />
                            ) : (
                                <CheckCircle2 className="h-4 w-4" />
                            )}
                            {/* Pulsating Ring */}
                            {isExpanded && (
                                <div className={cn(
                                    "absolute inset-0 rounded-full animate-ping opacity-20",
                                    isError ? "bg-destructive" : hasResponseDiff ? "bg-blue-500" : "bg-emerald-500"
                                )} />
                            )}
                        </div>

                        {/* Card Content */}
                        <div
                            className={cn(
                                "w-[calc(100%-3rem)] md:w-[calc(50%-2rem)] rounded-xl border p-4 cursor-pointer transition-all duration-300 ease-in-out font-sans",
                                isExpanded ? "bg-background shadow-md scale-[1.02]" : "bg-background/50 hover:bg-background hover:shadow-sm hover:scale-[1.01]",
                                isError ? "border-destructive/20 bg-destructive/5" : hasResponseDiff ? "border-blue-500/20 bg-blue-500/5" : "border-emerald-500/20 bg-emerald-500/5"
                            )}
                            onClick={() => toggleExpand(trace.id)}
                        >
                            <div className="flex flex-col justify-center">"""

content = content.replace(old_trace_div, new_trace_div)

# Replace the closing tags for the trace div
old_trace_div_end = """                            </div>
                        </div>
                    </div>"""
new_trace_div_end = """                            </div>
                        </div>
                    </div>"""
content = content.replace(old_trace_div_end, new_trace_div_end)

# Change trace title color
content = content.replace(
    '<span className="text-sm font-semibold truncate tracking-tight">',
    '<span className={cn("text-sm font-semibold truncate tracking-tight", isError ? "text-destructive" : "text-foreground")}>'
)

# Update expanded payload grids and colors
content = content.replace(
    'isExpanded ? "grid-rows-[1fr] mt-3 opacity-100" : "grid-rows-[0fr] opacity-0"',
    'isExpanded ? "grid-rows-[1fr] mt-4 opacity-100" : "grid-rows-[0fr] opacity-0"'
)

# Request Payload
content = content.replace(
    '<div className="text-xs font-medium text-muted-foreground flex items-center gap-1.5">',
    '<div className="text-[11px] uppercase tracking-wider font-semibold text-muted-foreground flex items-center gap-1.5">'
)
content = content.replace(
    '<div className="bg-muted/50 rounded-md border border-border/50 overflow-hidden">',
    '<div className="bg-background/80 backdrop-blur rounded-lg border border-border/50 overflow-hidden shadow-sm font-mono text-xs">'
)

# Error Message
content = content.replace(
    '<div className="text-xs font-medium text-destructive flex items-center gap-1.5">',
    '<div className="text-[11px] uppercase tracking-wider font-semibold text-destructive flex items-center gap-1.5">'
)
content = content.replace(
    '<div className="bg-destructive/10 rounded-md p-2 border border-destructive/20 overflow-x-auto text-destructive">',
    '<div className="bg-destructive/10 rounded-lg p-3 border border-destructive/20 overflow-x-auto text-destructive shadow-sm">'
)
content = content.replace(
    '<pre className="text-[11px] font-mono whitespace-pre-wrap break-all">',
    '<pre className="text-xs font-mono whitespace-pre-wrap break-all leading-relaxed">'
)

# Response Payload
content = content.replace(
    '<div className="text-xs font-medium text-emerald-600 dark:text-emerald-400 flex items-center gap-1.5">',
    '<div className="text-[11px] uppercase tracking-wider font-semibold text-emerald-600 dark:text-emerald-500 flex items-center gap-1.5">'
)
content = content.replace(
    '<div className="bg-emerald-500/5 rounded-md border border-emerald-500/20 overflow-hidden">',
    '<div className="bg-background/80 backdrop-blur rounded-lg border border-emerald-500/20 overflow-hidden shadow-sm font-mono text-xs">'
)

# Response Diff
content = content.replace(
    '<div className="text-xs font-medium text-blue-600 dark:text-blue-400 flex items-center gap-1.5">',
    '<div className="text-[11px] uppercase tracking-wider font-semibold text-blue-600 dark:text-blue-400 flex items-center gap-1.5">'
)
content = content.replace(
    '<div className="bg-background rounded-md border border-border overflow-hidden">',
    '<div className="bg-background rounded-lg border border-border overflow-hidden shadow-sm">'
)
content = content.replace(
    '<pre className="text-[11px] font-mono whitespace-pre-wrap break-all m-0">',
    '<pre className="text-xs font-mono whitespace-pre-wrap break-all m-0 leading-relaxed">'
)

# Replace diff rendering items
old_diff = """                                                            {String(trace.rootSpan.attributes['mcp.response_diff']).split('\\n').map((line, i) => {
                                                                if (line.startsWith('+')) {
                                                                    return <div key={i} className="bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 px-2 py-0.5">{line}</div>;
                                                                }
                                                                if (line.startsWith('-')) {
                                                                    return <div key={i} className="bg-destructive/20 text-destructive px-2 py-0.5">{line}</div>;
                                                                }
                                                                if (line.startsWith('@')) {
                                                                    return <div key={i} className="bg-blue-500/10 text-blue-600 px-2 py-0.5 opacity-70">{line}</div>;
                                                                }
                                                                return <div key={i} className="px-2 py-0.5 opacity-80">{line}</div>;
                                                            })}"""
new_diff = """                                                            {String(trace.rootSpan.attributes['mcp.response_diff']).split('\\n').map((line, i) => {
                                                                if (line.startsWith('+')) {
                                                                    return <div key={i} className="bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 px-3 py-1 border-l-2 border-emerald-500">{line}</div>;
                                                                }
                                                                if (line.startsWith('-')) {
                                                                    return <div key={i} className="bg-destructive/20 text-destructive px-3 py-1 border-l-2 border-destructive">{line}</div>;
                                                                }
                                                                if (line.startsWith('@')) {
                                                                    return <div key={i} className="bg-blue-500/10 text-blue-600 px-3 py-1 opacity-70">{line}</div>;
                                                                }
                                                                return <div key={i} className="px-3 py-1 opacity-80 text-foreground/80">{line}</div>;
                                                            })}"""

content = content.replace(old_diff, new_diff)

with open('ui/src/components/dashboard/recent-activity-widget.tsx', 'w') as f:
    f.write(content)
