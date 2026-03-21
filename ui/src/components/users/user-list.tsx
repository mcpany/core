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
 * @returns The rendered UserList component.
 */
export function UserList({ users, isLoading, onEdit, onDelete }: UserListProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const { toast } = useToast();
    const [selectedUsers, setSelectedUsers] = useState<Set<string>>(new Set());
    const [isDeletingBulk, setIsDeletingBulk] = useState(false);

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
            setSelectedUsers(new Set(filteredUsers.map(u => u.id)));
        } else {
            setSelectedUsers(new Set());
        }
    };

    const handleSelectUser = (id: string, checked: boolean) => {
        const next = new Set(selectedUsers);
        if (checked) {
            next.add(id);
        } else {
            next.delete(id);
        }
        setSelectedUsers(next);
    };

    const handleBulkDelete = async () => {
        if (selectedUsers.size === 0) return;
        setIsDeletingBulk(true);
        try {
            // Need API changes or multiple calls. The prompt mentions a floating bar for bulk delete.
            // We'll call onDelete for each as a quick workaround if bulk endpoint doesn't exist.
            for (const id of Array.from(selectedUsers)) {
                await onDelete(id);
            }
            setSelectedUsers(new Set());
            toast({
                title: "Success",
                description: `Successfully deleted ${selectedUsers.size} user(s).`,
            });
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to delete some users.",
                variant: "destructive",
            });
        } finally {
            setIsDeletingBulk(false);
        }
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

            <div className="rounded-md border bg-background relative overflow-hidden">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[50px] pr-0">
                                <Checkbox
                                    checked={selectedUsers.size > 0 && selectedUsers.size === filteredUsers.length}
                                    onCheckedChange={handleSelectAll}
                                    aria-label="Select all users"
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
                                <TableRow key={user.id} data-testid={`user-row-${user.id}`} data-state={selectedUsers.has(user.id) ? "selected" : undefined}>
                                    <TableCell className="pr-0">
                                        <Checkbox
                                            checked={selectedUsers.has(user.id)}
                                            onCheckedChange={(checked) => handleSelectUser(user.id, !!checked)}
                                            aria-label={`Select user ${user.id}`}
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
                                            {user.authentication?.apiKey || (user.authentication as Record<string, unknown>)?.api_key ? (
                                                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/50 border">
                                                    <Key className="h-3.5 w-3.5 text-orange-500" />
                                                    <span>API Key</span>
                                                </div>
                                            ) : user.authentication?.basicAuth || (user.authentication as Record<string, unknown>)?.basic_auth ? (
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

            {selectedUsers.size > 0 && (
                <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50">
                    <div className="flex items-center gap-4 px-6 py-4 rounded-full border bg-background/80 backdrop-blur-md shadow-lg animate-in slide-in-from-bottom-5">
                        <div className="flex items-center gap-2 border-r pr-4">
                            <Badge variant="secondary" className="h-6 w-6 rounded-full p-0 flex items-center justify-center">
                                {selectedUsers.size}
                            </Badge>
                            <span className="text-sm font-medium">users selected</span>
                        </div>
                        <Button
                            variant="destructive"
                            size="sm"
                            onClick={handleBulkDelete}
                            disabled={isDeletingBulk}
                            className="h-8 shadow-none"
                        >
                            {isDeletingBulk ? (
                                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2" />
                            ) : (
                                <Trash2 className="h-4 w-4 mr-2" />
                            )}
                            Delete Selected
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}
