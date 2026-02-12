/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState } from "react";
import { User } from "@proto/config/v1/user";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Search, MoreHorizontal, Pencil, Trash2, Key, Lock, Copy } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface UserListProps {
  users: User[];
  isLoading?: boolean;
  onEdit: (user: User) => void;
  onDelete: (id: string) => void;
}

/**
 * UserList component for displaying a list of users.
 * Supports filtering, editing, and deleting users.
 *
 * @param props - The component props.
 * @param props.users - The list of users to display.
 * @param props.isLoading - Whether the data is loading.
 * @param props.onEdit - Callback when editing a user.
 * @param props.onDelete - Callback when deleting a user.
 * @returns The rendered component.
 */
export function UserList({ users, isLoading, onEdit, onDelete }: UserListProps) {
  const { toast } = useToast();
  const [searchQuery, setSearchQuery] = useState("");

  const filteredUsers = useMemo(() => {
    return users.filter((user) =>
      user.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.roles.some((role) => role.toLowerCase().includes(searchQuery.toLowerCase()))
    );
  }, [users, searchQuery]);

  const getInitials = (name: string) => {
    return name.substring(0, 2).toUpperCase();
  };

  const copyId = (id: string) => {
    navigator.clipboard.writeText(id);
    toast({ title: "Copied", description: "User ID copied to clipboard." });
  };

  const getAuthIcon = (user: User) => {
      // Check auth method safely
      const auth = user.authentication;
      if (!auth) return <Badge variant="outline" className="text-muted-foreground">None</Badge>;

      if (auth.apiKey) {
          return (
              <div className="flex items-center gap-1 text-xs text-muted-foreground bg-secondary/50 px-2 py-1 rounded-full">
                  <Key className="h-3 w-3" />
                  <span>API Key</span>
              </div>
          );
      }
      if (auth.basicAuth) {
          return (
              <div className="flex items-center gap-1 text-xs text-muted-foreground bg-secondary/50 px-2 py-1 rounded-full">
                  <Lock className="h-3 w-3" />
                  <span>Password</span>
              </div>
          );
      }
      return <Badge variant="outline">Unknown</Badge>;
  };

  if (isLoading) {
      return (
          <div className="space-y-4">
               {[...Array(3)].map((_, i) => (
                  <div key={i} className="w-full h-12 bg-muted animate-pulse rounded-md" />
               ))}
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
        <div className="text-sm text-muted-foreground">
            {filteredUsers.length} users
        </div>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[80px]">Avatar</TableHead>
              <TableHead>User</TableHead>
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
                    <TableRow key={user.id}>
                        <TableCell>
                            <Avatar className="h-9 w-9">
                                <AvatarFallback className="bg-primary/10 text-primary">
                                    {getInitials(user.id)}
                                </AvatarFallback>
                            </Avatar>
                        </TableCell>
                        <TableCell>
                            <div className="flex flex-col">
                                <span className="font-medium">{user.id}</span>
                            </div>
                        </TableCell>
                        <TableCell>
                            <div className="flex gap-1 flex-wrap">
                                {user.roles.map((role) => (
                                    <Badge key={role} variant={role === 'admin' ? 'default' : 'secondary'} className="capitalize">
                                        {role}
                                    </Badge>
                                ))}
                            </div>
                        </TableCell>
                        <TableCell>
                            {getAuthIcon(user)}
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
                                    <DropdownMenuItem onClick={() => copyId(user.id)}>
                                        <Copy className="mr-2 h-4 w-4" />
                                        Copy ID
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem onClick={() => onEdit(user)}>
                                        <Pencil className="mr-2 h-4 w-4" />
                                        Edit
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => onDelete(user.id)} className="text-destructive focus:text-destructive">
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
  );
}
