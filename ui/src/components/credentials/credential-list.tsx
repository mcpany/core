/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
import { useEffect, useState } from "react"
import { apiClient } from "@/lib/client"
import { Credential } from "@proto/config/v1/auth"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"
import { CredentialForm } from "./credential-form"
import { Plus, Trash, Key, Lock, Globe, ExternalLink, KeyRound } from "lucide-react"
import { useToast } from "@/hooks/use-toast"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

/**
 * Intent: Document CredentialList
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * CredentialList component.
 * @returns The rendered component.
 */
export function CredentialList() {
  const { toast } = useToast()
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [loading, setLoading] = useState(true)
  const [isOpen, setIsOpen] = useState(false)
  const [editingCred, setEditingCred] = useState<Credential | null>(null)

  useEffect(() => {
    loadCredentials()
  }, [])

  async function loadCredentials() {
    setLoading(true)
    try {
      const data = await apiClient.listCredentials()
      // Sort by Name
      if (Array.isArray(data)) {
        data.sort((a: Credential, b: Credential) => a.name.localeCompare(b.name))
        setCredentials(data)
      } else {
        setCredentials([])
      }
    } catch (error) {

      console.error(error)
      toast({ variant: "destructive", description: "Failed to load credentials" })
    } finally {
      setLoading(false)
    }
  }

  async function handleConnect(cred: Credential) {
      try {
          const redirectUrl = `${window.location.origin}/oauth/callback`
          const res = await apiClient.initiateOAuth("", redirectUrl, cred.id)
          if (res.authorization_url) {
              // Store context for callback using unified JSON pattern
              sessionStorage.setItem(`oauth_pending_${res.state}`, JSON.stringify({
                  serviceId: '',
                  credentialId: cred.id,
                  state: res.state,
                  redirectUrl: redirectUrl,
                  returnPath: window.location.pathname + window.location.search
              }))

              window.location.href = res.authorization_url
          }
      } catch (e: any) {
          toast({ variant: "destructive", description: "Failed to initiate connection: " + e.message })
      }
  }

  async function handleDelete(id: string) {
      if (!confirm("Are you sure you want to delete this credential?")) return;
      try {
          await apiClient.deleteCredential(id)
          toast({ description: "Credential deleted" })
          loadCredentials()
      } catch (error) {
          toast({ variant: "destructive", description: "Failed to delete credential" })
      }
  }

  function handleEdit(cred: Credential) {
      setEditingCred(cred)
      setIsOpen(true)
  }

  function handleCreate() {
      setEditingCred(null)
      setIsOpen(true)
  }

  function onFormSuccess() {
      setIsOpen(false)
      loadCredentials()
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Credentials</h2>
        <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
                <Button onClick={handleCreate}><Plus className="mr-2 h-4 w-4" /> New Credential</Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>{editingCred ? "Edit Credential" : "Create Credential"}</DialogTitle>
                    <DialogDescription>
                        Configure authentication parameters for upstream services.
                    </DialogDescription>
                </DialogHeader>
                <CredentialForm initialData={editingCred} onSuccess={onFormSuccess} />
            </DialogContent>
        </Dialog>
      </div>

      <Card className="border shadow-sm">
        <CardContent className="p-0">
          <Table>
            <TableHeader className="bg-muted/50">
              <TableRow>
                <TableHead className="w-[30%]">Name</TableHead>
                <TableHead className="w-[20%]">Type</TableHead>
                <TableHead className="w-[30%]">Details</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                  <TableRow><TableCell colSpan={4} className="text-center h-24 text-muted-foreground">Loading credentials...</TableCell></TableRow>
              ) : credentials.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="h-64 text-center">
                      <div className="flex flex-col items-center justify-center space-y-3">
                        <div className="rounded-full bg-muted p-4">
                          <KeyRound className="h-8 w-8 text-muted-foreground" />
                        </div>
                        <div className="space-y-1">
                          <h3 className="text-lg font-medium">No credentials found</h3>
                          <p className="text-sm text-muted-foreground max-w-sm mx-auto">
                            Add credentials to authenticate your MCP Any instance with external services.
                          </p>
                        </div>
                        <Button onClick={handleCreate} variant="outline" className="mt-4">
                          <Plus className="mr-2 h-4 w-4" /> Create First Credential
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
              ) : (
                  credentials.map((cred) => (
                      <TableRow key={cred.id} className="group hover:bg-muted/50 transition-colors">
                          <TableCell className="font-medium">
                              <div className="flex items-center gap-3">
                                  <div className="rounded-md bg-primary/10 p-2">
                                      <Key className="h-4 w-4 text-primary" />
                                  </div>
                                  <span className="truncate">{cred.name}</span>
                              </div>
                          </TableCell>
                          <TableCell>
                              <Badge variant="secondary" className="font-normal">
                                {(cred.authentication?.apiKey || (cred.authentication as any)?.api_key) ? "API Key" :
                                (cred.authentication?.bearerToken || (cred.authentication as any)?.bearer_token) ? "Bearer Token" :
                                (cred.authentication?.basicAuth || (cred.authentication as any)?.basic_auth) ? "Basic Auth" :
                                cred.authentication?.oauth2 ? "OAuth 2.0" : "Unknown"}
                              </Badge>
                          </TableCell>
                          <TableCell className="text-muted-foreground text-sm">
                              <div className="flex items-center gap-2">
                                {(cred.authentication?.apiKey || (cred.authentication as any)?.api_key) && (
                                    <>
                                        <code className="bg-muted px-1.5 py-0.5 rounded text-xs">
                                            {cred.authentication.apiKey?.paramName || (cred.authentication as any)?.api_key?.param_name}
                                        </code>
                                        <span className="text-xs opacity-70">
                                            ({(cred.authentication.apiKey?.in || (cred.authentication as any)?.api_key?.in) === 0 ? "Header" : "Query"})
                                        </span>
                                    </>
                                )}
                                {(cred.authentication?.bearerToken || (cred.authentication as any)?.bearer_token) && <span>Bearer Auth</span>}
                                {(cred.authentication?.basicAuth || (cred.authentication as any)?.basic_auth) && (
                                    <code className="bg-muted px-1.5 py-0.5 rounded text-xs">
                                        {cred.authentication.basicAuth?.username || (cred.authentication as any)?.basic_auth?.username}
                                    </code>
                                )}
                                {cred.authentication?.oauth2 && (
                                    <Badge variant={(cred.token?.accessToken || (cred.token as any)?.access_token) ? "default" : "outline"} className="text-[10px] uppercase tracking-wider h-5">
                                        {(cred.token?.accessToken || (cred.token as any)?.access_token) ? "Authorized" : "Not Authorized"}
                                    </Badge>
                                )}
                              </div>
                          </TableCell>
                          <TableCell className="text-right">
                               <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                   {cred.authentication?.oauth2 && (
                                       <Button variant="outline" size="sm" onClick={() => handleConnect(cred)}>
                                           <ExternalLink className="mr-2 h-3 w-3" />
                                           {(cred.token?.accessToken || (cred.token as any)?.access_token) ? "Reconnect" : "Authorize"}
                                       </Button>
                                   )}
                                   <Button variant="ghost" size="sm" onClick={() => handleEdit(cred)}>Edit</Button>
                                   <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" onClick={() => handleDelete(cred.id)} aria-label="Delete">
                                       <Trash className="h-4 w-4" />
                                   </Button>
                               </div>
                          </TableCell>
                      </TableRow>
                  ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
