import * as React from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronDown, ChevronUp, ChevronsUpDown, Search, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";

/**
 * DataTableProps
 */
export interface DataTableProps<TData> {
  data: TData[];
  columns: {
    accessorKey: keyof TData | string;
    header: string;
    cell?: (item: TData) => React.ReactNode;
  }[];
  searchKey?: keyof TData | string;
}

/**
 * DataTable
 */
export function DataTable<TData>({ data, columns, searchKey }: DataTableProps<TData>) {
  const [sortConfig, setSortConfig] = React.useState<{ key: string; direction: "asc" | "desc" } | null>(null);
  const [filterText, setFilterText] = React.useState("");
  const [currentPage, setCurrentPage] = React.useState(1);
  const pageSize = 10;

  const handleSort = (key: string) => {
    let direction: "asc" | "desc" = "asc";
    if (sortConfig && sortConfig.key === key && sortConfig.direction === "asc") {
      direction = "desc";
    }
    setSortConfig({ key, direction });
  };

  const filteredData = React.useMemo(() => {
    if (!filterText) return data;
    return data.filter((item) => {
      // If searchKey is specified, search only in that field. Otherwise search all string values.
      if (searchKey) {
        const value = item[searchKey as keyof TData];
        if (value == null) return false;
        return String(value).toLowerCase().includes(filterText.toLowerCase());
      }
      return Object.values(item as Record<string, unknown>).some(val =>
          String(val).toLowerCase().includes(filterText.toLowerCase())
      );
    });
  }, [data, filterText, searchKey]);

  const sortedData = React.useMemo(() => {
    if (!sortConfig) return filteredData;
    return [...filteredData].sort((a, b) => {
      const aVal = a[sortConfig.key as keyof TData];
      const bVal = b[sortConfig.key as keyof TData];

      if (aVal === bVal) return 0;
      if (aVal == null) return 1;
      if (bVal == null) return -1;

      if (sortConfig.direction === "asc") {
        return aVal < bVal ? -1 : 1;
      } else {
        return aVal > bVal ? -1 : 1;
      }
    });
  }, [filteredData, sortConfig]);

  const paginatedData = React.useMemo(() => {
    const startIndex = (currentPage - 1) * pageSize;
    return sortedData.slice(startIndex, startIndex + pageSize);
  }, [sortedData, currentPage]);

  const totalPages = Math.ceil(sortedData.length / pageSize);

  // Reset page when filter changes
  React.useEffect(() => {
      setCurrentPage(1);
  }, [filterText]);

  return (
    <div className="space-y-4 w-full">
      <div className="flex items-center space-x-2">
        <div className="relative w-full max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search results..."
            value={filterText}
            onChange={(e) => setFilterText(e.target.value)}
            className="pl-8 h-9 bg-background"
          />
        </div>
      </div>

      <div className="rounded-md border bg-card shadow-sm overflow-hidden w-full overflow-x-auto max-h-[500px]">
        <Table>
          <TableHeader className="bg-muted/50 sticky top-0 z-10 backdrop-blur-sm">
            <TableRow>
              {columns.map((col) => (
                <TableHead
                    key={String(col.accessorKey)}
                    className="whitespace-nowrap px-4 py-2 font-semibold text-xs text-muted-foreground uppercase tracking-wider group cursor-pointer select-none hover:text-foreground hover:bg-muted/80 transition-colors h-9"
                    onClick={() => handleSort(String(col.accessorKey))}
                >
                  <div className="flex items-center space-x-1">
                    <span>{col.header}</span>
                    <div className="ml-1 w-4 h-4 flex items-center justify-center">
                      {sortConfig?.key === String(col.accessorKey) ? (
                        sortConfig.direction === "asc" ? (
                          <ChevronUp className="h-3 w-3 text-primary" />
                        ) : (
                          <ChevronDown className="h-3 w-3 text-primary" />
                        )
                      ) : (
                        <ChevronsUpDown className="h-3 w-3 opacity-0 group-hover:opacity-50 transition-opacity" />
                      )}
                    </div>
                  </div>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedData.length > 0 ? (
              paginatedData.map((row, i) => (
                <TableRow key={i} className="hover:bg-muted/30 transition-colors group">
                  {columns.map((col) => {
                    const rawVal = row[col.accessorKey as keyof TData] ?? "";
                    const displayVal = col.cell ? col.cell(row) : String(rawVal);
                    // Generate a readable title string
                    let titleStr = "";
                    if (typeof rawVal === 'object' && rawVal !== null) {
                        try { titleStr = JSON.stringify(rawVal); } catch (_) { titleStr = String(rawVal); }
                    } else {
                        titleStr = String(rawVal);
                    }

                    return (
                        <TableCell
                            key={String(col.accessorKey)}
                            className="px-4 py-2 text-sm max-w-[200px] truncate group-hover:text-foreground text-muted-foreground transition-colors"
                            title={titleStr}
                        >
                          {displayVal}
                        </TableCell>
                    );
                  })}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                  No matching results found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-2">
          <div className="text-xs text-muted-foreground font-medium">
            Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, sortedData.length)} of {sortedData.length} rows
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="icon"
              className="h-7 w-7 bg-background"
              onClick={() => setCurrentPage(1)}
              disabled={currentPage === 1}
            >
              <ChevronsLeft className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-7 w-7 bg-background"
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
            >
              <ChevronLeft className="h-3.5 w-3.5" />
            </Button>
            <span className="text-xs font-medium px-2 min-w-[3rem] text-center text-muted-foreground">
              {currentPage} / {totalPages}
            </span>
            <Button
              variant="outline"
              size="icon"
              className="h-7 w-7 bg-background"
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
            >
              <ChevronRight className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-7 w-7 bg-background"
              onClick={() => setCurrentPage(totalPages)}
              disabled={currentPage === totalPages}
            >
              <ChevronsRight className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}

      {totalPages <= 1 && sortedData.length > 0 && (
         <div className="flex items-center justify-between px-2 pt-1">
             <div className="text-xs text-muted-foreground font-medium">
                 Showing all {sortedData.length} rows
             </div>
         </div>
      )}
    </div>
  );
}
