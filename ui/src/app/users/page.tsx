"use client";

import { useEffect, useState } from "react";
import { User } from "@proto/config/v1/user";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { UserList } from "@/components/users/user-list";
import { UserSheet } from "@/components/users/user-sheet";
import { useToast } from "@/hooks/use-toast";

export default function UsersPage() {
  const { toast } = useToast();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  async function loadUsers() {
    setLoading(true);
    try {
      const resp = await apiClient.listUsers();
      // Handle various response formats and convert to User object
      let rawList: any[] = [];
      if (Array.isArray(resp)) {
        rawList = resp;
      } else if (resp && Array.isArray(resp.users)) {
        rawList = resp.users;
      }

      const userList = rawList.map((u: any) => User.fromJSON(u));
      setUsers(userList);
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
  }

  useEffect(() => {
    loadUsers();
  }, []);

  const handleCreate = () => {
    setEditingUser(null);
    setIsSheetOpen(true);
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setIsSheetOpen(true);
  };

  const handleDelete = async (id: string) => {
      if (!confirm(`Are you sure you want to delete user "${id}"?`)) return;
      try {
          await apiClient.deleteUser(id);
          toast({ title: "User Deleted", description: `User ${id} has been removed.` });
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
        isOpen={isSheetOpen}
        onOpenChange={setIsSheetOpen}
        user={editingUser}
        onSave={loadUsers}
      />
    </div>
  );
}
