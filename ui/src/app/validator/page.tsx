/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Textarea } from "@/components/ui/textarea";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";
import { Loader2, CheckCircle2, XCircle, AlertTriangle } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const formSchema = z.object({
  configJson: z.string().min(1, "Configuration JSON is required"),
});

export default function ValidatorPage() {
  const [isValidating, setIsValidating] = useState(false);
  const [validationResult, setValidationResult] = useState<{valid: boolean, message: string} | null>(null);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      configJson: "{\n  \"name\": \"my-service\",\n  \"httpService\": {\n    \"address\": \"https://api.example.com\"\n  }\n}",
    },
  });

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    setIsValidating(true);
    setValidationResult(null);

    try {
      let config: UpstreamServiceConfig;
      try {
        config = JSON.parse(values.configJson);
      } catch (e) {
        setValidationResult({ valid: false, message: "Invalid JSON syntax" });
        setIsValidating(false);
        return;
      }

      const response = await apiClient.validateService(config);
      setValidationResult({ valid: response.valid, message: response.message });
    } catch (error: any) {
      setValidationResult({ valid: false, message: error.message || "An unexpected error occurred" });
    } finally {
      setIsValidating(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight">Schema Validator</h1>
        <p className="text-muted-foreground">
          Validate your Upstream Service configuration against the server schema.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="h-full flex flex-col">
            <CardHeader>
                <CardTitle>Configuration Editor</CardTitle>
                <CardDescription>Paste your JSON configuration here.</CardDescription>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col">
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 flex-1 flex flex-col">
                        <FormField
                            control={form.control}
                            name="configJson"
                            render={({ field }) => (
                                <FormItem className="flex-1 flex flex-col">
                                    <FormControl>
                                        <Textarea
                                            className="font-mono flex-1 min-h-[400px]"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <Button type="submit" disabled={isValidating} className="w-full">
                            {isValidating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : "Validate Configuration"}
                        </Button>
                    </form>
                </Form>
            </CardContent>
        </Card>

        <Card className="h-full">
            <CardHeader>
                <CardTitle>Validation Result</CardTitle>
                <CardDescription>Real-time feedback on your configuration.</CardDescription>
            </CardHeader>
            <CardContent>
                {!validationResult && !isValidating && (
                    <div className="flex flex-col items-center justify-center h-[300px] text-muted-foreground border-2 border-dashed rounded-lg">
                        <AlertTriangle className="h-10 w-10 mb-2 opacity-50" />
                        <p>No validation run yet.</p>
                    </div>
                )}

                {isValidating && (
                    <div className="flex flex-col items-center justify-center h-[300px] text-muted-foreground">
                        <Loader2 className="h-10 w-10 mb-2 animate-spin text-primary" />
                        <p>Validating configuration...</p>
                    </div>
                )}

                {validationResult && (
                    <div className={`flex flex-col items-center justify-center p-6 rounded-lg border ${validationResult.valid ? "bg-green-50 border-green-200 dark:bg-green-900/10 dark:border-green-900" : "bg-red-50 border-red-200 dark:bg-red-900/10 dark:border-red-900"}`}>
                        {validationResult.valid ? (
                            <CheckCircle2 className="h-16 w-16 text-green-500 mb-4" />
                        ) : (
                            <XCircle className="h-16 w-16 text-red-500 mb-4" />
                        )}
                        <h3 className={`text-xl font-bold mb-2 ${validationResult.valid ? "text-green-700 dark:text-green-400" : "text-red-700 dark:text-red-400"}`}>
                            {validationResult.valid ? "Valid Configuration" : "Validation Failed"}
                        </h3>
                        <p className={`text-center ${validationResult.valid ? "text-green-600 dark:text-green-300" : "text-red-600 dark:text-red-300"}`}>
                            {validationResult.message}
                        </p>
                    </div>
                )}

                {validationResult?.valid && (
                    <div className="mt-6">
                        <h4 className="font-semibold mb-2">Next Steps</h4>
                        <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
                            <li>You can use this configuration in your <code>config.yaml</code> file.</li>
                            <li>Or create a new service in the <strong>Upstream Services</strong> page.</li>
                        </ul>
                    </div>
                )}
            </CardContent>
        </Card>
      </div>
    </div>
  );
}
