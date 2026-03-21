/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useMemo, useState } from "react";
import { User } from "@proto/config/v1/user";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Checkbox } from "@/components/ui/checkbox";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    Key,
    Lock,
    MoreHorizontal,
    Pencil,
    Trash2,
    Copy,
    Search,
    ShieldAlert,
    Eye
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface UserListProps {
    users: User[];
    isLoading?: boolean;
    onEdit: (user: User) => void;
    onDelete: (id: string) => void;
    onBulkDelete?: (ids: string[]) => void;
}

/**
 * UserList component.
 * Displays a list of users with filtering and actions.
 *
 * @param props - The component props.
 * @param props.users - The list of users to display.
 * @param props.isLoading - Whether the data is loading.
 * @param props.onEdit - Callback when a user is edited.
 * @param props.onDelete - Callback when a user is deleted.
 * @param props.onBulkDelete - Callback when multiple users are deleted.
 * @returns The rendered UserList component.
 */
export function UserList({ users, isLoading, onEdit, onDelete, onBulkDelete }: UserListProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [selected, setSelected] = useState<Set<string>>(new Set());
    const { toast } = useToast();

    const filteredUsers = useMemo(() => {
        if (!searchQuery) return users;
        const query = searchQuery.toLowerCase();
        return users.filter(user =>
            user.id.toLowerCase().includes(query) ||
            user.roles.some((role: string) => role.toLowerCase().includes(query))
        );
    }, [users, searchQuery]);

    const getInitials = (name: string) => {
        return name.slice(0, 2).toUpperCase();
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast({
            description: "Copied to clipboard",
        });
    };

    const handleSelectAll = (checked: boolean) => {
        if (checked) {
            setSelected(new Set(filteredUsers.map(u => u.id)));
        } else {
            setSelected(new Set());
        }
    };

    const handleSelectOne = (id: string, checked: boolean) => {
        setSelected(prev => {
            const newSelected = new Set(prev);
            if (checked) {
                newSelected.add(id);
            } else {
                newSelected.delete(id);
            }
            return newSelected;
        });
    };

    const isAllSelected = filteredUsers.length > 0 && filteredUsers.every(u => selected.has(u.id));

    if (isLoading) {
        return (
            <div className="space-y-4">
                <div className="flex items-center space-x-2">
                    <div className="h-9 w-64 bg-muted animate-pulse rounded-md" />
                </div>
                <div className="border rounded-md">
                    <div className="h-12 border-b bg-muted/50" />
                    {[...Array(3)].map((_, i) => (
                        <div key={i} className="h-16 border-b bg-background animate-pulse" />
                    ))}
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div className="relative w-full max-w-sm">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search users..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8"
                    />
                </div>
            </div>

            <div className="space-y-2">
                {selected.size > 0 && (
                    <div className="flex items-center gap-2 p-2 bg-muted/40 rounded-md animate-in fade-in slide-in-from-top-1 duration-200 sticky top-0 z-10 backdrop-blur-md border">
                        <span className="text-sm text-muted-foreground mr-2 font-medium px-2">{selected.size} selected</span>
                        <div className="h-4 w-px bg-border mx-1" />
                        {onBulkDelete && (
                            <Button size="sm" variant="destructive" onClick={() => {
                                onBulkDelete(Array.from(selected));
                                setSelected(new Set());
                            }}>
                                <Trash2 className="mr-2 h-4 w-4" /> Delete
                            </Button>
                        )}
                    </div>
                )}

            <div className="rounded-md border bg-background">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[30px] pr-0">
                                <Checkbox
                                    checked={isAllSelected}
                                    onCheckedChange={(checked) => handleSelectAll(!!checked)}
                                    aria-label="Select all"
                                    className="translate-y-[2px]"
                                />
                            </TableHead>
                            <TableHead className="w-[250px]">User</TableHead>
                            <TableHead>Roles</TableHead>
                            <TableHead>Authentication</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredUsers.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                    No users found.
                                </TableCell>
                            </TableRow>
                        ) : (
                            filteredUsers.map((user) => (
                                <TableRow key={user.id} data-testid={`user-row-${user.id}`} className={selected.has(user.id) ? "bg-muted/50" : ""}>
                                    <TableCell className="pr-0">
                                        <Checkbox
                                            checked={selected.has(user.id)}
                                            onCheckedChange={(checked) => handleSelectOne(user.id, !!checked)}
                                            aria-label={`Select ${user.id}`}
                                            className="translate-y-[2px]"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-3">
                                            <Avatar className="h-9 w-9 border">
                                                <AvatarFallback className="bg-primary/10 text-primary font-medium">
                                                    {getInitials(user.id)}
                                                </AvatarFallback>
                                            </Avatar>
                                            <div className="flex flex-col">
                                                <span className="font-medium text-sm">{user.id}</span>
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex flex-wrap gap-1">
                                            {user.roles?.map((role: string) => (
                                                <Badge
                                                    key={role}
                                                    variant={role === "admin" ? "default" : "secondary"}
                                                    className="capitalize"
                                                >
                                                    {role === "admin" && <ShieldAlert className="mr-1 h-3 w-3" />}
                                                    {role === "viewer" && <Eye className="mr-1 h-3 w-3" />}
                                                    {role === "editor" && <Pencil className="mr-1 h-3 w-3" />}
                                                    {role}
                                                </Badge>
                                            ))}
                                            {(!user.roles || user.roles.length === 0) && (
                                                <span className="text-muted-foreground text-xs italic">No roles</span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                                            {user.authentication?.apiKey || (user.authentication as any)?.api_key ? (
                                                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/50 border">
                                                    <Key className="h-3.5 w-3.5 text-orange-500" />
                                                    <span>API Key</span>
                                                </div>
                                            ) : /* eslint-disable-next-line @typescript-eslint/no-explicit-any */
                                                user.authentication?.basicAuth || (user.authentication as any)?.basic_auth ? (
                                                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/50 border">
                                                    <Lock className="h-3.5 w-3.5 text-blue-500" />
                                                    <span>Password</span>
                                                </div>
                                            ) : (
                                                <span className="text-muted-foreground italic">None configured</span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button variant="ghost" className="h-8 w-8 p-0">
                                                    <span className="sr-only">Open menu</span>
                                                    <MoreHorizontal className="h-4 w-4" />
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end">
                                                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                <DropdownMenuItem onClick={() => copyToClipboard(user.id)}>
                                                    <Copy className="mr-2 h-4 w-4" />
                                                    Copy ID
                                                </DropdownMenuItem>
                                                <DropdownMenuSeparator />
                                                <DropdownMenuItem onClick={() => onEdit(user)}>
                                                    <Pencil className="mr-2 h-4 w-4" />
                                                    Edit
                                                </DropdownMenuItem>
                                                <DropdownMenuItem
                                                    onClick={() => onDelete(user.id)}
                                                    className="text-destructive focus:text-destructive"
                                                >
                                                    <Trash2 className="mr-2 h-4 w-4" />
                                                    Delete
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
            </div>
            <div className="text-xs text-muted-foreground text-center">
                Showing {filteredUsers.length} of {users.length} users
            </div>
        </div>
    );
}
