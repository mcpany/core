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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { apiClient } from "@/lib/client";

const authSettingsSchema = z.object({
  oidc_issuer: z.string().optional(),
  oidc_client_id: z.string().optional(),
  oidc_client_secret: z.string().optional(),
  oidc_redirect_url: z.string().optional(),
});

type AuthSettingsValues = z.infer<typeof authSettingsSchema>;

export function AuthSettingsForm() {
  const [loading, setLoading] = useState(false);

  const form = useForm<AuthSettingsValues>({
    resolver: zodResolver(authSettingsSchema),
    defaultValues: {
      oidc_issuer: "",
      oidc_client_id: "",
      oidc_client_secret: "",
      oidc_redirect_url: "",
    },
  });

  useEffect(() => {
    async function loadSettings() {
      try {
        const settings = await apiClient.getGlobalSettings();
        if (settings?.oidc) {
          form.reset({
            oidc_issuer: settings.oidc.issuer || "",
            oidc_client_id: settings.oidc.client_id || "",
            oidc_client_secret: settings.oidc.client_secret || "",
            oidc_redirect_url: settings.oidc.redirect_url || "",
          });
        }
      } catch (e) {
        console.error("Failed to load auth settings", e);
      }
    }
    loadSettings();
  }, [form]);

  async function onSubmit(data: AuthSettingsValues) {
    setLoading(true);
    try {
      // Fetch current settings first to avoid overwriting other fields
      // In a real implementation we might have a dedicated patch endpoint or partial update
      const current = await apiClient.getGlobalSettings();

      const payload = {
        ...current,
        oidc: {
            issuer: data.oidc_issuer,
            client_id: data.oidc_client_id,
            client_secret: data.oidc_client_secret,
            redirect_url: data.oidc_redirect_url,
        }
      };

      await apiClient.saveGlobalSettings(payload);
    } catch (e) {
      console.error("Failed to save auth settings", e);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="backdrop-blur-sm bg-background/50">
      <CardHeader>
        <CardTitle>Authentication Configuration</CardTitle>
        <CardDescription>Configure OpenID Connect (OIDC) settings.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <div className="grid grid-cols-1 gap-6">
                <FormField
                control={form.control}
                name="oidc_issuer"
                render={({ field }) => (
                    <FormItem>
                    <FormLabel>Issuer URL</FormLabel>
                    <FormControl>
                        <Input placeholder="https://accounts.google.com" {...field} />
                    </FormControl>
                    <FormDescription>
                        The OIDC Issuer URL (e.g. Auth0, Google, Keycloak).
                    </FormDescription>
                    <FormMessage />
                    </FormItem>
                )}
                />

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <FormField
                    control={form.control}
                    name="oidc_client_id"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Client ID</FormLabel>
                        <FormControl>
                            <Input placeholder="..." {...field} />
                        </FormControl>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                    <FormField
                    control={form.control}
                    name="oidc_client_secret"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Client Secret</FormLabel>
                        <FormControl>
                            <Input type="password" placeholder="..." {...field} />
                        </FormControl>
                         <FormDescription>
                           Stored securely.
                        </FormDescription>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                </div>

                 <FormField
                control={form.control}
                name="oidc_redirect_url"
                render={({ field }) => (
                    <FormItem>
                    <FormLabel>Redirect URL</FormLabel>
                    <FormControl>
                        <Input placeholder="http://localhost:8080/callback" {...field} />
                    </FormControl>
                     <FormDescription>
                        The callback URL allowed in your IdP.
                    </FormDescription>
                    <FormMessage />
                    </FormItem>
                )}
                />
            </div>
            <Button type="submit" disabled={loading}>
                {loading ? "Saving..." : "Save Settings"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
