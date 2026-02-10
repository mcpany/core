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
    FormDescription,
} from "@/components/ui/form";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { User } from "./user-list";
import { Copy, RotateCw, Check, AlertTriangle } from "lucide-react";

// Validation schema
const userSchema = z.object({
    id: z.string()
        .min(3, "Username must be at least 3 characters")
        .max(50, "Username too long")
        .regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
    role: z.string().min(1, "Role is required"),
    password: z.string().optional(),
});

type UserValues = z.infer<typeof userSchema>;

interface UserSheetProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    user: User | null;
    onSave: (user: UserValues, authType: "password" | "api_key", generatedKey?: string) => Promise<void>;
}

export function UserSheet({ open, onOpenChange, user, onSave }: UserSheetProps) {
    const [authType, setAuthType] = useState<"password" | "api_key">("password");
    const [generatedKey, setGeneratedKey] = useState("");
    const [copied, setCopied] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

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
            form.reset({
                id: user?.id || "",
                role: user?.roles[0] || "viewer",
                password: "",
            });

            if (user?.authentication?.api_key) {
                setAuthType("api_key");
            } else {
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
    };

    const handleSubmit = async (data: UserValues) => {
        // Custom validation
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
            // API Key
            if (!user && !generatedKey) {
                // Must generate key for new user
                form.setError("root", { message: "Please generate an API Key" });
                return;
            }
             // If switching to API key for existing user, we need a key if one doesn't exist?
             // Or if they want to rotate it.
             // If user already has API key, they might just be updating role.
             if (user && !user.authentication?.api_key && !generatedKey) {
                  form.setError("root", { message: "Please generate an API Key to switch authentication method" });
                  return;
             }
        }

        setIsSaving(true);
        try {
            await onSave(data, authType, generatedKey);
            onOpenChange(false);
        } catch (error) {
            console.error(error);
            form.setError("root", { message: "Failed to save user" });
        } finally {
            setIsSaving(false);
        }
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="sm:max-w-md overflow-y-auto">
                <SheetHeader>
                    <SheetTitle>{user ? "Edit User" : "Add User"}</SheetTitle>
                    <SheetDescription>
                        {user ? "Update user details and permissions." : "Create a new user to access the system."}
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
                                            <SelectItem value="admin">Admin</SelectItem>
                                            <SelectItem value="editor">Editor</SelectItem>
                                            <SelectItem value="viewer">Viewer</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormDescription>
                                        Determines the user's permissions.
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <div className="space-y-4 rounded-lg border p-4">
                            <FormLabel className="text-base">Authentication Method</FormLabel>
                            <Tabs value={authType} onValueChange={(v) => setAuthType(v as any)} className="w-full">
                                <TabsList className="grid w-full grid-cols-2">
                                    <TabsTrigger value="password">Password</TabsTrigger>
                                    <TabsTrigger value="api_key">API Key</TabsTrigger>
                                </TabsList>
                                <TabsContent value="password" className="pt-4 space-y-4">
                                    <FormField
                                        control={form.control}
                                        name="password"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>{user ? "New Password (Optional)" : "Password"}</FormLabel>
                                                <FormControl>
                                                    <Input type="password" placeholder={user ? "Leave blank to keep current" : "Min. 8 characters"} {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </TabsContent>
                                <TabsContent value="api_key" className="pt-4 space-y-4">
                                    <div className="text-sm text-muted-foreground">
                                        API Keys are suitable for machine-to-machine communication (Agents).
                                    </div>

                                    {user?.authentication?.api_key && !generatedKey && (
                                        <div className="flex items-center gap-2 p-2 bg-yellow-500/10 text-yellow-600 rounded-md text-sm border border-yellow-500/20">
                                            <AlertTriangle className="h-4 w-4 shrink-0" />
                                            <span>User already has an API Key configured. Generating a new one will replace it.</span>
                                        </div>
                                    )}

                                    <div className="flex gap-2">
                                        <div className="relative flex-1">
                                             <Input
                                                value={generatedKey}
                                                readOnly
                                                placeholder="Click generate..."
                                                className="font-mono text-xs pr-10"
                                             />
                                        </div>
                                        <Button type="button" onClick={generateApiKey} variant="secondary" size="sm">
                                            <RotateCw className="mr-2 h-3 w-3"/> Generate
                                        </Button>
                                    </div>

                                    {generatedKey && (
                                        <div className="space-y-2 animate-in fade-in slide-in-from-top-2">
                                            <Button type="button" variant="outline" className="w-full" onClick={copyKey}>
                                                {copied ? <Check className="mr-2 h-4 w-4 text-green-500" /> : <Copy className="mr-2 h-4 w-4" />}
                                                {copied ? "Copied to Clipboard" : "Copy Key"}
                                            </Button>
                                            <p className="text-[10px] text-red-500 font-medium">
                                                This key will not be shown again. Please copy it now.
                                            </p>
                                        </div>
                                    )}
                                </TabsContent>
                            </Tabs>
                        </div>

                        {form.formState.errors.root && (
                             <p className="text-sm font-medium text-destructive">{form.formState.errors.root.message}</p>
                        )}

                        <SheetFooter>
                             <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
                             <Button type="submit" disabled={isSaving}>{isSaving ? "Saving..." : "Save User"}</Button>
                        </SheetFooter>
                    </form>
                </Form>
            </SheetContent>
        </Sheet>
    );
}
