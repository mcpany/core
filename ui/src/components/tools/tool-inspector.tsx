
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { ToolDefinition } from "@/lib/client";
import { ScrollArea } from "@/components/ui/scroll-area";
import SyntaxHighlighter from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/esm/styles/hljs';

interface ToolInspectorProps {
    tool: ToolDefinition | null;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
    if (!tool) return null;

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="w-[400px] sm:w-[640px]">
                <SheetHeader>
                    <SheetTitle className="flex items-center space-x-2">
                        <span>{tool.name}</span>
                        <Badge variant={tool.enabled ? "default" : "secondary"}>
                            {tool.enabled ? "Active" : "Disabled"}
                        </Badge>
                    </SheetTitle>
                    <SheetDescription>
                        {tool.description}
                    </SheetDescription>
                </SheetHeader>

                <div className="py-6 space-y-6">
                    <div>
                        <h4 className="text-sm font-medium mb-2">Service Origin</h4>
                        <div className="bg-muted p-2 rounded text-sm font-mono">{tool.serviceName}</div>
                    </div>

                    <div>
                        <h4 className="text-sm font-medium mb-2">Input Schema</h4>
                        <ScrollArea className="h-[400px] rounded border">
                            <SyntaxHighlighter language="json" style={docco} customStyle={{ background: 'transparent', padding: '1rem' }}>
                                {JSON.stringify(tool.schema || tool.input_schema, null, 2)}
                            </SyntaxHighlighter>
                        </ScrollArea>
                    </div>
                </div>
            </SheetContent>
        </Sheet>
    );
}
