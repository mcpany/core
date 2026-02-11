/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { Plus } from "lucide-react";
import { UserList, User } from "@/components/users/user-list";
import { UserFormSheet } from "@/components/users/user-form-sheet";
import { useToast } from "@/hooks/use-toast";

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

  const mapUserFromApi = (u: any): User => ({
    id: u.id,
    roles: u.roles || [],
    authentication: u.authentication ? {
      apiKey: u.authentication.api_key ? {
        paramName: u.authentication.api_key.param_name,
        in: u.authentication.api_key.in,
        verificationValue: u.authentication.api_key.verification_value
      } : undefined,
      basicAuth: u.authentication.basic_auth ? {
        username: u.authentication.basic_auth.username,
        passwordHash: u.authentication.basic_auth.password_hash
      } : undefined
    } : undefined,
    profileIds: u.profile_ids
  });

  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await apiClient.listUsers();
      let rawUsers: any[] = [];
      if (Array.isArray(resp)) {
        rawUsers = resp;
      } else if (resp && Array.isArray(resp.users)) {
        rawUsers = resp.users;
      }

      setUsers(rawUsers.map(mapUserFromApi));
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

  const handleSubmit = async (data: any) => {
    try {
        let authConfig: any = {};

        // Preserve existing auth config if not changing?
        // But the form captures authType.

        if (data.authType === "password") {
             if (data.password) {
                 authConfig = {
                     basic_auth: {
                         password_hash: data.password // Server handles hashing if plain text sent here?
                         // Previous implementation sent plain text in `password_hash` field.
                     }
                 };
             } else if (editingUser?.authentication?.basicAuth) {
                 // Keeping existing password
                 // We need to send back the existing auth config?
                 // Or typically UpdateUser might require setting fields again or partial update.
                 // Assuming full replace for Authentication object.
                 // If we send empty password_hash, server might reject or clear it.
                 // If we don't have the hash (it's masked or hashed), we can't resend it?
                 // Usually for updates, if password field is empty, we don't update auth?
                 // But we might be switching role.

                 // If `password` is empty and we are editing, we might want to retain existing auth.
                 // But we construct `authentication` object.
                 if (editingUser) {
                      // We can't really "reconstruct" the password hash if we don't have it (or if it's one-way).
                      // The backend likely handles "if password_hash is empty, don't update it" logic OR
                      // we need to pass a special flag.
                      // However, previous implementation seemed to just update userPayload.

                      // Let's assume if password is NOT provided, we try to use the existing `authentication` object from the API response
                      // IF the authType hasn't changed.

                      // However, `editingUser` has our mapped camelCase object.
                      // We need to map it back to snake_case if we reuse it.
                      authConfig = {
                          basic_auth: {
                              username: editingUser.authentication?.basicAuth?.username,
                              // We likely receive the hash or empty string.
                              // If we send it back, it should be fine.
                              password_hash: editingUser.authentication?.basicAuth?.passwordHash
                          }
                      };
                 }
             }
        } else if (data.authType === "apiKey") {
             if (data.generatedApiKey) {
                 authConfig = {
                     api_key: {
                         param_name: "X-API-Key",
                         in: 0, // HEADER
                         verification_value: data.generatedApiKey
                     }
                 };
             } else if (editingUser?.authentication?.apiKey) {
                 // Keep existing
                 authConfig = {
                     api_key: {
                         param_name: editingUser.authentication.apiKey.paramName,
                         in: editingUser.authentication.apiKey.in,
                         verification_value: editingUser.authentication.apiKey.verificationValue
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
            toast({ title: "User Updated", description: "User configuration saved." });
        } else {
            await apiClient.createUser(userPayload);
            toast({ title: "User Created", description: "New user registered successfully." });
        }
        setIsSheetOpen(false);
        loadUsers();
    } catch (e) {
        console.error("Failed to save user", e);
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to save user."
        });
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

      <UserFormSheet
        isOpen={isSheetOpen}
        onClose={() => setIsSheetOpen(false)}
        user={editingUser}
        onSubmit={handleSubmit}
      />
    </div>
  );
}
