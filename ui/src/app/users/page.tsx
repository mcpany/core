/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { Plus } from "lucide-react";
import { UserList } from "@/components/users/user-list";
import { UserSheet, User } from "@/components/users/user-sheet";
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

  async function loadUsers() {
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
          variant: "destructive"
      });
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadUsers();
  }, []);

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setIsSheetOpen(true);
  };

  const handleCreate = () => {
    setEditingUser(null);
    setIsSheetOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm(`Are you sure you want to delete user "${id}"?`)) return;
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
            variant: "destructive"
        });
    }
  };

  const handleSave = () => {
      loadUsers();
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Users</h2>
            <p className="text-muted-foreground">Manage system access, roles, and authentication.</p>
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
