/**
 * Copyright 2025 Author(s) of MCP Any
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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import {
  MoreHorizontal,
  Search,
  User as UserIcon,
  Key,
  Lock,
  Shield,
  Trash2,
  Edit,
  Copy
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * Interface matching the backend User proto.
 */
export interface User {
  id: string;
  roles: string[];
  authentication?: {
    apiKey?: {
        paramName?: string;
        in?: number;
        verificationValue?: string;
    };
    basicAuth?: {
        username: string;
        passwordHash?: string;
    };
    // Add other auth methods as needed
  };
  profileIds?: string[];
}

interface UserListProps {
  users: User[];
  isLoading: boolean;
  onEdit: (user: User) => void;
  onDelete: (userId: string) => void;
}

/**
 * UserList component.
 * Displays a searchable table of users with actions.
 * @param props The component props.
 * @returns The rendered component.
 */
export function UserList({ users, isLoading, onEdit, onDelete }: UserListProps) {
  const [filter, setFilter] = useState("");
  const { toast } = useToast();

  const filteredUsers = useMemo(() => {
    if (!filter) return users;
    const lowerFilter = filter.toLowerCase();
    return users.filter(
      (user) =>
        user.id.toLowerCase().includes(lowerFilter) ||
        user.roles.some((role) => role.toLowerCase().includes(lowerFilter))
    );
  }, [users, filter]);

  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase();
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: "Copied",
      description: `${label} copied to clipboard.`,
    });
  };

  if (isLoading) {
      return (
          <div className="space-y-4">
               {[...Array(3)].map((_, i) => (
                  <div key={i} className="w-full h-16 bg-muted/50 animate-pulse rounded-md" />
               ))}
          </div>
      );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 max-w-sm">
        <div className="relative flex-1">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
            placeholder="Filter users..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="pl-8"
            />
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
                    <Avatar>
                      <AvatarImage src={`https://avatar.vercel.sh/${user.id}.png`} alt={user.id} />
                      <AvatarFallback>{getInitials(user.id)}</AvatarFallback>
                    </Avatar>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-col">
                      <span className="font-medium">{user.id}</span>
                      {user.authentication?.basicAuth?.username && user.authentication.basicAuth.username !== user.id && (
                          <span className="text-xs text-muted-foreground">aka {user.authentication.basicAuth.username}</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {user.roles.map((role) => (
                        <Badge key={role} variant="secondary" className="text-xs font-normal">
                          {role}
                        </Badge>
                      ))}
                      {user.roles.length === 0 && (
                          <span className="text-xs text-muted-foreground italic">No roles</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        {user.authentication?.apiKey ? (
                            <>
                                <Key className="h-4 w-4 text-primary" />
                                <span>API Key</span>
                            </>
                        ) : user.authentication?.basicAuth ? (
                            <>
                                <Lock className="h-4 w-4 text-amber-500" />
                                <span>Password</span>
                            </>
                        ) : (
                            <>
                                <Shield className="h-4 w-4" />
                                <span>None</span>
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
                        <DropdownMenuItem onClick={() => copyToClipboard(user.id, "User ID")}>
                          <Copy className="mr-2 h-4 w-4" />
                          Copy ID
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={() => onEdit(user)}>
                          <Edit className="mr-2 h-4 w-4" />
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
