/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { UserList, User } from "@/components/users/user-list";
import { UserSheet } from "@/components/users/user-sheet";

/**
 * UsersPage component.
 * @returns The rendered component.
 */
export default function UsersPage() {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingUser, setEditingUser] = useState<User | null>(null);
    const [isSheetOpen, setIsSheetOpen] = useState(false);
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
                variant: "destructive",
                title: "Error",
                description: "Failed to load users."
            });
            setUsers([]);
        } finally {
            setLoading(false);
        }
    }, [toast]);

    useEffect(() => {
        loadUsers();
    }, [loadUsers]);

    const handleEdit = (user: User) => {
        setEditingUser(user);
        setIsSheetOpen(true);
    };

    const handleCreate = () => {
        setEditingUser(null);
        setIsSheetOpen(true);
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this user?")) return;
        try {
            await apiClient.deleteUser(id);
            toast({
                title: "User Deleted",
                description: `User ${id} has been removed.`
            });
            loadUsers();
        } catch (e) {
            console.error("Failed to delete user", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to delete user."
            });
        }
    };

    const handleSave = async (data: any, authType: "password" | "api_key", generatedKey?: string) => {
        let authConfig = editingUser?.authentication;

        if (authType === "password") {
            if (data.password) {
                authConfig = {
                    basic_auth: {
                        password_hash: data.password // sent as plain text, server hashes it
                    }
                };
            }
        } else {
            // API Key
            if (generatedKey) {
                authConfig = {
                    api_key: {
                        param_name: "X-API-Key",
                        in: 0, // HEADER
                        verification_value: generatedKey
                    }
                };
            }
        }

        const userPayload = {
            id: data.id,
            roles: data.role ? [data.role] : [],
            authentication: authConfig
        };

        if (editingUser) {
            await apiClient.updateUser(userPayload);
            toast({
                title: "User Updated",
                description: `User ${data.id} has been updated.`
            });
        } else {
            await apiClient.createUser(userPayload);
            toast({
                title: "User Created",
                description: `New user ${data.id} has been registered.`
            });
        }
        loadUsers();
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Users</h2>
                    <p className="text-muted-foreground">Manage system access and roles.</p>
                </div>
                <Button onClick={handleCreate}>
                    <Plus className="mr-2 h-4 w-4" />
                    Add User
                </Button>
            </div>

            <UserList
                users={users}
                loading={loading}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />

            <UserSheet
                open={isSheetOpen}
                onOpenChange={setIsSheetOpen}
                user={editingUser}
                onSave={handleSave}
            />
        </div>
    );
}
