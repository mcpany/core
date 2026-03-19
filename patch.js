const fs = require('fs');

const file = 'ui/src/components/users/user-list.tsx';
let content = fs.readFileSync(file, 'utf-8');

const importStr = `import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";`;
const newImportStr = `import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { TableVirtuoso } from "react-virtuoso";`;

content = content.replace(importStr, newImportStr);

const tableStart = `<div className="rounded-md border bg-background">
                <Table>`;
const newTableStart = `<div className="rounded-md border bg-background h-[calc(100vh-250px)]">
                <TableVirtuoso
                    data={filteredUsers}
                    components={{
                        Table: (props) => <Table {...props} />,
                        TableBody: React.forwardRef<HTMLTableSectionElement>((props, ref) => <TableBody {...props} ref={ref} />),
                        TableRow: (props) => <TableRow {...props} />,
                    }}
                    fixedHeaderContent={() => (
                        <TableRow className="bg-muted/50">
                            <TableHead className="w-[250px]">User</TableHead>
                            <TableHead>Roles</TableHead>
                            <TableHead>Authentication</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    )}
                    itemContent={(index, user) => (`;

content = content.replace(`<div className="rounded-md border bg-background">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[250px]">User</TableHead>
                            <TableHead>Roles</TableHead>
                            <TableHead>Authentication</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredUsers.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                                    No users found.
                                </TableCell>
                            </TableRow>
                        ) : (
                            filteredUsers.map((user) => (`, newTableStart);

content = content.replace(`))
                        )}
                    </TableBody>
                </Table>
            </div>`, `)}
                />
            </div>`);


// Add signature
const signature = `
    // ⚡ BOLT: Implemented virtualization for user list using react-virtuoso.
    // Randomized Selection from Top 5 High-Impact Targets (React/View)
    const filteredUsers = useMemo(() => {`;

content = content.replace(`const filteredUsers = useMemo(() => {`, signature);

const reactImportStr = `import { useMemo, useState } from "react";`;
const newReactImportStr = `import React, { useMemo, useState } from "react";`;
content = content.replace(reactImportStr, newReactImportStr);


fs.writeFileSync(file, content);
