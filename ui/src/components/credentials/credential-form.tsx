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
import { useToast } from "@/hooks/use-toast"
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
  // OAuth 2.0
  oauthClientId: z.string().optional(),
  oauthClientSecret: z.string().optional(),
  oauthAuthUrl: z.string().optional(),
  oauthTokenUrl: z.string().optional(),
  oauthScopes: z.string().optional(),
})

interface CredentialFormProps {
  initialData?: Credential | null
  onSuccess: (cred: Credential) => void
}

/**
 * CredentialForm.
 *
 * @param onSuccess - The onSuccess.
 */
export function CredentialForm({ initialData, onSuccess }: CredentialFormProps) {
  const { toast } = useToast()
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
      oauthClientId: initialData?.authentication?.oauth2?.clientId?.plainText || "",
      oauthClientSecret: initialData?.authentication?.oauth2?.clientSecret?.plainText || "",
      oauthAuthUrl: initialData?.authentication?.oauth2?.authorizationUrl || "",
      oauthTokenUrl: initialData?.authentication?.oauth2?.tokenUrl || "",
      oauthScopes: initialData?.authentication?.oauth2?.scopes || "",
    },
  })

  // Watch authType to conditionally render fields
  const authType = form.watch("authType")

  async function onSubmit(values: z.infer<typeof formSchema>) {
    try {
        const auth: any = {}
        if (values.authType === "api_key") {
            auth.api_key = {
                param_name: values.apiKeyParamName || "X-API-Key",
                in: parseInt(values.apiKeyLocation || "0"),
                value: { plain_text: values.apiKeyValue || "" },
                verification_value: ""
            }
        } else if (values.authType === "bearer_token") {
            auth.bearer_token = {
                token: { plain_text: values.bearerToken || "" }
            }
        } else if (values.authType === "basic_auth") {
            auth.basic_auth = {
                username: values.basicUsername || "",
                password: { plain_text: values.basicPassword || "" },
                password_hash: ""
            }
        } else if (values.authType === "oauth2") {
            auth.oauth2 = {
                client_id: { plain_text: values.oauthClientId || "" },
                client_secret: { plain_text: values.oauthClientSecret || "" },
                authorization_url: values.oauthAuthUrl || "",
                token_url: values.oauthTokenUrl || "",
                scopes: values.oauthScopes || "",
                issuer_url: "",
                audience: "",
            }
        }

        const payload: any = {
            id: initialData?.id || "",
            name: values.name,
            authentication: auth,
            token: initialData?.token
        }

        let saved: Credential;
        if (initialData?.id) {
            saved = await apiClient.updateCredential(payload)
            toast({ description: "Credential updated" })
        } else {
            saved = await apiClient.createCredential(payload)
            toast({ description: "Credential created" })
        }
        onSuccess(saved)
    } catch (error: any) {

        toast({ variant: "destructive", description: "Failed to save credential: " + error.message })
    }
  }

  async function handleTest() {
      if (!testUrl) {
      if (!testUrl) {
          toast({ variant: "destructive", description: "Please enter a URL to test" })
          return
      }
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
        } else if (values.authType === "oauth2") {
             // For OAuth2, test connection might mean using the stored token?
             // Or verifying the config?
             // If we have a token in initialData, we can use it?
             // But valid test requires a real request.
             // If we don't have a token, we can't really test "connection" to a protected resource yet.
             // Maybe warn user?
             if (initialData?.token?.accessToken) {
                 // Use the stored token (simulated as Bearer?)
                 // `testAuth` endpoint likely expects `Authentication` loaded.
                 // If we send `auth.oauth2` config, `testAuth` might try to use it?
                 // But `testAuth` probably doesn't handle OAuth flow.
                 // It checks if `Authenticator` works.
                 // OAuth2 Authenticator uses `UserToken`?
                 // If `testAuth` allows passing a UserToken, we could test.
                 // But `Authentication` proto doesn't hold the token. `Credential` does.
                 // `testAuth` API handler takes `TestAuthRequest` which likely contains `Authentication`.
                 // It doesn't seem to take `UserToken`.
                 // So `testAuth` for OAuth might be tricky without token.
                 // Let's defer OAuth test logic or just basic config check.
                 auth.oauth2 = {
                    clientId: { plainText: values.oauthClientId || "" },
                    clientSecret: { plainText: values.oauthClientSecret || "" },
                    authorizationUrl: values.oauthAuthUrl || "",
                    tokenUrl: values.oauthTokenUrl || "",
                    scopes: values.oauthScopes || "",
                    issuerUrl: "",
                    audience: "",
                 }
             } else {
                 toast({ description: "OAuth 2.0 requires connecting (getting a token) before testing." })
                 return
             }
        }

        const res = await apiClient.testAuth({
            authentication: auth, // This might not be enough for OAuth if logic requires Token
            target_url: testUrl,
            method: "GET"
        })
        if (res.status >= 200 && res.status < 300) {
            toast({ description: `Test passed: ${res.status} ${res.status_text}` })
        } else {
            toast({ variant: "destructive", description: `Test returned: ${res.status} ${res.status_text}` })
        }

      } catch (error: any) {
          toast({ variant: "destructive", description: "Test failed: " + error.message })
      } finally {
          setIsTesting(false)
      }
  }

  async function handleConnect() {
      if (!initialData?.id) {
          toast({ variant: "destructive", description: "Please save the credential first." })
          return
      }
      try {
          // Check if we have necessary config saved?
          // We assume user saved what is in the form.
          // Initiate OAuth
          const redirectUrl = `${window.location.origin}/auth/callback`
          // Store credential ID in session for callback to use
          sessionStorage.setItem('oauth_credential_id', initialData.id)
          sessionStorage.setItem('oauth_redirect_url', redirectUrl)
          sessionStorage.setItem('oauth_return_path', '/credentials')

          const res = await apiClient.initiateOAuth("", redirectUrl, initialData.id)
          if (res.authorization_url) {
              // Store state if returned (it should be!)
              if (res.state) {
                  sessionStorage.setItem('oauth_state', res.state)
              }
              window.location.href = res.authorization_url
          } else {
              toast({ variant: "destructive", description: "Failed to get authorization URL" })
          }
      } catch (e: any) {
          toast({ variant: "destructive", description: "Failed to initiate connection: " + e.message })
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
                            <SelectItem value="oauth2">OAuth 2.0</SelectItem>
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

        {authType === "oauth2" && (
            <div className="space-y-4 border p-4 rounded-md">
                <FormField
                    control={form.control}
                    name="oauthClientId"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Client ID</FormLabel>
                            <FormControl><Input placeholder="...client id..." {...field} /></FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="oauthClientSecret"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Client Secret</FormLabel>
                            <FormControl><Input type="password" placeholder="...client secret..." {...field} /></FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <div className="grid grid-cols-2 gap-4">
                    <FormField
                        control={form.control}
                        name="oauthAuthUrl"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Auth URL</FormLabel>
                                <FormControl><Input placeholder="https://..." {...field} /></FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name="oauthTokenUrl"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Token URL</FormLabel>
                                <FormControl><Input placeholder="https://..." {...field} /></FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                </div>
                <FormField
                    control={form.control}
                    name="oauthScopes"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Scopes</FormLabel>
                            <FormControl><Input placeholder="scope1 scope2" {...field} /></FormControl>
                            <FormDescription>Space separated scopes</FormDescription>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {initialData?.id && (
                    <div className="pt-2">
                         <div className="flex items-center gap-4">
                             <Button type="button" variant="secondary" onClick={handleConnect}>
                                 {initialData.token?.accessToken ? "Reconnect" : "Connect Account"}
                             </Button>
                             {initialData.token?.accessToken && <span className="text-green-600 text-sm">Connected</span>}
                         </div>
                    </div>
                )}
                 {!initialData?.id && (
                     <div className="pt-2">
                        <span className="text-muted-foreground text-sm">Save credential to connect.</span>
                     </div>
                 )}
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
