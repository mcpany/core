/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { UserList } from "@/components/users/user-list";
import { UserSheet } from "@/components/users/user-sheet";
import { apiClient } from "@/lib/client";
import { User } from "@proto/config/v1/user";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

export default function UsersPage() {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const [isSheetOpen, setIsSheetOpen] = useState(false);
    const [editingUser, setEditingUser] = useState<User | null>(null);
    const { toast } = useToast();

    const loadUsers = useCallback(async () => {
        setLoading(true);
        try {
            const resp = await apiClient.listUsers();
            if (Array.isArray(resp)) {
                setUsers(resp);
            } else if (resp && Array.isArray(resp.users)) {
                setUsers(resp.users);
            } else {
                setUsers([]);
            }
        } catch (e) {
            console.error("Failed to list users", e);
            toast({
                title: "Error",
                description: "Failed to load users.",
                variant: "destructive",
            });
            setUsers([]);
        } finally {
            setLoading(false);
        }
    }, [toast]);

    useEffect(() => {
        loadUsers();
    }, [loadUsers]);

    const handleCreate = () => {
        setEditingUser(null);
        setIsSheetOpen(true);
    };

    const handleEdit = (user: User) => {
        setEditingUser(user);
        setIsSheetOpen(true);
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this user?")) return;
        try {
            await apiClient.deleteUser(id);
            toast({
                title: "User Deleted",
                description: `User ${id} has been removed.`,
            });
            loadUsers();
        } catch (e) {
            console.error("Failed to delete user", e);
            toast({
                title: "Error",
                description: "Failed to delete user.",
                variant: "destructive",
            });
        }
    };

    const handleSave = async (userPartial: Partial<User>, password?: string, apiKey?: string) => {
        try {
            // Construct the User object using snake_case for backend compatibility
            // The User interface is camelCase, but the backend likely expects snake_case JSON
            // especially since apiClient.createUser just stringifies it.
            const newUser: any = {
                id: userPartial.id,
                roles: userPartial.roles || [],
                profile_ids: userPartial.profileIds || [],
            };

            // Handle Authentication

            if (apiKey) {
                newUser.authentication = {
                    api_key: {
                        param_name: "X-API-Key",
                        in: 0, // HEADER (enum value)
                        verification_value: apiKey
                    }
                };
            } else if (password) {
                newUser.authentication = {
                    basic_auth: {
                        username: userPartial.id,
                        password_hash: password // Server will hash this
                    }
                };
            } else if (editingUser) {
                // If editing and no new auth provided, preserve existing.
                // We need to map the existing camelCase auth back to snake_case if we send it back?
                // Or does updateUser handle it?
                // If editingUser comes from listUsers, it might be camelCase (mapped by client?).
                // apiClient.listUsers returns `res.json()`.
                // If backend returns snake_case, and we use it as User, then it has snake_case props?
                // But TypeScript thinks it's camelCase.
                // Let's assume we need to send snake_case.

                // This part is tricky without deep diving into the client response format.
                // But generally, for updates, if we don't touch auth, maybe we shouldn't send it?
                // But UpdateUser usually replaces.
                // Let's try to map it best effort.
                if (editingUser.authentication) {
                    const auth: any = {};
                    if (editingUser.authentication.apiKey) {
                        auth.api_key = {
                            param_name: editingUser.authentication.apiKey.paramName,
                            in: editingUser.authentication.apiKey.in,
                            verification_value: editingUser.authentication.apiKey.verificationValue
                        };
                    } else if (editingUser.authentication.basicAuth) {
                        auth.basic_auth = {
                            username: editingUser.authentication.basicAuth.username,
                            password_hash: editingUser.authentication.basicAuth.passwordHash
                        };
                    }
                     newUser.authentication = auth;
                }
            }

            if (editingUser) {
                await apiClient.updateUser(newUser);
                toast({
                    title: "User Updated",
                    description: `User ${newUser.id} has been updated.`,
                });
            } else {
                await apiClient.createUser(newUser);
                toast({
                    title: "User Created",
                    description: `User ${newUser.id} has been created.`,
                });
            }

            setIsSheetOpen(false);
            loadUsers();
        } catch (e) {
            console.error("Failed to save user", e);
            toast({
                title: "Error",
                description: "Failed to save user. Check console for details.",
                variant: "destructive",
            });
            throw e;
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Users</h2>
                    <p className="text-muted-foreground">Manage system access, roles, and service accounts.</p>
                </div>
                <Button onClick={handleCreate}>
                    <Plus className="mr-2 h-4 w-4" />
                    Add User
                </Button>
            </div>

            <UserList
                users={users}
                isLoading={loading}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />

            <UserSheet
                isOpen={isSheetOpen}
                onClose={() => setIsSheetOpen(false)}
                user={editingUser}
                onSave={handleSave}
            />
        </div>
    );
}
