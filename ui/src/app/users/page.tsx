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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { apiClient } from "@/lib/client";
import { Badge } from "@/components/ui/badge";
import { Plus, Pencil, Trash2, Key, RotateCw, Copy, Check, UserCheck } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ProfileDefinition } from "@proto/config/v1/config";

interface User {
  id: string;
  roles: string[];
  profile_ids?: string[];
  authentication?: {
    basic_auth?: Record<string, never>;
    api_key?: {
        param_name?: string;
        verification_value?: string;
    };
  };
}

// Validation schema
const userSchema = z.object({
  id: z.string().min(3, "Username must be at least 3 characters").max(50, "Username too long").regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
  role: z.string().min(1, "Role is required"),
  profile_id: z.string().optional(),
  password: z.string().optional(),
});

type UserValues = z.infer<typeof userSchema>;

/**
 * UsersPage component.
 * @returns The rendered component.
 */
export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [profiles, setProfiles] = useState<ProfileDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [authType, setAuthType] = useState<"password" | "api_key">("password");
  const [generatedKey, setGeneratedKey] = useState("");
  const [copied, setCopied] = useState(false);

  const form = useForm<UserValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      id: "",
      role: "",
      profile_id: "none",
      password: "",
    },
  });

  async function loadData() {
    setLoading(true);
    try {
      const [usersResp, profilesResp] = await Promise.all([
          apiClient.listUsers(),
          apiClient.listProfiles()
      ]);

      if (Array.isArray(usersResp)) {
        setUsers(usersResp);
      } else if (usersResp && Array.isArray(usersResp.users)) {
        setUsers(usersResp.users);
      } else {
        setUsers([]);
      }

      setProfiles(profilesResp || []);
    } catch (e) {
      console.error("Failed to load data", e);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadData();
  }, []);

  // Reset form when dialog opens/closes or editing user changes
  useEffect(() => {
    if (isDialogOpen) {
      setGeneratedKey("");
      setCopied(false);
      if (editingUser) {
        form.reset({
          id: editingUser.id,
          role: editingUser.roles[0] || "viewer",
          profile_id: (editingUser.profile_ids && editingUser.profile_ids.length > 0) ? editingUser.profile_ids[0] : "none",
          password: "",
        });
        if (editingUser.authentication?.api_key) {
            setAuthType("api_key");
        } else {
            setAuthType("password");
        }
      } else {
        form.reset({
          id: "",
          role: "viewer",
          profile_id: "none",
          password: "",
        });
        setAuthType("password");
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
        loadData();
    } catch (e) {
        console.error("Failed to delete user", e);
    }
  };

  const generateApiKey = () => {
    // secure random string
    const array = new Uint8Array(24);
    window.crypto.getRandomValues(array);
    const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
    setGeneratedKey(key);
    // Clear password error if any
    form.clearErrors("password");
  };

  const copyKey = () => {
    navigator.clipboard.writeText(generatedKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const onSubmit = async (data: UserValues) => {
    // Custom validation
    if (authType === "password") {
        if (!editingUser && !data.password) {
            form.setError("password", { message: "Password is required for new users" });
            return;
        }
        if (data.password && data.password.length < 8) {
            form.setError("password", { message: "Password must be at least 8 characters" });
            return;
        }
    } else {
        if (!editingUser && !generatedKey) {
             const alreadyApiKey = editingUser?.authentication?.api_key;
             if (!alreadyApiKey) {
                 // Must generate
                 alert("Please generate an API Key first.");
                 return;
             }
        }
    }

    try {
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

        const userPayload: any = {
            id: data.id,
            roles: data.role ? [data.role] : [],
            authentication: authConfig,
            profile_ids: (data.profile_id && data.profile_id !== "none") ? [data.profile_id] : []
        };

        if (editingUser) {
            await apiClient.updateUser(userPayload);
        } else {
            await apiClient.createUser(userPayload);
        }
        setIsDialogOpen(false);
        loadData();
    } catch (e) {
        console.error("Failed to save user", e);
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Users</h2>
            <p className="text-muted-foreground">Manage system access, roles, and profiles.</p>
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
                        <TableHead>Profile</TableHead>
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
                                {user.profile_ids && user.profile_ids.length > 0 ? (
                                    <Badge variant="secondary" className="gap-1">
                                        <UserCheck className="h-3 w-3" /> {user.profile_ids[0]}
                                    </Badge>
                                ) : (
                                    <span className="text-muted-foreground text-sm italic">None</span>
                                )}
                            </TableCell>
                            <TableCell>
                                {user.authentication?.basic_auth ? <div className="flex items-center gap-1"><Key className="h-3 w-3"/> Password</div> :
                                 user.authentication?.api_key ? <div className="flex items-center gap-1"><Key className="h-3 w-3 text-primary"/> API Key</div> : "None"}
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
                            <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
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
                      <FormLabel>Role (Tag)</FormLabel>
                      <FormControl>
                        <Input placeholder="admin, viewer, etc." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="profile_id"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Assigned Profile</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select a profile" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="none">None (Full Access)</SelectItem>
                          {profiles.map((profile) => (
                            <SelectItem key={profile.name} value={profile.name}>
                              {profile.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="space-y-2 pt-2 border-t">
                    <FormLabel>Authentication Method</FormLabel>
                    <Tabs value={authType} onValueChange={(v) => setAuthType(v as any)} className="w-full">
                        <TabsList className="grid w-full grid-cols-2">
                            <TabsTrigger value="password">Password</TabsTrigger>
                            <TabsTrigger value="api_key">API Key</TabsTrigger>
                        </TabsList>
                        <TabsContent value="password" className="pt-2">
                            <FormField
                            control={form.control}
                            name="password"
                            render={({ field }) => (
                                <FormItem>
                                <FormControl>
                                    <Input type="password" placeholder={editingUser ? "(Unchanged)" : "Required for new user"} {...field} />
                                </FormControl>
                                <FormMessage />
                                </FormItem>
                            )}
                            />
                        </TabsContent>
                        <TabsContent value="api_key" className="pt-2 space-y-4">
                            <div className="text-sm text-muted-foreground">
                                Generate a unique API Key for this user (Agent).
                            </div>
                            <div className="flex gap-2">
                                <Input value={generatedKey} readOnly placeholder={editingUser?.authentication?.api_key ? "(Existing Key Configured)" : "Generate a key..."} className="font-mono text-xs" />
                                <Button type="button" onClick={generateApiKey} variant="secondary"><RotateCw className="mr-2 h-4 w-4"/> Generate</Button>
                            </div>
                            {generatedKey && (
                                <div className="space-y-2">
                                    <Button type="button" variant="outline" className="w-full" onClick={copyKey}>
                                        {copied ? <Check className="mr-2 h-4 w-4 text-green-500" /> : <Copy className="mr-2 h-4 w-4" />}
                                        {copied ? "Copied to Clipboard" : "Copy Key"}
                                    </Button>
                                    <p className="text-xs text-red-500 dark:text-red-400">
                                        <strong>Important:</strong> This key will be shown only once. Please copy it now.
                                    </p>
                                </div>
                            )}
                        </TabsContent>
                    </Tabs>
                </div>

                <DialogFooter>
                    <Button type="submit">Save changes</Button>
                </DialogFooter>
              </form>
            </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
