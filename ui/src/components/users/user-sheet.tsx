/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiClient } from "@/lib/client";
import { RotateCw, Copy, Check, Key, AlertTriangle, ShieldCheck } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";

/**
 * User interface matching the backend definition (simplified).
 */
export interface User {
    id: string;
    roles: string[];
    authentication?: {
        basic_auth?: { password_hash?: string };
        api_key?: {
            param_name?: string;
            verification_value?: string;
        };
    };
}

const userSchema = z.object({
    id: z.string().min(3, "Username must be at least 3 characters").max(50, "Username too long").regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
    role: z.string().min(1, "Role is required"),
    password: z.string().optional(),
});

type UserValues = z.infer<typeof userSchema>;

interface UserSheetProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    user: User | null;
    onSave: () => void;
}

/**
 * UserSheet component for creating and editing users.
 *
 * @param props - The component props.
 * @param props.open - Whether the sheet is open.
 * @param props.onOpenChange - Callback for changing the open state.
 * @param props.user - The user being edited, or null for new user.
 * @param props.onSave - Callback triggered after successful save.
 * @returns The rendered component.
 */
export function UserSheet({ open, onOpenChange, user, onSave }: UserSheetProps) {
    const { toast } = useToast();
    const [authType, setAuthType] = useState<"password" | "api_key">("password");
    const [generatedKey, setGeneratedKey] = useState("");
    const [copied, setCopied] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const form = useForm<UserValues>({
        resolver: zodResolver(userSchema),
        defaultValues: {
            id: "",
            role: "viewer",
            password: "",
        },
    });

    useEffect(() => {
        if (open) {
            setGeneratedKey("");
            setCopied(false);
            if (user) {
                form.reset({
                    id: user.id,
                    role: user.roles[0] || "viewer",
                    password: "",
                });
                if (user.authentication?.api_key) {
                    setAuthType("api_key");
                } else {
                    setAuthType("password");
                }
            } else {
                form.reset({
                    id: "",
                    role: "viewer",
                    password: "",
                });
                setAuthType("password");
            }
        }
    }, [open, user, form]);

    const generateApiKey = () => {
        const array = new Uint8Array(24);
        window.crypto.getRandomValues(array);
        const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
        setGeneratedKey(key);
        form.clearErrors("password");
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
        // Validation logic
        if (authType === "password") {
            if (!user && !data.password) {
                form.setError("password", { message: "Password is required for new users" });
                return;
            }
            if (data.password && data.password.length < 8) {
                form.setError("password", { message: "Password must be at least 8 characters" });
                return;
            }
        } else {
            // API Key Mode
            if (!user && !generatedKey) {
                toast({
                    title: "Missing API Key",
                    description: "You must generate an API Key for this user.",
                    variant: "destructive"
                });
                return;
            }
        }

        setIsSubmitting(true);
        try {
            let authConfig = user?.authentication;

            if (authType === "password") {
                if (data.password) {
                    authConfig = {
                        basic_auth: {
                            password_hash: data.password // Server handles hashing
                        }
                    };
                }
            } else {
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

            // If switching auth types and no new credentials provided, warn or fail?
            // Assuming for now if they switch tabs they intend to switch auth methods.
            // If editing and switching to API Key but didn't generate one, we should stop them above.

            const userPayload = {
                id: data.id,
                roles: [data.role], // Single role for now
                authentication: authConfig
            };

            if (user) {
                await apiClient.updateUser(userPayload);
                toast({
                    title: "User Updated",
                    description: `Successfully updated user ${data.id}.`,
                });
            } else {
                await apiClient.createUser(userPayload);
                toast({
                    title: "User Created",
                    description: `Successfully created user ${data.id}.`,
                });
            }
            onSave();
            onOpenChange(false);
        } catch (error) {
            console.error("Failed to save user", error);
            toast({
                title: "Error",
                description: "Failed to save user details. Please try again.",
                variant: "destructive"
            });
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="sm:max-w-md overflow-y-auto">
                <SheetHeader>
                    <SheetTitle>{user ? "Edit User" : "Add New User"}</SheetTitle>
                    <SheetDescription>
                        {user ? "Update user details and permissions." : "Create a new user account for accessing MCP Any."}
                    </SheetDescription>
                </SheetHeader>

                <div className="py-6">
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                            <FormField
                                control={form.control}
                                name="id"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Username</FormLabel>
                                        <FormControl>
                                            <Input disabled={!!user} placeholder="jdoe" {...field} />
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
                                                <SelectItem value="admin">
                                                    <div className="flex items-center gap-2">
                                                        <ShieldCheck className="h-4 w-4 text-primary" />
                                                        <span>Admin</span>
                                                    </div>
                                                </SelectItem>
                                                <SelectItem value="editor">Editor</SelectItem>
                                                <SelectItem value="viewer">Viewer</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormDescription>
                                            Determines access level within the platform.
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <div className="space-y-3 pt-2">
                                <FormLabel>Authentication Method</FormLabel>
                                <Tabs value={authType} onValueChange={(v) => setAuthType(v as any)} className="w-full">
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
                                                        <Input type="password" placeholder={user ? "••••••••" : "Min. 8 characters"} {...field} />
                                                    </FormControl>
                                                    <FormDescription>
                                                        {user ? "Leave blank to keep current password." : "Required for password authentication."}
                                                    </FormDescription>
                                                    <FormMessage />
                                                </FormItem>
                                            )}
                                        />
                                    </TabsContent>
                                    <TabsContent value="api_key" className="space-y-4 pt-4">
                                        <div className="rounded-md bg-muted/50 p-4 border text-sm">
                                            <div className="flex items-center gap-2 mb-2 font-medium">
                                                <Key className="h-4 w-4 text-primary" />
                                                Generate API Key
                                            </div>
                                            <p className="text-muted-foreground mb-4">
                                                Use this for automated agents or CLI tools. The key will only be shown once.
                                            </p>

                                            <div className="flex gap-2 mb-4">
                                                <Input
                                                    value={generatedKey}
                                                    readOnly
                                                    placeholder={user?.authentication?.api_key ? "Key exists (hidden)" : "Click generate..."}
                                                    className="font-mono text-xs bg-background"
                                                />
                                                <Button type="button" onClick={generateApiKey} variant="secondary">
                                                    <RotateCw className="mr-2 h-4 w-4" /> Generate
                                                </Button>
                                            </div>

                                            {generatedKey && (
                                                <div className="space-y-2 animate-in fade-in slide-in-from-top-2">
                                                    <Button type="button" variant="outline" className="w-full border-primary/20 hover:bg-primary/5" onClick={copyKey}>
                                                        {copied ? <Check className="mr-2 h-4 w-4 text-green-500" /> : <Copy className="mr-2 h-4 w-4" />}
                                                        {copied ? "Copied!" : "Copy to Clipboard"}
                                                    </Button>
                                                    <div className="flex items-start gap-2 text-xs text-amber-500 bg-amber-500/10 p-2 rounded">
                                                        <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
                                                        <p>
                                                            Please copy this key now. You won&apos;t be able to see it again after saving.
                                                        </p>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    </TabsContent>
                                </Tabs>
                            </div>

                            <SheetFooter className="pt-4">
                                <Button type="submit" disabled={isSubmitting}>
                                    {isSubmitting && <RotateCw className="mr-2 h-4 w-4 animate-spin" />}
                                    {user ? "Save Changes" : "Create User"}
                                </Button>
                            </SheetFooter>
                        </form>
                    </Form>
                </div>
            </SheetContent>
        </Sheet>
    );
}
