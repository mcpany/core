/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState } from "react";
import { User } from "./user-sheet";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { MoreHorizontal, Search, Trash2, Edit2, Shield, User as UserIcon, Key, Lock, Copy } from "lucide-react";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";

interface UserListProps {
    users: User[];
    loading?: boolean;
    onEdit: (user: User) => void;
    onDelete: (id: string) => void;
}

/**
 * UserList component displays a sortable and filterable list of users.
 *
 * @param props - The component props.
 * @param props.users - The list of users to display.
 * @param props.loading - Whether the data is loading.
 * @param props.onEdit - Callback for editing a user.
 * @param props.onDelete - Callback for deleting a user.
 * @returns The rendered component.
 */
export function UserList({ users, loading, onEdit, onDelete }: UserListProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const { toast } = useToast();

    const filteredUsers = useMemo(() => {
        return users.filter(u =>
            u.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
            u.roles.some(r => r.toLowerCase().includes(searchQuery.toLowerCase()))
        );
    }, [users, searchQuery]);

    const getInitials = (name: string) => {
        return name.substring(0, 2).toUpperCase();
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast({
            title: "Copied",
            description: "User ID copied to clipboard.",
        });
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center gap-2">
                <div className="relative flex-1 max-w-sm">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search users..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8 bg-background"
                    />
                </div>
            </div>

            <div className="rounded-md border bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[250px]">User</TableHead>
                            <TableHead>Roles</TableHead>
                            <TableHead>Auth Method</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground animate-pulse">
                                    Loading users...
                                </TableCell>
                            </TableRow>
                        ) : filteredUsers.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                    No users found matching "{searchQuery}".
                                </TableCell>
                            </TableRow>
                        ) : (
                            filteredUsers.map((user) => (
                                <TableRow key={user.id} className="group">
                                    <TableCell className="font-medium">
                                        <div className="flex items-center gap-3">
                                            <Avatar className="h-9 w-9 border-2 border-border">
                                                <AvatarImage src={`https://api.dicebear.com/9.x/notionists/svg?seed=${user.id}`} alt={user.id} />
                                                <AvatarFallback>{getInitials(user.id)}</AvatarFallback>
                                            </Avatar>
                                            <div className="flex flex-col">
                                                <span className="font-semibold">{user.id}</span>
                                                <span className="text-[10px] text-muted-foreground font-mono opacity-0 group-hover:opacity-100 transition-opacity">
                                                    ID: {user.id}
                                                </span>
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex flex-wrap gap-1">
                                            {user.roles.map((role) => (
                                                <Badge
                                                    key={role}
                                                    variant={role === "admin" ? "default" : "secondary"}
                                                    className={cn(
                                                        "text-xs px-2 py-0.5",
                                                        role === "admin" && "bg-primary text-primary-foreground"
                                                    )}
                                                >
                                                    {role === "admin" && <Shield className="mr-1 h-3 w-3" />}
                                                    {role.charAt(0).toUpperCase() + role.slice(1)}
                                                </Badge>
                                            ))}
                                            {user.roles.length === 0 && (
                                                <Badge variant="outline" className="text-xs text-muted-foreground">No Roles</Badge>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                            {user.authentication?.api_key ? (
                                                <div className="flex items-center gap-1.5" title="API Key">
                                                    <div className="p-1.5 bg-green-500/10 rounded-full text-green-600">
                                                        <Key className="h-3.5 w-3.5" />
                                                    </div>
                                                    <span>API Key</span>
                                                </div>
                                            ) : user.authentication?.basic_auth ? (
                                                <div className="flex items-center gap-1.5" title="Password">
                                                    <div className="p-1.5 bg-blue-500/10 rounded-full text-blue-600">
                                                        <Lock className="h-3.5 w-3.5" />
                                                    </div>
                                                    <span>Password</span>
                                                </div>
                                            ) : (
                                                <div className="flex items-center gap-1.5 opacity-50">
                                                    <UserIcon className="h-3.5 w-3.5" />
                                                    <span>None</span>
                                                </div>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="outline" className="border-green-500/50 text-green-600 bg-green-500/10">Active</Badge>
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
                                                    <Edit2 className="mr-2 h-4 w-4" />
                                                    Edit User
                                                </DropdownMenuItem>
                                                <DropdownMenuItem
                                                    className="text-destructive focus:text-destructive focus:bg-destructive/10"
                                                    onClick={() => onDelete(user.id)}
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
            {!loading && filteredUsers.length > 0 && (
                <div className="text-xs text-muted-foreground text-center">
                    Showing {filteredUsers.length} users.
                </div>
            )}
        </div>
    );
}
