/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { apiClient } from "@/lib/client";
import { UserList } from "@/components/users/user-list";
import { UserSheet } from "@/components/users/user-sheet";
import { User } from "@/types/user";
import { useToast } from "@/hooks/use-toast";

/**
 * UsersPage component.
 * @returns The rendered component.
 */
export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
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
        description: "Failed to load users.",
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
    setSelectedUser(user);
    setIsSheetOpen(true);
  };

  const handleCreate = () => {
    setSelectedUser(null);
    setIsSheetOpen(true);
  };

  const handleDelete = async (user: User) => {
    if (!confirm(`Are you sure you want to delete user "${user.id}"?`)) return;
    try {
      await apiClient.deleteUser(user.id);
      toast({
        title: "User Deleted",
        description: `User ${user.id} has been removed.`,
      });
      loadUsers();
    } catch (e) {
      console.error("Failed to delete user", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to delete user.",
      });
    }
  };

  const handleSave = async (user: User) => {
    try {
      if (selectedUser) {
        await apiClient.updateUser(user);
        toast({
          title: "User Updated",
          description: `User ${user.id} has been updated.`,
        });
      } else {
        await apiClient.createUser(user);
        toast({
          title: "User Created",
          description: `User ${user.id} has been created.`,
        });
      }
      loadUsers();
    } catch (e) {
      console.error("Failed to save user", e);
      throw e; // Re-throw to be caught by the sheet
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
        user={selectedUser}
        isOpen={isSheetOpen}
        onOpenChange={setIsSheetOpen}
        onSave={handleSave}
      />
    </div>
  );
}
