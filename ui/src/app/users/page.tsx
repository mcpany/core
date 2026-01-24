/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { apiClient } from "@/lib/client";
import { Badge } from "@/components/ui/badge";
import { Plus, Pencil, Trash2, Key } from "lucide-react";
import { ApiKeyDialog } from "@/components/api-key-dialog";

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

// Validation schema
const userSchema = z.object({
  id: z.string().min(3, "Username must be at least 3 characters").max(50, "Username too long").regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
  role: z.string().min(1, "Role is required"),
  password: z.string().optional(),
}).refine(() => {
  return true;
});

type UserValues = z.infer<typeof userSchema>;

/**
 * UsersPage component.
 * @returns The rendered component.
 */
export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  // API Key Dialog State
  const [isKeyDialogOpen, setIsKeyDialogOpen] = useState(false);
  const [generatedKey, setGeneratedKey] = useState("");
  const [keyDialogUser, setKeyDialogUser] = useState("");

  const form = useForm<UserValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      id: "",
      role: "",
      password: "",
    },
  });

  async function loadUsers() {
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
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadUsers();
  }, []);

  // Reset form when dialog opens/closes or editing user changes
  useEffect(() => {
    if (isDialogOpen) {
      if (editingUser) {
        form.reset({
          id: editingUser.id,
          role: editingUser.roles[0] || "viewer",
          password: "",
        });
      } else {
        form.reset({
          id: "",
          role: "viewer",
          password: "",
        });
      }
    }
  }, [isDialogOpen, editingUser, form]);

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setIsDialogOpen(true);
  };

  const handleCreate = () => {
    setEditingUser(null);
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

  const handleGenerateKey = async (user: User) => {
    if (!confirm(`Generate a new API Key for ${user.id}? This will invalidate any existing key.`)) return;

    // Generate a secure random key
    const array = new Uint8Array(24);
    crypto.getRandomValues(array);
    const key = "mcp_sk_" + Array.from(array).map(b => b.toString(16).padStart(2, '0')).join('');

    try {
        // Update user with new key
        const userPayload = {
            ...user,
            authentication: {
                ...user.authentication,
                api_key: {
                    key_value: key
                }
            }
        };

        await apiClient.updateUser(userPayload);

        setGeneratedKey(key);
        setKeyDialogUser(user.id);
        setIsKeyDialogOpen(true);
        loadUsers();
    } catch (e) {
        console.error("Failed to generate key", e);
        alert("Failed to generate API key. Check console for details.");
    }
  };

  const onSubmit = async (data: UserValues) => {
    // Custom validation for new user password
    if (!editingUser && !data.password) {
        form.setError("password", { message: "Password is required for new users" });
        return;
    }
    // Enforce strong password if provided
    if (data.password && data.password.length < 8) {
        form.setError("password", { message: "Password must be at least 8 characters" });
        return;
    }

    try {
        const userPayload = {
            id: data.id,
            roles: data.role ? [data.role] : [],
             authentication: data.password ? {
                basic_auth: {
                    password_hash: data.password // sent as plain text, server hashes it
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
        // Could set a form error here if API returns a message
    }
  };

  const getRoleBadgeVariant = (role: string) => {
      switch (role.toLowerCase()) {
          case 'admin': return 'destructive'; // Red
          case 'editor': return 'default'; // Primary/Black
          case 'viewer': return 'secondary'; // Gray
          default: return 'outline';
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Users</h2>
            <p className="text-muted-foreground">Manage system access, roles, and API keys.</p>
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
                                    <Badge key={role} variant={getRoleBadgeVariant(role)} className="mr-1 capitalize">
                                        {role}
                                    </Badge>
                                ))}
                            </TableCell>
                            <TableCell>
                                {user.authentication?.basic_auth ? "Basic Auth" :
                                 user.authentication?.api_key ? "API Key" : "None"}
                            </TableCell>
                            <TableCell className="text-right">
                                <Button variant="ghost" size="icon" onClick={() => handleGenerateKey(user)} title="Generate API Key">
                                    <Key className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" onClick={() => handleEdit(user)} title="Edit User">
                                    <Pencil className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" className="text-destructive" onClick={() => handleDelete(user.id)} title="Delete User">
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
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="id"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Username</FormLabel>
                      <FormControl>
                        <Input disabled={!!editingUser} placeholder="username" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="role"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Role</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select a role" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="admin">Admin</SelectItem>
                          <SelectItem value="editor">Editor</SelectItem>
                          <SelectItem value="viewer">Viewer</SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Password</FormLabel>
                      <FormControl>
                        <Input type="password" placeholder={editingUser ? "(Unchanged)" : "Required for new user"} {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <DialogFooter>
                    <Button type="submit">Save changes</Button>
                </DialogFooter>
              </form>
            </Form>
        </DialogContent>
      </Dialog>

      <ApiKeyDialog
        open={isKeyDialogOpen}
        onOpenChange={setIsKeyDialogOpen}
        apiKey={generatedKey}
        username={keyDialogUser}
      />
    </div>
  );
}
