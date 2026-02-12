/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { User } from "@proto/config/v1/user";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { RotateCw, Copy, Check, ShieldAlert, Key } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface UserSheetProps {
    isOpen: boolean;
    onClose: () => void;
    user: User | null;
    onSave: (user: Partial<User>, password?: string, apiKey?: string) => Promise<void>;
}

// Validation schema
const userSchema = z.object({
    id: z.string()
        .min(3, "Username must be at least 3 characters")
        .max(50, "Username too long")
        .regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
    role: z.string().min(1, "Role is required"),
    authMethod: z.enum(["password", "api_key", "none"]),
    password: z.string().optional(),
});

type UserValues = z.infer<typeof userSchema>;

export function UserSheet({ isOpen, onClose, user, onSave }: UserSheetProps) {
    const [generatedKey, setGeneratedKey] = useState("");
    const [copied, setCopied] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    const form = useForm<UserValues>({
        resolver: zodResolver(userSchema),
        defaultValues: {
            id: "",
            role: "viewer",
            authMethod: "password",
            password: "",
        },
    });

    const authMethod = form.watch("authMethod");

    // Reset form when opening/closing or changing user
    useEffect(() => {
        if (isOpen) {
            setGeneratedKey("");
            setCopied(false);
            form.reset({
                id: user?.id || "",
                role: user?.roles?.[0] || "viewer",
                authMethod: user?.authentication?.apiKey ? "api_key" : "password",
                password: "",
            });
        }
    }, [isOpen, user, form]);

    const generateApiKey = () => {
        const array = new Uint8Array(24);
        window.crypto.getRandomValues(array);
        const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
        setGeneratedKey(key);
    };

    const copyKey = () => {
        navigator.clipboard.writeText(generatedKey);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const onSubmit = async (data: UserValues) => {
        // Validation for specific auth methods
        if (data.authMethod === "password") {
            if (!user && !data.password) {
                form.setError("password", { message: "Password is required for new users" });
                return;
            }
            if (data.password && data.password.length < 8) {
                form.setError("password", { message: "Password must be at least 8 characters" });
                return;
            }
        }

        if (data.authMethod === "api_key") {
            if (!user && !generatedKey) {
                // Should generate key
                 generateApiKey();
                 // Don't submit yet, user needs to copy it
                 return;
            }
            if (user && !user.authentication?.apiKey && !generatedKey) {
                 // Switching to API Key, must generate
                 generateApiKey();
                 return;
            }
        }

        setIsSaving(true);
        try {
            const partialUser: Partial<User> = {
                id: data.id,
                roles: [data.role],
            };

            // Pass the raw secrets to the parent handler to construct the complex Authentication object
            await onSave(partialUser, data.password, generatedKey);
            onClose();
        } catch (error) {
            console.error(error);
            // Form error could be set here if passed back
        } finally {
            setIsSaving(false);
        }
    };

    return (
        <Sheet open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <SheetContent className="sm:max-w-md overflow-y-auto">
                <SheetHeader>
                    <SheetTitle>{user ? "Edit User" : "Add User"}</SheetTitle>
                    <SheetDescription>
                        Configure user access and authentication details.
                    </SheetDescription>
                </SheetHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-6">
                        <FormField
                            control={form.control}
                            name="id"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Username / Service ID</FormLabel>
                                    <FormControl>
                                        <Input disabled={!!user} placeholder="e.g. jdoe or service-bot" {...field} />
                                    </FormControl>
                                    <FormDescription>
                                        Unique identifier for this account.
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
                                        Determines the permissions level.
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="authMethod"
                            render={({ field }) => (
                                <FormItem className="space-y-1">
                                    <FormLabel>Authentication Method</FormLabel>
                                    <Tabs onValueChange={field.onChange} defaultValue={field.value} value={field.value} className="w-full">
                                        <TabsList className="grid w-full grid-cols-2">
                                            <TabsTrigger value="password">Password</TabsTrigger>
                                            <TabsTrigger value="api_key">API Key</TabsTrigger>
                                        </TabsList>

                                        <TabsContent value="password" className="pt-4 space-y-4">
                                            <FormField
                                                control={form.control}
                                                name="password"
                                                render={({ field: passField }) => (
                                                    <FormItem>
                                                        <FormLabel>{user ? "New Password (Optional)" : "Password"}</FormLabel>
                                                        <FormControl>
                                                            <Input type="password" {...passField} />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </FormItem>
                                                )}
                                            />
                                        </TabsContent>

                                        <TabsContent value="api_key" className="pt-4 space-y-4">
                                            <div className="rounded-md border p-4 bg-muted/50">
                                                <div className="flex items-center gap-2 mb-2">
                                                    <Key className="h-4 w-4 text-amber-500" />
                                                    <span className="font-medium text-sm">Service Account Key</span>
                                                </div>
                                                <p className="text-xs text-muted-foreground mb-4">
                                                    API keys are used for machine-to-machine authentication.
                                                </p>

                                                {!generatedKey ? (
                                                     <div className="text-center">
                                                         {user?.authentication?.apiKey && (
                                                             <div className="mb-4 text-xs text-green-600 font-medium flex items-center justify-center gap-1">
                                                                 <Check className="h-3 w-3" /> Active Key Configured
                                                             </div>
                                                         )}
                                                         <Button type="button" onClick={generateApiKey} variant="outline" size="sm" className="w-full">
                                                             <RotateCw className="mr-2 h-3 w-3" />
                                                             {user?.authentication?.apiKey ? "Rotate Key" : "Generate Key"}
                                                         </Button>
                                                     </div>
                                                ) : (
                                                    <div className="space-y-3 animate-in fade-in zoom-in-95 duration-200">
                                                        <div className="relative">
                                                            <Input readOnly value={generatedKey} className="font-mono text-xs pr-20 bg-background" />
                                                            <Button
                                                                type="button"
                                                                variant="ghost"
                                                                size="sm"
                                                                className="absolute right-1 top-1 h-7 text-xs"
                                                                onClick={copyKey}
                                                            >
                                                                {copied ? <Check className="h-3 w-3 mr-1" /> : <Copy className="h-3 w-3 mr-1" />}
                                                                {copied ? "Copied" : "Copy"}
                                                            </Button>
                                                        </div>
                                                        <Alert variant="destructive" className="py-2">
                                                            <ShieldAlert className="h-4 w-4" />
                                                            <AlertTitle className="text-xs font-semibold">Save this key now!</AlertTitle>
                                                            <AlertDescription className="text-[10px]">
                                                                It will not be shown again after you save.
                                                            </AlertDescription>
                                                        </Alert>
                                                    </div>
                                                )}
                                            </div>
                                        </TabsContent>
                                    </Tabs>
                                </FormItem>
                            )}
                        />

                        <SheetFooter>
                            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
                            <Button type="submit" disabled={isSaving || (authMethod === 'api_key' && !generatedKey && !user?.authentication?.apiKey)}>
                                {isSaving ? "Saving..." : "Save User"}
                            </Button>
                        </SheetFooter>
                    </form>
                </Form>
            </SheetContent>
        </Sheet>
    );
}
