/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from "@/components/ui/sheet";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { User } from "@/types/user";
import { Loader2, RotateCw, Copy, Check, Eye, EyeOff } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

const userSchema = z.object({
  id: z
    .string()
    .min(3, "Username must be at least 3 characters")
    .max(50, "Username too long")
    .regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
  role: z.string().min(1, "Role is required"),
  password: z.string().optional(),
  authType: z.enum(["password", "api_key"]),
});

type UserValues = z.infer<typeof userSchema>;

interface UserSheetProps {
  user: User | null;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (user: User) => Promise<void>;
}

/**
 * UserSheet component.
 * A slide-over form for creating and editing users.
 * @param props - The component props.
 * @param props.user - The user to edit, or null for creating a new user.
 * @param props.isOpen - Whether the sheet is open.
 * @param props.onOpenChange - Callback for changing the open state.
 * @param props.onSave - Callback for saving the user.
 * @returns The rendered component.
 */
export function UserSheet({ user, isOpen, onOpenChange, onSave }: UserSheetProps) {
  const { toast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [generatedKey, setGeneratedKey] = useState("");
  const [copied, setCopied] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const form = useForm<UserValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      id: "",
      role: "viewer",
      password: "",
      authType: "password",
    },
  });

  // Watch authType to conditionally validate
  const authType = form.watch("authType");

  useEffect(() => {
    if (isOpen) {
      setGeneratedKey("");
      setCopied(false);
      setShowPassword(false);

      if (user) {
        form.reset({
          id: user.id,
          role: user.roles[0] || "viewer",
          password: "",
          authType: user.authentication?.api_key ? "api_key" : "password",
        });
      } else {
        form.reset({
          id: "",
          role: "viewer",
          password: "",
          authType: "password",
        });
      }
    }
  }, [isOpen, user, form]);

  const generateApiKey = () => {
    const array = new Uint8Array(24);
    window.crypto.getRandomValues(array);
    const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, "0")).join("");
    setGeneratedKey(key);
    form.setValue("password", ""); // Clear password if generating key
    toast({
        title: "API Key Generated",
        description: "Please copy the key immediately.",
    });
  };

  const copyKey = () => {
    navigator.clipboard.writeText(generatedKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
    toast({
        title: "Copied",
        description: "API Key copied to clipboard.",
    });
  };

  const onSubmit = async (data: UserValues) => {
    // Custom validation logic
    if (data.authType === "password") {
      if (!user && !data.password) {
        form.setError("password", { message: "Password is required for new users" });
        return;
      }
      if (data.password && data.password.length < 8) {
        form.setError("password", { message: "Password must be at least 8 characters" });
        return;
      }
    } else {
      if (!user && !generatedKey) {
          toast({
              variant: "destructive",
              title: "Error",
              description: "Please generate an API Key first.",
          });
          return;
      }
    }

    setIsSubmitting(true);
    try {
      // Merge with existing user data to preserve fields like profile_ids
      const newUser: User = {
        ...(user || {}),
        id: data.id,
        roles: [data.role],
        authentication: user?.authentication || {},
      };

      if (data.authType === "password") {
         if (data.password) {
             newUser.authentication = {
                 basic_auth: {
                     password_hash: data.password // Server will hash it
                 }
             };
         } else if (user && user.authentication?.basic_auth) {
             // Keep existing password. Explicitly set to only basic_auth to clear potential API key if switching back without changing password (edge case)
             newUser.authentication = {
                 basic_auth: user.authentication.basic_auth
             };
         } else {
             // User selected Password tab, but no password provided, and no existing password.
             form.setError("password", { message: "Password is required" });
             setIsSubmitting(false);
             return;
         }
      } else {
          // API Key
          if (generatedKey) {
              newUser.authentication = {
                  api_key: {
                      param_name: "X-API-Key",
                      in: 0, // HEADER
                      verification_value: generatedKey
                  }
              };
          } else if (user && user.authentication?.api_key) {
              // Keeping existing key
              newUser.authentication = {
                  api_key: user.authentication.api_key
              };
          } else {
               // Switching to API Key but didn't generate one? Error.
               toast({
                  variant: "destructive",
                  title: "Error",
                  description: "Please generate a new API Key when switching auth methods.",
              });
              setIsSubmitting(false);
              return;
          }
      }

      await onSave(newUser);
      onOpenChange(false);
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to save user.",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Sheet open={isOpen} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md overflow-y-auto">
        <SheetHeader>
          <SheetTitle>{user ? "Edit User" : "Add New User"}</SheetTitle>
          <SheetDescription>
            {user
              ? "Update user details and permissions."
              : "Create a new user account with specific roles."}
          </SheetDescription>
        </SheetHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-6">
            <FormField
              control={form.control}
              name="id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Username</FormLabel>
                  <FormControl>
                    <Input placeholder="jdoe" {...field} disabled={!!user} />
                  </FormControl>
                  <FormDescription>
                    Unique identifier for the user.
                  </FormDescription>
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

            <div className="space-y-4 border-t pt-4">
                <FormLabel className="text-base">Authentication</FormLabel>
                <Tabs
                    defaultValue={authType}
                    onValueChange={(v) => form.setValue("authType", v as "password" | "api_key")}
                    className="w-full"
                >
                    <TabsList className="grid w-full grid-cols-2">
                        <TabsTrigger value="password">Password</TabsTrigger>
                        <TabsTrigger value="api_key">API Key</TabsTrigger>
                    </TabsList>

                    <TabsContent value="password" className="space-y-4 pt-4">
                        <FormField
                            control={form.control}
                            name="password"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Password</FormLabel>
                                    <FormControl>
                                        <div className="relative">
                                            <Input
                                                type={showPassword ? "text" : "password"}
                                                placeholder={user ? "Leave empty to keep unchanged" : "Enter password"}
                                                {...field}
                                            />
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="sm"
                                                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                                                onClick={() => setShowPassword(!showPassword)}
                                            >
                                                {showPassword ? (
                                                    <EyeOff className="h-4 w-4 text-muted-foreground" />
                                                ) : (
                                                    <Eye className="h-4 w-4 text-muted-foreground" />
                                                )}
                                            </Button>
                                        </div>
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </TabsContent>

                    <TabsContent value="api_key" className="space-y-4 pt-4">
                        <div className="rounded-md bg-muted p-4 space-y-3">
                            <div className="flex items-center justify-between">
                                <span className="text-sm font-medium">API Key</span>
                                <Button type="button" size="sm" variant="secondary" onClick={generateApiKey}>
                                    <RotateCw className="mr-2 h-3 w-3" />
                                    Generate New
                                </Button>
                            </div>

                            {generatedKey ? (
                                <div className="space-y-2 animate-in fade-in slide-in-from-top-2">
                                    <div className="relative">
                                        <div className="font-mono text-xs bg-background p-2 rounded border break-all pr-10">
                                            {generatedKey}
                                        </div>
                                        <Button
                                            type="button"
                                            size="icon"
                                            variant="ghost"
                                            className="absolute right-1 top-1 h-6 w-6"
                                            onClick={copyKey}
                                        >
                                            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
                                        </Button>
                                    </div>
                                    <p className="text-[10px] text-destructive font-medium">
                                        Make sure to copy this key now. You won't be able to see it again!
                                    </p>
                                </div>
                            ) : (
                                <p className="text-xs text-muted-foreground italic">
                                    {user?.authentication?.api_key ? "An API key is currently configured." : "No API key generated."}
                                </p>
                            )}
                        </div>
                    </TabsContent>
                </Tabs>
            </div>

            <SheetFooter>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {user ? "Save Changes" : "Create User"}
              </Button>
            </SheetFooter>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  );
}
