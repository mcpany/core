/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { User } from "@proto/config/v1/user";
import { UserList } from "@/components/users/user-list";
import { UserSheet } from "@/components/users/user-sheet";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * UsersPage component.
 * @returns The rendered component.
 */
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

    const handleCreate = () => {
        setEditingUser(null);
        setIsSheetOpen(true);
    };

    const handleEdit = (user: User) => {
        setEditingUser(user);
        setIsSheetOpen(true);
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this user? This action cannot be undone.")) return;
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

    const handleSave = async (userUpdate: Partial<User>, password?: string, apiKey?: string) => {
        try {
            let authConfig: any = editingUser?.authentication || {};

            if (apiKey) {
                // Configure for API Key
                authConfig = {
                    api_key: {
                        param_name: "X-API-Key",
                        in: 0, // HEADER
                        verification_value: apiKey
                    }
                };
            } else if (password) {
                // Configure for Password
                authConfig = {
                    basic_auth: {
                        username: userUpdate.id, // Username usually matches ID
                        password_hash: password
                    }
                };
            }

            // Construct the payload
            // Server expects snake_case for fields.
            const payload: any = {
                id: userUpdate.id,
                roles: userUpdate.roles,
                authentication: authConfig,
                profile_ids: editingUser?.profileIds || []
            };



            if (editingUser) {
                await apiClient.updateUser(payload);
                toast({
                    title: "User Updated",
                    description: `User ${payload.id} updated successfully.`
                });
            } else {
                await apiClient.createUser(payload);
                 toast({
                    title: "User Created",
                    description: `User ${payload.id} created successfully.`
                });
            }
            loadUsers();
            setIsSheetOpen(false);
        } catch (e) {
            console.error("Failed to save user (handleSave):", e);
            if (e instanceof Error) {
                console.error("Error message:", e.message);
                console.error("Error stack:", e.stack);
            }
            throw e; // Re-throw to be caught by the Sheet's error handling if needed, or handle here.
            // Actually Sheet handles it via `onSave` promise rejection?
            // Yes, Sheet calls `await onSave(...)`.
        }
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
                isLoading={loading}
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
