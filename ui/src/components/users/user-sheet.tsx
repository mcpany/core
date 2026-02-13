/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { User } from "@proto/config/v1/user";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetFooter
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
import { RotateCw, Copy, Check, Key, Lock, ShieldAlert, Eye, Pencil } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface UserSheetProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    user: User | null;
    onSave: (user: Partial<User>, password?: string, apiKey?: string) => Promise<void>;
}

const userSchema = z.object({
    id: z.string().min(3, "Username must be at least 3 characters").max(50, "Username too long").regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
    role: z.string().min(1, "Role is required"),
    authType: z.enum(["password", "api_key"]),
    password: z.string().optional(),
});

type UserValues = z.infer<typeof userSchema>;

/**
 * UserSheet component provides a slide-out form for creating and editing users.
 *
 * @param props - The component props.
 * @param props.open - Whether the sheet is open.
 * @param props.onOpenChange - Callback when open state changes.
 * @param props.user - The user object to edit (or null for new user).
 * @param props.onSave - Callback to save user details.
 */
export function UserSheet({ open, onOpenChange, user, onSave }: UserSheetProps) {
    const { toast } = useToast();
    const [generatedKey, setGeneratedKey] = useState("");
    const [copied, setCopied] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const form = useForm<UserValues>({
        resolver: zodResolver(userSchema),
        defaultValues: {
            id: "",
            role: "viewer",
            authType: "password",
            password: "",
        },
    });

    const authType = form.watch("authType");

    useEffect(() => {
        if (open) {
            setGeneratedKey("");
            setCopied(false);
            if (user) {
                form.reset({
                    id: user.id,
                    role: user.roles[0] || "viewer",
                    authType: user.authentication?.apiKey ? "api_key" : "password",
                    password: "",
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
    }, [open, user, form]);

    const generateApiKey = () => {
        const array = new Uint8Array(24);
        window.crypto.getRandomValues(array);
        const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
        setGeneratedKey(key);
        form.setValue("password", ""); // Clear password if switching
    };

    const copyKey = () => {
        navigator.clipboard.writeText(generatedKey);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
        toast({
            description: "API Key copied to clipboard",
        });
    };

    const onSubmit = async (data: UserValues) => {
        setIsSubmitting(true);
        try {
            // Validation Logic
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
                         description: "Please generate an API Key for the new user."
                     });
                     setIsSubmitting(false);
                     return;
                }
            }

            const userUpdate: Partial<User> = {
                id: data.id,
                roles: [data.role],
            };

            await onSave(userUpdate, data.password, generatedKey);
            onOpenChange(false);
        } catch (error) {
            console.error(error);
             toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save user."
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
                        {user ? "Update user details and permissions." : "Create a new user to access the platform."}
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
                                                    <ShieldAlert className="h-4 w-4 text-destructive" />
                                                    <span>Admin</span>
                                                </div>
                                            </SelectItem>
                                            <SelectItem value="editor">
                                                <div className="flex items-center gap-2">
                                                    <Pencil className="h-4 w-4 text-primary" />
                                                    <span>Editor</span>
                                                </div>
                                            </SelectItem>
                                            <SelectItem value="viewer">
                                                <div className="flex items-center gap-2">
                                                    <Eye className="h-4 w-4 text-muted-foreground" />
                                                    <span>Viewer</span>
                                                </div>
                                            </SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormDescription>
                                        Determines the user's permissions.
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="authType"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Authentication Method</FormLabel>
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
                                            <TabsTrigger value="api_key">
                                                <Key className="mr-2 h-4 w-4" /> API Key
                                            </TabsTrigger>
                                        </TabsList>

                                        <TabsContent value="password" className="pt-4 space-y-4">
                                             <FormField
                                                control={form.control}
                                                name="password"
                                                render={({ field: passField }) => (
                                                    <FormItem>
                                                        <FormLabel>{user ? "New Password" : "Password"}</FormLabel>
                                                        <FormControl>
                                                            <Input type="password" placeholder={user ? "Leave blank to keep current" : "Minimum 8 characters"} {...passField} />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </FormItem>
                                                )}
                                            />
                                            {user && (
                                                <p className="text-xs text-muted-foreground">
                                                    Only enter a password if you want to change it.
                                                </p>
                                            )}
                                        </TabsContent>

                                        <TabsContent value="api_key" className="pt-4 space-y-4">
                                            <Alert>
                                                <Key className="h-4 w-4" />
                                                <AlertTitle>API Key Access</AlertTitle>
                                                <AlertDescription>
                                                    Ideal for agents and automated tools.
                                                </AlertDescription>
                                            </Alert>

                                            <div className="space-y-2">
                                                <div className="flex justify-between items-center">
                                                    <FormLabel>Generated Key</FormLabel>
                                                    <Button type="button" variant="outline" size="sm" onClick={generateApiKey}>
                                                        <RotateCw className="mr-2 h-3 w-3" />
                                                        {generatedKey ? "Regenerate" : "Generate Key"}
                                                    </Button>
                                                </div>

                                                {generatedKey ? (
                                                    <div className="relative">
                                                        <div className="p-3 bg-muted rounded-md font-mono text-xs break-all pr-10 border border-primary/20 bg-primary/5">
                                                            {generatedKey}
                                                        </div>
                                                        <Button
                                                            type="button"
                                                            variant="ghost"
                                                            size="icon"
                                                            className="absolute right-1 top-1 h-7 w-7 text-muted-foreground hover:text-foreground"
                                                            onClick={copyKey}
                                                        >
                                                            {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                                                        </Button>
                                                    </div>
                                                ) : (
                                                    <div className="p-8 border-2 border-dashed rounded-md flex flex-col items-center justify-center text-muted-foreground bg-muted/10">
                                                        <Key className="h-8 w-8 mb-2 opacity-20" />
                                                        <span className="text-xs">Click generate to create a secure key</span>
                                                    </div>
                                                )}

                                                {generatedKey && (
                                                    <p className="text-[10px] text-destructive font-medium mt-1 animate-pulse">
                                                        Warning: This key will only be shown once. Copy it now.
                                                    </p>
                                                )}
                                                {user && !generatedKey && (
                                                    <p className="text-xs text-muted-foreground">
                                                        Existing API key is hidden. Generate a new one to rotate.
                                                    </p>
                                                )}
                                            </div>
                                        </TabsContent>
                                    </Tabs>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <SheetFooter className="pt-4">
                            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? "Saving..." : "Save Changes"}
                            </Button>
                        </SheetFooter>
                    </form>
                </Form>
            </SheetContent>
        </Sheet>
    );
}
