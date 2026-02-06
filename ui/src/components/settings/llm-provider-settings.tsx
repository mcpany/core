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
import { useToast } from "@/hooks/use-toast";

const llmSchema = z.object({
  type: z.enum(["openai", "anthropic", "gemini"]),
  api_key: z.string().min(1, "API Key is required"),
  model: z.string().optional(),
});

type LLMValues = z.infer<typeof llmSchema>;

/**
 * LLMProviderSettings component.
 * Manages configuration for LLM providers (OpenAI, Anthropic, Gemini).
 */
export function LLMProviderSettings() {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  const form = useForm<LLMValues>({
    resolver: zodResolver(llmSchema),
    defaultValues: {
      type: "openai",
      api_key: "",
      model: "gpt-4",
    },
  });

  const providerType = form.watch("type");

  useEffect(() => {
    async function loadSettings() {
      try {
        const configs = await apiClient.getLLMProviders();
        const currentType = form.getValues("type");
        // Check if configs is an array
        if (Array.isArray(configs)) {
            const found = configs.find((c: any) => c.type === currentType);
            if (found) {
                form.setValue("api_key", found.api_key || "");
                form.setValue("model", found.model || "");
            } else {
                form.setValue("api_key", "");
                form.setValue("model", getDefaultModel(currentType));
            }
        }
      } catch (e) {
        console.error("Failed to load LLM settings", e);
      }
    }
    loadSettings();
  }, [providerType, form]);

  function getDefaultModel(type: string) {
      switch(type) {
          case "openai": return "gpt-4";
          case "anthropic": return "claude-3-opus-20240229";
          case "gemini": return "gemini-pro";
          default: return "";
      }
  }

  async function onSubmit(data: LLMValues) {
    setLoading(true);
    try {
      await apiClient.saveLLMProvider(data);
      toast({
        title: "Settings saved",
        description: "LLM provider configuration updated successfully.",
      });
    } catch (e) {
      console.error("Failed to save settings", e);
      toast({
        title: "Error",
        description: "Failed to save settings.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="backdrop-blur-sm bg-background/50">
      <CardHeader>
        <CardTitle>AI Provider Configuration</CardTitle>
        <CardDescription>Configure LLM providers for Auto Craft and other AI features.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Provider</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a provider" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="openai">OpenAI</SelectItem>
                      <SelectItem value="anthropic">Anthropic (Claude)</SelectItem>
                      <SelectItem value="gemini">Google Gemini</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="api_key"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>API Key</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="sk-..." {...field} />
                  </FormControl>
                  <FormDescription>
                    Your API key. It will be stored securely.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="model"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Default Model</FormLabel>
                  <FormControl>
                    <Input placeholder={getDefaultModel(providerType)} {...field} />
                  </FormControl>
                  <FormDescription>
                    The model to use (optional).
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" disabled={loading}>
              {loading ? "Saving..." : "Save Configuration"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
