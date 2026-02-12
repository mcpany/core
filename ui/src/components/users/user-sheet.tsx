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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { RotateCw, Copy, Check, Eye, EyeOff, AlertTriangle } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { User } from "@proto/config/v1/user";

// Validation schema
const userSchema = z.object({
  id: z
    .string()
    .min(3, "Username must be at least 3 characters")
    .max(50, "Username too long")
    .regex(/^[a-zA-Z0-9_-]+$/, "Username can only contain letters, numbers, underscores, and dashes"),
  role: z.string().min(1, "Role is required"),
  authType: z.enum(["password", "api_key"]),
  password: z.string().optional(),
  apiKey: z.string().optional(),
}).refine((data) => {
  if (data.authType === "password" && !data.password) {
      // Only require password for new users?
      // Ideally we check if we are editing, but here we can just say if it's password type, it needs a password
      // UNLESS we are editing and want to keep existing.
      return true; // We'll handle "required for new" in the component logic
  }
  if (data.authType === "api_key" && !data.apiKey) {
      return false;
  }
  return true;
}, {
  message: "API Key is required",
  path: ["apiKey"],
});

type UserValues = z.infer<typeof userSchema>;

interface UserSheetProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  user: User | null;
  onSave: () => void;
}

export function UserSheet({ isOpen, onOpenChange, user, onSave }: UserSheetProps) {
  const { toast } = useToast();
  const [generatedKey, setGeneratedKey] = useState("");
  const [copied, setCopied] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const form = useForm<UserValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      id: "",
      role: "viewer",
      authType: "password",
      password: "",
      apiKey: "",
    },
  });

  // Reset form when dialog opens/closes or editing user changes
  useEffect(() => {
    if (isOpen) {
      setGeneratedKey("");
      setCopied(false);
      setShowPassword(false);

      if (user) {
        // Editing existing user
        let authType: "password" | "api_key" = "password";
        if (user.authentication?.apiKey) {
             authType = "api_key";
        }

        form.reset({
          id: user.id,
          role: user.roles[0] || "viewer",
          authType: authType,
          password: "", // Always empty for security
          apiKey: "",
        });
      } else {
        // New User
        form.reset({
          id: "",
          role: "viewer",
          authType: "password",
          password: "",
          apiKey: "",
        });
      }
    }
  }, [isOpen, user, form]);

  const generateApiKey = () => {
    const array = new Uint8Array(24);
    window.crypto.getRandomValues(array);
    const key = "mcp_sk_" + Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
    setGeneratedKey(key);
    form.setValue("apiKey", key);
    form.clearErrors("apiKey");
  };

  const copyKey = () => {
    navigator.clipboard.writeText(generatedKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const onSubmit = async (data: UserValues) => {
    // Custom validation for password requirement on new users
    if (!user && data.authType === "password" && !data.password) {
        form.setError("password", { message: "Password is required for new users" });
        return;
    }

    // Validate password length if provided
    if (data.authType === "password" && data.password && data.password.length < 8) {
        form.setError("password", { message: "Password must be at least 8 characters" });
        return;
    }

    try {
        let authConfig: any = {};

        if (data.authType === "password") {
            if (data.password) {
                authConfig = {
                    basicAuth: {
                        username: data.id,
                        password: { plainText: data.password }
                    }
                };
            } else if (user?.authentication?.basicAuth) {
                 // Keep existing
                 authConfig = user.authentication;
            }
        } else {
            // API Key
            if (generatedKey) {
                authConfig = {
                    apiKey: {
                        paramName: "X-API-Key",
                        in: 0, // HEADER
                        verificationValue: generatedKey
                    }
                };
            } else if (user?.authentication?.apiKey) {
                 // Keep existing
                 authConfig = user.authentication;
            } else {
                 // Switching to API Key but didn't generate one
                 toast({ variant: "destructive", title: "Error", description: "Please generate an API Key." });
                 return;
            }
        }

        const userPayload: any = {
            id: data.id,
            roles: [data.role],
            authentication: authConfig
        };

        if (user) {
            await apiClient.updateUser(userPayload);
            toast({ title: "User Updated", description: `User ${data.id} has been updated.` });
        } else {
            await apiClient.createUser(userPayload);
            toast({ title: "User Created", description: `User ${data.id} has been created.` });
        }
        onSave();
        onOpenChange(false);
    } catch (e: any) {
        console.error("Failed to save user", e);
        toast({ variant: "destructive", title: "Error", description: e.message || "Failed to save user." });
    }
  };

  const authType = form.watch("authType");

  return (
    <Sheet open={isOpen} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md w-full overflow-y-auto">
        <SheetHeader className="mb-6">
          <SheetTitle>{user ? "Edit User" : "Add New User"}</SheetTitle>
          <SheetDescription>
            {user ? "Modify user details and permissions." : "Create a new user to access the system."}
          </SheetDescription>
        </SheetHeader>

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
                      <SelectItem value="admin">Admin</SelectItem>
                      <SelectItem value="editor">Editor</SelectItem>
                      <SelectItem value="viewer">Viewer</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Determines the user's access level.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="space-y-4 border rounded-lg p-4 bg-muted/20">
                <FormField
                  control={form.control}
                  name="authType"
                  render={({ field }) => (
                    <FormItem className="space-y-3">
                      <FormLabel>Authentication Method</FormLabel>
                      <FormControl>
                          <Tabs value={field.value} onValueChange={field.onChange} className="w-full">
                            <TabsList className="grid w-full grid-cols-2">
                                <TabsTrigger value="password">Password</TabsTrigger>
                                <TabsTrigger value="api_key">API Key</TabsTrigger>
                            </TabsList>
                          </Tabs>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {authType === "password" && (
                    <FormField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>{user ? "New Password (Optional)" : "Password"}</FormLabel>
                        <FormControl>
                            <div className="relative">
                                <Input
                                    type={showPassword ? "text" : "password"}
                                    placeholder={user ? "Leave blank to keep existing" : "Min. 8 characters"}
                                    {...field}
                                />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
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
                )}

                {authType === "api_key" && (
                    <div className="space-y-4">
                        <div className="text-sm text-muted-foreground">
                            Generate a secure API Key for this user (e.g. for AI Agents).
                        </div>

                        {generatedKey ? (
                            <div className="space-y-2 animate-in fade-in slide-in-from-top-2">
                                <div className="p-3 bg-muted rounded-md border border-dashed border-primary/50 relative">
                                    <p className="font-mono text-xs break-all pr-8 text-primary">
                                        {generatedKey}
                                    </p>
                                </div>
                                <Button type="button" variant="secondary" className="w-full" onClick={copyKey}>
                                    {copied ? <Check className="mr-2 h-4 w-4 text-green-500" /> : <Copy className="mr-2 h-4 w-4" />}
                                    {copied ? "Copied!" : "Copy Key"}
                                </Button>
                                <div className="flex items-start gap-2 text-[10px] text-amber-500 bg-amber-500/10 p-2 rounded">
                                    <AlertTriangle className="h-3 w-3 shrink-0 mt-0.5" />
                                    <p>Make sure to copy this key now. You won't be able to see it again!</p>
                                </div>
                            </div>
                        ) : (
                            <div className="flex flex-col gap-2">
                                {user?.authentication?.apiKey && (
                                    <div className="text-xs text-muted-foreground bg-green-500/10 text-green-600 p-2 rounded flex items-center gap-2">
                                        <Check className="h-3 w-3" /> Existing API Key is active.
                                    </div>
                                )}
                                <Button type="button" onClick={generateApiKey} variant="outline" className="w-full border-dashed">
                                    <RotateCw className="mr-2 h-4 w-4"/>
                                    {user?.authentication?.apiKey ? "Rotate Key (Generate New)" : "Generate Key"}
                                </Button>
                            </div>
                        )}
                    </div>
                )}
            </div>

            <div className="flex justify-end pt-4">
                <Button type="button" variant="ghost" onClick={() => onOpenChange(false)} className="mr-2">Cancel</Button>
                <Button type="submit">Save User</Button>
            </div>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  );
}
