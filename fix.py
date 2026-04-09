import re

with open('ui/src/components/ui/json-view.tsx', 'r') as f:
    content = f.read()

# Replace the specific block
search_block = """    const renderRaw = () => (
        <div className={cn("relative group/code rounded-md bg-[#1e1e1e]", className)}>
            <div
                className={cn(
                    "overflow-hidden transition-all",
                    showCollapse && !isExpanded ? "relative" : ""
                )}
                style={{
                    maxHeight: showCollapse && !isExpanded ? `${maxHeight}px` : undefined
                }}
            >
                <Suspense fallback={<pre className="p-4 text-xs text-muted-foreground">Loading…</pre>}>
                    <SyntaxHighlighter
                        language="json"
                        style={vs2015}
                        customStyle={{
                            margin: 0,
                            padding: '1rem',
                            borderRadius: '0.375rem',
                            fontSize: '12px',
                            lineHeight: '1.5',
                            backgroundColor: 'transparent' // We set bg on parent
                        }}
                        wrapLines={true}
                        wrapLongLines={true}
                    >
                        {typeof data === 'string' ? data : JSON.stringify(data, null, 2)}
                    </SyntaxHighlighter>
                </Suspense>

                {showCollapse && !isExpanded && (
                    <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-[#1e1e1e] to-transparent pointer-events-none" />
                )}
            </div>

            <div className="absolute right-2 top-2 flex gap-1 opacity-0 group-hover/code:opacity-100 transition-opacity">
                <Button
                    size="icon"
                    variant="ghost"
                    className="h-6 w-6 bg-white/10 hover:bg-white/20 text-white"
                    onClick={handleCopy}
                    title="Copy JSON"
                >
                    {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                </Button>
            </div>

            {showCollapse && renderCollapseButton()}
        </div>
    );"""

replace_block = """    const renderRaw = () => {
        if (typeof data === 'string') {
            return (
                <div className={cn("relative group/code rounded-md bg-[#1e1e1e]", className)}>
                    <div
                        className={cn(
                            "overflow-hidden transition-all",
                            showCollapse && !isExpanded ? "relative" : ""
                        )}
                        style={{
                            maxHeight: showCollapse && !isExpanded ? `${maxHeight}px` : undefined
                        }}
                    >
                        <pre className="p-4 text-sm text-foreground font-mono whitespace-pre-wrap break-all m-0 text-white">
                            {data}
                        </pre>
                        {showCollapse && !isExpanded && (
                            <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-[#1e1e1e] to-transparent pointer-events-none" />
                        )}
                    </div>
                    <div className="absolute right-2 top-2 flex gap-1 opacity-0 group-hover/code:opacity-100 transition-opacity">
                        <Button
                            size="icon"
                            variant="ghost"
                            className="h-6 w-6 bg-white/10 hover:bg-white/20 text-white"
                            onClick={handleCopy}
                            title="Copy string"
                        >
                            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                        </Button>
                    </div>
                    {showCollapse && renderCollapseButton()}
                </div>
            );
        }

        return (
            <div className={cn("relative group/code rounded-md bg-[#1e1e1e]", className)}>
                <div
                    className={cn(
                        "overflow-hidden transition-all",
                        showCollapse && !isExpanded ? "relative" : ""
                    )}
                    style={{
                        maxHeight: showCollapse && !isExpanded ? `${maxHeight}px` : undefined
                    }}
                >
                    <Suspense fallback={<pre className="p-4 text-xs text-muted-foreground">Loading…</pre>}>
                        <SyntaxHighlighter
                            language="json"
                            style={vs2015}
                            customStyle={{
                                margin: 0,
                                padding: '1rem',
                                borderRadius: '0.375rem',
                                fontSize: '12px',
                                lineHeight: '1.5',
                                backgroundColor: 'transparent' // We set bg on parent
                            }}
                            wrapLines={true}
                            wrapLongLines={true}
                        >
                            {JSON.stringify(data, null, 2)}
                        </SyntaxHighlighter>
                    </Suspense>

                    {showCollapse && !isExpanded && (
                        <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-[#1e1e1e] to-transparent pointer-events-none" />
                    )}
                </div>

                <div className="absolute right-2 top-2 flex gap-1 opacity-0 group-hover/code:opacity-100 transition-opacity">
                    <Button
                        size="icon"
                        variant="ghost"
                        className="h-6 w-6 bg-white/10 hover:bg-white/20 text-white"
                        onClick={handleCopy}
                        title="Copy JSON"
                    >
                        {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                    </Button>
                </div>

                {showCollapse && renderCollapseButton()}
            </div>
        );
    };"""

content = content.replace(search_block, replace_block)

with open('ui/src/components/ui/json-view.tsx', 'w') as f:
    f.write(content)
