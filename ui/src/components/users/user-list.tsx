/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
    MoreHorizontal,
    Pencil,
    Trash2,
    Key,
    Lock,
    Search,
    Shield,
    User as UserIcon
} from "lucide-react";
import { User } from "@proto/config/v1/user";

interface UserListProps {
    users: User[];
    isLoading?: boolean;
    onEdit: (user: User) => void;
    onDelete: (id: string) => void;
}

export function UserList({ users, isLoading, onEdit, onDelete }: UserListProps) {
    const [filter, setFilter] = useState("");

    const filteredUsers = useMemo(() => {
        if (!filter) return users;
        return users.filter(u =>
            u.id.toLowerCase().includes(filter.toLowerCase()) ||
            u.roles.some(r => r.toLowerCase().includes(filter.toLowerCase()))
        );
    }, [users, filter]);

    if (isLoading) {
        return (
            <div className="space-y-4">
                <div className="flex items-center space-x-2 w-full md:w-1/3">
                    <div className="h-9 w-full bg-muted animate-pulse rounded-md" />
                </div>
                <div className="border rounded-md">
                    <div className="h-12 border-b bg-muted/50 animate-pulse" />
                    {[...Array(3)].map((_, i) => (
                        <div key={i} className="h-16 border-b bg-muted/20 animate-pulse" />
                    ))}
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div className="relative w-full md:w-1/3">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Filter users..."
                        value={filter}
                        onChange={(e) => setFilter(e.target.value)}
                        className="pl-9"
                    />
                </div>
                <div className="text-sm text-muted-foreground">
                    {filteredUsers.length} users
                </div>
            </div>

            <div className="rounded-md border bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[80px]">Avatar</TableHead>
                            <TableHead>User ID</TableHead>
                            <TableHead>Roles</TableHead>
                            <TableHead>Auth Method</TableHead>
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
                                <UserRow
                                    key={user.id}
                                    user={user}
                                    onEdit={onEdit}
                                    onDelete={onDelete}
                                />
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}

function UserRow({ user, onEdit, onDelete }: { user: User, onEdit: (u: User) => void, onDelete: (id: string) => void }) {
    // Generate initials
    const initials = user.id.slice(0, 2).toUpperCase();

    // Determine auth type
    const authType = useMemo(() => {
        if (user.authentication?.apiKey) return "API Key";
        if (user.authentication?.basicAuth) return "Password";
        return "None";
    }, [user]);

    return (
        <TableRow>
            <TableCell>
                <Avatar className="h-9 w-9">
                    <AvatarImage src={`https://avatar.vercel.sh/${user.id}`} alt={user.id} />
                    <AvatarFallback>{initials}</AvatarFallback>
                </Avatar>
            </TableCell>
            <TableCell className="font-medium">
                <div className="flex flex-col">
                    <span>{user.id}</span>
                    <span className="text-xs text-muted-foreground font-mono truncate max-w-[200px]">
                        {authType === "API Key" ? "Service Account" : "Human User"}
                    </span>
                </div>
            </TableCell>
            <TableCell>
                <div className="flex flex-wrap gap-1">
                    {user.roles.map(role => (
                        <Badge key={role} variant={role === "admin" ? "default" : "secondary"} className="capitalize">
                            {role}
                        </Badge>
                    ))}
                    {user.roles.length === 0 && <span className="text-muted-foreground text-xs italic">No roles</span>}
                </div>
            </TableCell>
            <TableCell>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    {authType === "API Key" ? (
                        <>
                            <Key className="h-4 w-4 text-amber-500" />
                            <span>API Key</span>
                        </>
                    ) : authType === "Password" ? (
                        <>
                            <Lock className="h-4 w-4 text-blue-500" />
                            <span>Password</span>
                        </>
                    ) : (
                        <>
                            <UserIcon className="h-4 w-4" />
                            <span>No Auth</span>
                        </>
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
                        <DropdownMenuItem onClick={() => onEdit(user)}>
                            <Pencil className="mr-2 h-4 w-4" />
                            Edit User
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={() => onDelete(user.id)} className="text-destructive focus:text-destructive">
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete User
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </TableCell>
        </TableRow>
    );
}
