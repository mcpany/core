/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState } from "react";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
    MoreHorizontal,
    Search,
    User as UserIcon,
    Key,
    Lock,
    Shield,
    Trash2,
    Edit,
    Copy,
    UserPlus
} from "lucide-react";
import { cn } from "@/lib/utils";

// Define the User interface locally as it's not exported from client.ts
/**
 * User interface definition.
 */
export interface User {
    id: string;
    roles: string[];
    authentication?: {
        basic_auth?: Record<string, any>;
        api_key?: {
            param_name?: string;
            in?: number;
            verification_value?: string;
        };
    };
    profile_ids?: string[];
}

interface UserListProps {
    users: User[];
    loading?: boolean;
    onEdit: (user: User) => void;
    onDelete: (id: string) => void;
}

/**
 * UserList component.
 * @param props - The component props.
 * @param props.users - The list of users to display.
 * @param props.loading - Whether the list is loading.
 * @param props.onEdit - Callback for editing a user.
 * @param props.onDelete - Callback for deleting a user.
 * @returns The rendered component.
 */
export function UserList({ users, loading, onEdit, onDelete }: UserListProps) {
    const [searchQuery, setSearchQuery] = useState("");

    const filteredUsers = useMemo(() => {
        if (!searchQuery) return users;
        return users.filter(user =>
            user.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
            user.roles.some(role => role.toLowerCase().includes(searchQuery.toLowerCase()))
        );
    }, [users, searchQuery]);

    const getInitials = (name: string) => {
        return name.slice(0, 2).toUpperCase();
    };

    const getAuthIcon = (user: User) => {
        if (user.authentication?.api_key) {
            return <Key className="h-4 w-4 text-orange-500" />;
        }
        if (user.authentication?.basic_auth) {
            return <Lock className="h-4 w-4 text-blue-500" />;
        }
        return <UserIcon className="h-4 w-4 text-muted-foreground" />;
    };

    const getAuthLabel = (user: User) => {
        if (user.authentication?.api_key) return "API Key";
        if (user.authentication?.basic_auth) return "Password";
        return "No Auth";
    };

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

            <div className="rounded-md border bg-card">
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
                        {loading ? (
                            <TableRow>
                                <TableCell colSpan={4} className="h-24 text-center">
                                    <div className="flex items-center justify-center gap-2 text-muted-foreground">
                                        Loading users...
                                    </div>
                                </TableCell>
                            </TableRow>
                        ) : filteredUsers.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={4} className="h-32 text-center">
                                    <div className="flex flex-col items-center justify-center gap-2 text-muted-foreground">
                                        <UserIcon className="h-8 w-8 opacity-20" />
                                        <p>No users found matching your criteria.</p>
                                    </div>
                                </TableCell>
                            </TableRow>
                        ) : (
                            filteredUsers.map((user) => (
                                <TableRow key={user.id} className="group">
                                    <TableCell>
                                        <div className="flex items-center gap-3">
                                            <Avatar>
                                                <AvatarImage src={`https://avatar.vercel.sh/${user.id}.png`} />
                                                <AvatarFallback>{getInitials(user.id)}</AvatarFallback>
                                            </Avatar>
                                            <div className="flex flex-col">
                                                <span className="font-medium">{user.id}</span>
                                                <span className="text-xs text-muted-foreground font-mono opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1 cursor-pointer" onClick={() => navigator.clipboard.writeText(user.id)}>
                                                    {user.id} <Copy className="h-3 w-3" />
                                                </span>
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex flex-wrap gap-1">
                                            {user.roles.map(role => (
                                                <Badge key={role} variant={role === 'admin' ? "default" : "secondary"} className="gap-1">
                                                    {role === 'admin' && <Shield className="h-3 w-3" />}
                                                    {role}
                                                </Badge>
                                            ))}
                                            {user.roles.length === 0 && (
                                                <span className="text-muted-foreground text-xs italic">No roles</span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2 text-sm">
                                            {getAuthIcon(user)}
                                            <span>{getAuthLabel(user)}</span>
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
                                                <DropdownMenuItem onClick={() => navigator.clipboard.writeText(user.id)}>
                                                    <Copy className="mr-2 h-4 w-4" /> Copy ID
                                                </DropdownMenuItem>
                                                <DropdownMenuSeparator />
                                                <DropdownMenuItem onClick={() => onEdit(user)}>
                                                    <Edit className="mr-2 h-4 w-4" /> Edit User
                                                </DropdownMenuItem>
                                                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onDelete(user.id)}>
                                                    <Trash2 className="mr-2 h-4 w-4" /> Delete User
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
