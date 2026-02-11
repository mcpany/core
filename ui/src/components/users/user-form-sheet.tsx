/**
 * Copyright 2025 Author(s) of MCP Any
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
  FormDescription,
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
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { RotateCw, Check, Copy, Eye, EyeOff, AlertTriangle, Key, Lock } from "lucide-react";
import { User } from "./user-list";
import { useToast } from "@/hooks/use-toast";

// Zod Schema
const userFormSchema = z.object({
  id: z.string()
    .min(3, "Username must be at least 3 characters")
    .max(50, "Username too long")
    .regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
  role: z.string().min(1, "Role is required"),
  authType: z.enum(["password", "apiKey"]),
  password: z.string().optional(),
  // apiKey is generated, not input by user usually, but we need to track if it's set
  // We'll handle API key generation state separately
});

type UserFormValues = z.infer<typeof userFormSchema>;

interface UserFormSheetProps {
  isOpen: boolean;
  onClose: () => void;
  user?: User | null; // If null, creating new user
  onSubmit: (data: UserFormValues & { generatedApiKey?: string }) => Promise<void>;
}

/**
 * UserFormSheet component.
 * A sheet/drawer containing the form to create or edit a user.
 * @param props The component props.
 * @returns The rendered component.
 */
export function UserFormSheet({ isOpen, onClose, user, onSubmit }: UserFormSheetProps) {
  const { toast } = useToast();
  const [generatedKey, setGeneratedKey] = useState("");
  const [copied, setCopied] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm<UserFormValues>({
    resolver: zodResolver(userFormSchema),
    defaultValues: {
      id: "",
      role: "viewer",
      authType: "password",
      password: "",
    },
  });

  // Reset form when opening/closing or changing user
  useEffect(() => {
    if (isOpen) {
      setGeneratedKey("");
      setCopied(false);
      setShowPassword(false);

      if (user) {
        form.reset({
          id: user.id,
          role: user.roles[0] || "viewer",
          authType: user.authentication?.apiKey ? "apiKey" : "password",
          password: "", // Never show existing password
        });
      } else {
        form.reset({
          id: "",
          role: "viewer",
          authType: "password",
          password: "",
        });
      }
    }
  }, [isOpen, user, form]);

  const generateApiKey = () => {
    const array = new Uint8Array(24);
    window.crypto.getRandomValues(array);
    const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
    setGeneratedKey(key);
    // Clear password error if we switch to API Key
    form.clearErrors("password");
  };

  const copyKey = () => {
    navigator.clipboard.writeText(generatedKey);
    setCopied(true);
    toast({
      title: "Copied",
      description: "API Key copied to clipboard.",
    });
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSubmit = async (data: UserFormValues) => {
    setIsSubmitting(true);
    try {
      // Custom validation
      if (data.authType === "password") {
        if (!user && !data.password) {
          form.setError("password", { message: "Password is required for new users" });
          setIsSubmitting(false);
          return;
        }
        if (data.password && data.password.length < 8) {
          form.setError("password", { message: "Password must be at least 8 characters" });
          setIsSubmitting(false);
          return;
        }
      } else {
        // API Key
        if (!user && !generatedKey) {
            toast({
                variant: "destructive",
                title: "API Key Required",
                description: "Please generate an API Key for the new user.",
            });
            setIsSubmitting(false);
            return;
        }
      }

      await onSubmit({ ...data, generatedApiKey: generatedKey });
      onClose();
    } catch (error) {
      console.error(error);
      // Toast handled by parent usually, but fallback
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Sheet open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className="sm:max-w-md w-full overflow-y-auto">
        <SheetHeader>
          <SheetTitle>{user ? "Edit User" : "Add User"}</SheetTitle>
          <SheetDescription>
            {user ? `Update details for ${user.id}.` : "Create a new user to access the system."}
          </SheetDescription>
        </SheetHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6 py-6">
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
                    Unique identifier for the user. Cannot be changed.
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
                      <SelectItem value="admin">Admin (Full Access)</SelectItem>
                      <SelectItem value="editor">Editor (Can Configure)</SelectItem>
                      <SelectItem value="viewer">Viewer (Read Only)</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="space-y-4">
              <FormLabel>Authentication Method</FormLabel>
              <FormField
                control={form.control}
                name="authType"
                render={({ field }) => (
                  <Tabs
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                    value={field.value}
                    className="w-full"
                  >
                    <TabsList className="grid w-full grid-cols-2">
                      <TabsTrigger value="password">
                         <Lock className="mr-2 h-4 w-4" /> Password
                      </TabsTrigger>
                      <TabsTrigger value="apiKey">
                         <Key className="mr-2 h-4 w-4" /> API Key
                      </TabsTrigger>
                    </TabsList>

                    <TabsContent value="password" className="space-y-4 pt-4">
                      <FormField
                        control={form.control}
                        name="password"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>{user ? "New Password (Optional)" : "Password"}</FormLabel>
                            <div className="relative">
                              <FormControl>
                                <Input
                                    type={showPassword ? "text" : "password"}
                                    placeholder={user ? "Leave empty to keep unchanged" : "Min 8 characters"}
                                    {...field}
                                />
                              </FormControl>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                                onClick={() => setShowPassword(!showPassword)}
                                tabIndex={-1}
                              >
                                {showPassword ? (
                                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                                ) : (
                                  <Eye className="h-4 w-4 text-muted-foreground" />
                                )}
                              </Button>
                            </div>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </TabsContent>

                    <TabsContent value="apiKey" className="space-y-4 pt-4">
                      <Alert>
                        <AlertTriangle className="h-4 w-4" />
                        <AlertTitle>Security Warning</AlertTitle>
                        <AlertDescription>
                          API Keys are shown only once. If lost, you must generate a new one.
                        </AlertDescription>
                      </Alert>

                      {user && !generatedKey && (
                          <div className="text-sm text-muted-foreground italic">
                              This user currently has an API Key configured. Generating a new one will replace it.
                          </div>
                      )}

                      <div className="flex gap-2">
                         <div className="relative flex-1">
                            <Input
                                value={generatedKey}
                                readOnly
                                placeholder="Click Generate to create a key"
                                className="font-mono text-xs pr-10"
                            />
                         </div>
                         <Button type="button" onClick={generateApiKey} variant="secondary">
                            <RotateCw className="mr-2 h-4 w-4"/> Generate
                         </Button>
                      </div>

                      {generatedKey && (
                        <div className="animate-in fade-in slide-in-from-top-2 duration-300">
                             <Button type="button" variant="outline" className="w-full" onClick={copyKey}>
                                {copied ? <Check className="mr-2 h-4 w-4 text-green-500" /> : <Copy className="mr-2 h-4 w-4" />}
                                {copied ? "Copied!" : "Copy to Clipboard"}
                            </Button>
                            <p className="text-[10px] text-muted-foreground mt-2 text-center">
                                Please copy this key now. It will not be shown again.
                            </p>
                        </div>
                      )}
                    </TabsContent>
                  </Tabs>
                )}
              />
            </div>

            <SheetFooter className="pt-4">
               <Button variant="outline" type="button" onClick={onClose} disabled={isSubmitting}>
                 Cancel
               </Button>
               <Button type="submit" disabled={isSubmitting}>
                 {isSubmitting ? <RotateCw className="mr-2 h-4 w-4 animate-spin" /> : null}
                 {user ? "Save Changes" : "Create User"}
               </Button>
            </SheetFooter>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  );
}
