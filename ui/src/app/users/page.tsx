/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/client";
import { Badge } from "@/components/ui/badge";
import { Plus, Pencil, Trash2 } from "lucide-react";

interface User {
  id: string;
  roles: string[];
  authentication?: {
    basic_auth?: Record<string, never>;
    api_key?: {
        key_value: string;
    };
  };
}

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  // Form state
  const [formData, setFormData] = useState({
    id: "",
    role: "",
    password: "", // Only for new/update
  });

  async function loadUsers() {
    try {
      const resp = await apiClient.listUsers();
      if (resp && Array.isArray(resp.users)) {
        setUsers(resp.users);
      } else {
        setUsers([]);
      }
    } catch (e) {
      console.error("Failed to list users", e);
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
    setFormData({
        id: user.id,
        role: user.roles[0] || "",
        password: "",
    });
    setIsDialogOpen(true);
  };

  const handleCreate = () => {
    setEditingUser(null);
    setFormData({
        id: "",
        role: "viewer",
        password: "",
    });
    setIsDialogOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this user?")) return;
    try {
        await apiClient.deleteUser(id);
        loadUsers();
    } catch (e) {
        console.error("Failed to delete user", e);
    }
  };

  const handleSubmit = async () => {
    try {
        const userPayload = {
            id: formData.id,
            roles: formData.role ? [formData.role] : [],
             authentication: formData.password ? {
                basic_auth: {
                    password_hash: formData.password // sent as plain text, server hashes it
                }
            } : editingUser?.authentication // keep existing auth if password not changed
        };

        if (editingUser) {
            await apiClient.updateUser(userPayload);
        } else {
            await apiClient.createUser(userPayload);
        }
        setIsDialogOpen(false);
        loadUsers();
    } catch (e) {
        console.error("Failed to save user", e);
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

      <Card>
        <CardContent className="p-0">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>User ID</TableHead>
                        <TableHead>Roles</TableHead>
                        <TableHead>Auth Method</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {users.map((user) => (
                        <TableRow key={user.id}>
                            <TableCell className="font-medium">{user.id}</TableCell>
                            <TableCell>
                                {user.roles.map(role => (
                                    <Badge key={role} variant="outline" className="mr-1">
                                        {role}
                                    </Badge>
                                ))}
                            </TableCell>
                            <TableCell>
                                {user.authentication?.basic_auth ? "Basic Auth" :
                                 user.authentication?.api_key ? "API Key" : "None"}
                            </TableCell>
                            <TableCell className="text-right">
                                <Button variant="ghost" size="icon" onClick={() => handleEdit(user)}>
                                    <Pencil className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" className="text-destructive" onClick={() => handleDelete(user.id)}>
                                    <Trash2 className="h-4 w-4" />
                                </Button>
                            </TableCell>
                        </TableRow>
                    ))}
                     {users.length === 0 && !loading && (
                        <TableRow>
                            <TableCell colSpan={4} className="text-center h-24 text-muted-foreground">
                                No users found.
                            </TableCell>
                        </TableRow>
                    )}
                </TableBody>
            </Table>
        </CardContent>
      </Card>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
            <DialogHeader>
                <DialogTitle>{editingUser ? "Edit User" : "Create User"}</DialogTitle>
                <DialogDescription>
                    {editingUser ? "Update user details." : "Add a new user to the system."}
                </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
                <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="username" className="text-right">
                        Username
                    </Label>
                    <Input
                        id="username"
                        value={formData.id}
                        onChange={(e) => setFormData({...formData, id: e.target.value})}
                        className="col-span-3"
                        disabled={!!editingUser}
                    />
                </div>
                <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="role" className="text-right">
                        Role
                    </Label>
                    <Input
                        id="role"
                        value={formData.role}
                        onChange={(e) => setFormData({...formData, role: e.target.value})}
                        className="col-span-3"
                        placeholder="admin, viewer, etc."
                    />
                </div>
                 <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="password" className="text-right">
                        Password
                    </Label>
                    <Input
                        id="password"
                        type="password"
                        value={formData.password}
                        onChange={(e) => setFormData({...formData, password: e.target.value})}
                        className="col-span-3"
                        placeholder={editingUser ? "(Unchanged)" : "Required for Basic Auth"}
                    />
                </div>
            </div>
            <DialogFooter>
                <Button onClick={handleSubmit}>Save changes</Button>
            </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
