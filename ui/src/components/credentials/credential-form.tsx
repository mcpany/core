/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Credential, Authentication, APIKeyAuth_Location } from "@proto/config/v1/auth"
import { apiClient } from "@/lib/client"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { toast } from "sonner"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

const formSchema = z.object({
  name: z.string().min(2, {
    message: "Name must be at least 2 characters.",
  }),
  authType: z.string(),
  // API Key
  apiKeyParamName: z.string().optional(),
  apiKeyLocation: z.string().optional(),
  apiKeyValue: z.string().optional(),
  // Bearer Token
  bearerToken: z.string().optional(),
  // Basic Auth
  basicUsername: z.string().optional(),
  basicPassword: z.string().optional(),
})

interface CredentialFormProps {
  initialData?: Credential | null
  onSuccess: (cred: Credential) => void
}

export function CredentialForm({ initialData, onSuccess }: CredentialFormProps) {
  const [isTesting, setIsTesting] = useState(false)
  const [testUrl, setTestUrl] = useState("")

  // Determine initial auth type
  let defaultAuthType = "api_key"
  if (initialData?.authentication?.bearerToken) defaultAuthType = "bearer_token"
  if (initialData?.authentication?.basicAuth) defaultAuthType = "basic_auth"
  if (initialData?.authentication?.oauth2) defaultAuthType = "oauth2"

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: initialData?.name || "",
      authType: defaultAuthType,
      apiKeyParamName: initialData?.authentication?.apiKey?.paramName || "Authorization",
      apiKeyLocation: initialData?.authentication?.apiKey?.in?.toString() || APIKeyAuth_Location.HEADER.toString(),
      apiKeyValue: initialData?.authentication?.apiKey?.value?.plainText || "",
      bearerToken: initialData?.authentication?.bearerToken?.token?.plainText || "",
      basicUsername: initialData?.authentication?.basicAuth?.username || "",
      basicPassword: initialData?.authentication?.basicAuth?.password?.plainText || "",
    },
  })

  // Watch authType to conditionally render fields
  const authType = form.watch("authType")

  async function onSubmit(values: z.infer<typeof formSchema>) {
    try {
        const auth: Authentication = {}
        if (values.authType === "api_key") {
            auth.apiKey = {
                paramName: values.apiKeyParamName || "X-API-Key",
                in: parseInt(values.apiKeyLocation || "0") as APIKeyAuth_Location,
                value: { plainText: values.apiKeyValue || "" },
                verificationValue: "" // Not used for client
            }
        } else if (values.authType === "bearer_token") {
            auth.bearerToken = {
                token: { plainText: values.bearerToken || "" }
            }
        } else if (values.authType === "basic_auth") {
            auth.basicAuth = {
                username: values.basicUsername || "",
                password: { plainText: values.basicPassword || "" },
                passwordHash: ""
            }
        }
        // TODO: OAuth2 (requires more fields)

        const payload: Credential = {
            id: initialData?.id || "",
            name: values.name,
            authentication: auth
        }

        let saved: Credential;
        if (initialData?.id) {
            saved = await apiClient.updateCredential(payload)
            toast.success("Credential updated")
        } else {
            saved = await apiClient.createCredential(payload)
            toast.success("Credential created")
        }
        onSuccess(saved)
    } catch (error: any) {
        toast.error("Failed to save credential: " + error.message)
    }
  }

  async function handleTest() {
      if (!testUrl) {
          toast.error("Please enter a URL to test")
          return
      }
      setIsTesting(true)
      try {
        // Construct temporary credential from form values to test without saving (or use saved if not changed?)
        // Better to use form values
        // Reuse logic from onSubmit
        const values = form.getValues()
        const auth: Authentication = {}
        if (values.authType === "api_key") {
            auth.apiKey = {
                paramName: values.apiKeyParamName || "X-API-Key",
                in: parseInt(values.apiKeyLocation || "0") as APIKeyAuth_Location,
                value: { plainText: values.apiKeyValue || "" },
                verificationValue: ""
            }
        } else if (values.authType === "bearer_token") {
             auth.bearerToken = { token: { plainText: values.bearerToken || "" } }
        } else if (values.authType === "basic_auth") {
            auth.basicAuth = { username: values.basicUsername || "", password: { plainText: values.basicPassword || "" }, passwordHash: "" }
        }

        const res = await apiClient.testAuth({
            authentication: auth,
            target_url: testUrl,
            method: "GET"
        })
        if (res.status >= 200 && res.status < 300) {
            toast.success(`Test passed: ${res.status} ${res.status_text}`)
        } else {
            toast.warning(`Test returned: ${res.status} ${res.status_text}`)
        }
        console.log("Test Response:", res)
      } catch (error: any) {
          toast.error("Test failed: " + error.message)
      } finally {
          setIsTesting(false)
      }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="My Credential" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
            control={form.control}
            name="authType"
            render={({ field }) => (
                <FormItem>
                    <FormLabel>Type</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                            <SelectTrigger>
                                <SelectValue placeholder="Select type" />
                            </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                            <SelectItem value="api_key">API Key</SelectItem>
                            <SelectItem value="bearer_token">Bearer Token</SelectItem>
                            <SelectItem value="basic_auth">Basic Auth</SelectItem>
                            <SelectItem value="oauth2" disabled>OAuth 2.0 (Client Credentials) - Coming Soon</SelectItem>
                        </SelectContent>
                    </Select>
                    <FormMessage />
                </FormItem>
            )}
        />

        {authType === "api_key" && (
            <div className="space-y-4 border p-4 rounded-md">
                <div className="grid grid-cols-2 gap-4">
                    <FormField
                        control={form.control}
                        name="apiKeyParamName"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Parameter Name</FormLabel>
                                <FormControl><Input placeholder="X-API-Key" {...field} /></FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name="apiKeyLocation"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Location</FormLabel>
                                <Select onValueChange={field.onChange} defaultValue={field.value?.toString()}>
                                    <FormControl>
                                        <SelectTrigger><SelectValue /></SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value={APIKeyAuth_Location.HEADER.toString()}>Header</SelectItem>
                                        <SelectItem value={APIKeyAuth_Location.QUERY.toString()}>Query</SelectItem>
                                        <SelectItem value={APIKeyAuth_Location.COOKIE.toString()}>Cookie</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                </div>
                <FormField
                    control={form.control}
                    name="apiKeyValue"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Value</FormLabel>
                            <FormControl><Input type="password" placeholder="...secret key..." {...field} /></FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
            </div>
        )}

        {authType === "bearer_token" && (
            <div className="space-y-4 border p-4 rounded-md">
                <FormField
                    control={form.control}
                    name="bearerToken"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Token</FormLabel>
                            <FormControl><Input type="password" placeholder="...bearer token..." {...field} /></FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
            </div>
        )}

        {authType === "basic_auth" && (
             <div className="space-y-4 border p-4 rounded-md">
                <FormField
                    control={form.control}
                    name="basicUsername"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Username</FormLabel>
                            <FormControl><Input placeholder="username" {...field} /></FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="basicPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Password</FormLabel>
                            <FormControl><Input type="password" placeholder="password" {...field} /></FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
             </div>
        )}

        <div className="space-y-2 pt-4 border-t">
            <FormLabel>Test Connection</FormLabel>
            <div className="flex gap-2">
                <Input
                    placeholder="https://api.example.com/test"
                    value={testUrl}
                    onChange={(e) => setTestUrl(e.target.value)}
                />
                <Button type="button" variant="outline" onClick={handleTest} disabled={isTesting}>
                    {isTesting ? "Testing..." : "Test"}
                </Button>
            </div>
        </div>

        <div className="flex justify-end gap-2 pt-4">
             <Button type="submit">Save</Button>
        </div>
      </form>
    </Form>
  )
}
