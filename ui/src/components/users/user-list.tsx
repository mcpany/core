/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState } from "react";
import { User } from "@proto/config/v1/user";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
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
}

/**
 * UserList component.
 * Displays a table of users with filtering and actions.
 *
 * @param props - The component props.
 * @param props.users - The list of users to display.
 * @param props.isLoading - Whether the data is loading.
 * @param props.onEdit - Callback when the edit button is clicked.
 * @param props.onDelete - Callback when the delete button is clicked.
 * @returns The rendered component.
 */
export function UserList({ users, isLoading, onEdit, onDelete }: UserListProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const { toast } = useToast();

    const filteredUsers = useMemo(() => {
        if (!searchQuery) return users;
        const query = searchQuery.toLowerCase();
        return users.filter(user =>
            user.id.toLowerCase().includes(query) ||
            user.roles.some(role => role.toLowerCase().includes(query))
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

            <div className="rounded-md border bg-background">
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
                            filteredUsers.map((user) => (
                                <TableRow key={user.id}>
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
                                            {user.roles.map((role) => (
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
                                            {user.roles.length === 0 && (
                                                <span className="text-muted-foreground text-xs italic">No roles</span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                            {user.authentication?.apiKey ? (
                                                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/50 border">
                                                    <Key className="h-3.5 w-3.5 text-orange-500" />
                                                    <span>API Key</span>
                                                </div>
                                            ) : user.authentication?.basicAuth ? (
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
            <div className="text-xs text-muted-foreground text-center">
                Showing {filteredUsers.length} of {users.length} users
            </div>
        </div>
    );
}
