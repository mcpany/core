import { useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { JsonView } from "@/components/ui/json-view";
import { Button } from "@/components/ui/button";

interface TraceTableViewerProps {
    data: unknown;
}

export function TraceTableViewer({ data }: TraceTableViewerProps) {
    const [viewMode, setViewMode] = useState<"table" | "json">("table");

    if (!data || typeof data !== "object") {
        return <JsonView data={data} maxHeight={400} />;
    }

    let items: Record<string, unknown>[] = [];

    // Attempt to extract an array to tabulate
    if (Array.isArray(data)) {
        items = data.filter(item => typeof item === "object" && item !== null);
    } else {
        // Sometimes responses wrap the list in a key like 'tools' or 'resources'
        for (const value of Object.values(data as Record<string, unknown>)) {
            if (Array.isArray(value)) {
                const potentialItems = value.filter(item => typeof item === "object" && item !== null);
                if (potentialItems.length > 0) {
                    items = potentialItems;
                    break;
                }
            }
        }
    }

    if (items.length === 0 || viewMode === "json") {
        return (
            <div className="flex flex-col gap-2">
                {items.length > 0 && (
                    <div className="flex justify-end">
                        <Button variant="outline" size="sm" onClick={() => setViewMode("table")}>
                            Switch to Table View
                        </Button>
                    </div>
                )}
                <JsonView data={data} maxHeight={400} />
            </div>
        );
    }

    const columns = Array.from(new Set(items.flatMap(Object.keys)));

    return (
        <div className="flex flex-col gap-2">
            <div className="flex justify-between items-center bg-muted/30 px-2 py-1 rounded-t-md border border-b-0">
                <span className="text-xs font-medium text-muted-foreground">{items.length} items found</span>
                <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={() => setViewMode("json")}>
                    View Raw JSON
                </Button>
            </div>
            <div className="rounded-b-md border overflow-x-auto bg-card">
                <Table>
                    <TableHeader className="bg-muted/50">
                        <TableRow>
                            {columns.map(col => (
                                <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-3 py-2 h-8">
                                    {col}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {items.map((row, idx) => (
                            <TableRow key={row.name ? String(row.name) : (row.id ? String(row.id) : `row-${idx}`)} className="hover:bg-muted/50">
                                {columns.map(col => {
                                    const val = row[col];
                                    let displayVal = val;
                                    if (typeof val === 'object' && val !== null) {
                                        displayVal = JSON.stringify(val);
                                    } else if (typeof val === 'boolean') {
                                        displayVal = val ? "true" : "false";
                                    }
                                    return (
                                        <TableCell key={col} className="px-3 py-2 text-xs max-w-[250px] truncate" title={String(displayVal)}>
                                            {String(displayVal ?? "")}
                                        </TableCell>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
