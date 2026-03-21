import re

with open("ui/src/components/audit/audit-log-viewer.tsx", "r") as f:
    content = f.read()

# Replace imports
content = content.replace("import { Light as SyntaxHighlighter } from 'react-syntax-highlighter'\nimport json from 'react-syntax-highlighter/dist/esm/languages/hljs/json'\nimport vs2015 from 'react-syntax-highlighter/dist/esm/styles/hljs/vs2015'",
"import { JsonView } from \"@/components/ui/json-view\"\nimport { RichResultViewer } from \"@/components/playground/rich-result-viewer\"\nimport { Tabs, TabsContent, TabsList, TabsTrigger } from \"@/components/ui/tabs\"")

content = content.replace("SyntaxHighlighter.registerLanguage('json', json)\n", "")
content = content.replace("SyntaxHighlighter.registerLanguage('json', json);\n", "")

# Replace Arguments section
args_old = """                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vs2015}
                                        customStyle={{ margin: 0, fontSize: '12px' }}
                                    >
                                        {formatJson(selectedLog.arguments) || "{}"}
                                    </SyntaxHighlighter>
                                </div>"""

args_new = """                                <div className="rounded-md border p-2 bg-muted/50">
                                    <JsonView
                                        data={selectedLog.arguments ? JSON.parse(selectedLog.arguments) : {}}
                                        className="text-xs"
                                    />
                                </div>"""
content = content.replace(args_old, args_new)

# Replace Result section
res_old = """                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vs2015}
                                        customStyle={{ margin: 0, fontSize: '12px', maxHeight: '300px' }}
                                    >
                                        {formatJson(selectedLog.result) || (selectedLog.error ? "null" : "{}")}
                                    </SyntaxHighlighter>
                                </div>"""

res_new = """                                <Tabs defaultValue="rich" className="w-full">
                                    <TabsList className="mb-2">
                                        <TabsTrigger value="rich">Rich Result</TabsTrigger>
                                        <TabsTrigger value="raw">Raw JSON</TabsTrigger>
                                    </TabsList>
                                    <TabsContent value="rich" className="m-0 border rounded-md p-4 max-h-[400px] overflow-y-auto">
                                        {selectedLog.result ? (
                                            <RichResultViewer result={JSON.parse(selectedLog.result)} />
                                        ) : (
                                            <span className="text-muted-foreground text-sm">No result data available.</span>
                                        )}
                                    </TabsContent>
                                    <TabsContent value="raw" className="m-0 border rounded-md p-2 bg-muted/50 max-h-[400px] overflow-y-auto">
                                        <JsonView
                                            data={selectedLog.result ? JSON.parse(selectedLog.result) : (selectedLog.error ? null : {})}
                                            className="text-xs"
                                        />
                                    </TabsContent>
                                </Tabs>"""
content = content.replace(res_old, res_new)

with open("ui/src/components/audit/audit-log-viewer.tsx", "w") as f:
    f.write(content)
