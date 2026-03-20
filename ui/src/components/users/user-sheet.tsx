/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { User } from "@proto/config/v1/user";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
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
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Key, Lock, ShieldAlert, Pencil, Eye, Copy, Check, RotateCw } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

const userSchema = z.object({
    id: z.string().min(3, "Username must be at least 3 characters").max(50),
    role: z.enum(["admin", "editor", "viewer"]),
    authType: z.enum(["password", "api_key"]),
    password: z.string().optional()
}).refine(data => {
    if (data.authType === "password" && !data.password) {
        return false;
    }
    return true;
}, {
    message: "Password is required for password authentication",
    path: ["password"]
});

type UserFormValues = z.infer<typeof userSchema>;

interface UserSheetProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    user: User | null;
    onSave: (user: Partial<User>, password?: string, apiKey?: string) => Promise<void>;
}

export function UserSheet({ open, onOpenChange, user, onSave }: UserSheetProps) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [generatedKey, setGeneratedKey] = useState("");
    const [copied, setCopied] = useState(false);
    const { toast } = useToast();

    const form = useForm<UserFormValues>({
        resolver: zodResolver(userSchema),
        defaultValues: {
            id: "",
            role: "viewer",
            authType: "password",
            password: "",
        }
    });

    useEffect(() => {
        if (open) {
            if (user) {
                form.reset({
                    id: user.id,
                    role: user.roles[0] || "viewer",
                    authType: user.authentication?.apiKey || (user.authentication as any)?.api_key ? "api_key" : "password",
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
            setGeneratedKey("");
            setCopied(false);
        }
    }, [open, user, form]);

    const authType = form.watch("authType");

    const generateApiKey = () => {
        const prefix = "sk-";
        const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
        let key = prefix;
        for (let i = 0; i < 32; i++) {
            key += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        setGeneratedKey(key);
        setCopied(false);
        // We do NOT set form value, API key is handled separately in onSave
    };

    const copyKey = () => {
        navigator.clipboard.writeText(generatedKey);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
        toast({ title: "Copied to clipboard" });
    };

    const onSubmit = async (data: UserFormValues) => {
        setIsSubmitting(true);
        try {
            let pwd = undefined;
            let key = undefined;

            if (data.authType === "password") {
                if (data.password) {
                    pwd = data.password;
                } else if (!user) {
                    toast({ variant: "destructive", title: "Password required" });
                    setIsSubmitting(false);
                    return;
                }
            } else if (data.authType === "api_key") {
                if (generatedKey) {
                    key = generatedKey;
                } else if (!user) {
                    toast({ variant: "destructive", title: "API Key required", description: "Generate a key first" });
                    setIsSubmitting(false);
                    return;
                }
            }

            await onSave({
                id: data.id,
                roles: [data.role],
            }, pwd, key);
        } catch (e) {
            // Error handled by parent
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="sm:max-w-md overflow-y-auto">
                <SheetHeader>
                    <SheetTitle>{user ? "Edit User" : "Add User"}</SheetTitle>
                    <SheetDescription>
                        {user ? "Modify user settings and permissions." : "Create a new user to access the system."}
                    </SheetDescription>
                </SheetHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 pt-6">
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
                                        Unique identifier for the user. Cannot be changed later.
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
