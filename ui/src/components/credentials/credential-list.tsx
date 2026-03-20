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
import { Plus, Trash, Key, Lock, Globe, ExternalLink, Trash2 } from "lucide-react"
import { useToast } from "@/hooks/use-toast"
import { Checkbox } from "@/components/ui/checkbox"

/**
 * CredentialList component.
 * @returns The rendered component.
 */
export function CredentialList() {
  const { toast } = useToast()
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [loading, setLoading] = useState(true)
  const [isOpen, setIsOpen] = useState(false)
  const [editingCred, setEditingCred] = useState<Credential | null>(null)
  const [selectedCredentials, setSelectedCredentials] = useState<Set<string>>(new Set())

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
      setSelectedCredentials(new Set())
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

  async function handleBulkDelete() {
      if (selectedCredentials.size === 0) return;
      if (!confirm(`Are you sure you want to delete ${selectedCredentials.size} credentials?`)) return;

      try {
          await Promise.all(Array.from(selectedCredentials).map(id => apiClient.deleteCredential(id)));
          toast({ description: `${selectedCredentials.size} credentials deleted` })
          loadCredentials()
      } catch (error) {
          toast({ variant: "destructive", description: "Failed to delete some credentials" })
          loadCredentials() // reload anyway to show what was successfully deleted
      }
  }

  const toggleSelectAll = () => {
      if (selectedCredentials.size === credentials.length) {
          setSelectedCredentials(new Set())
      } else {
          setSelectedCredentials(new Set(credentials.map(c => c.id)))
      }
  }

  const toggleSelect = (id: string) => {
      const next = new Set(selectedCredentials)
      if (next.has(id)) {
          next.delete(id)
      } else {
          next.add(id)
      }
      setSelectedCredentials(next)
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
        <div className="flex items-center gap-4">
            <h2 className="text-xl font-semibold">Credentials</h2>
            {selectedCredentials.size > 0 && (
                <Button variant="destructive" size="sm" onClick={handleBulkDelete}>
                    <Trash2 className="mr-2 h-4 w-4" /> Delete Selected ({selectedCredentials.size})
                </Button>
            )}
        </div>
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

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                  <Checkbox
                      checked={credentials.length > 0 && selectedCredentials.size === credentials.length}
                      onCheckedChange={toggleSelectAll}
                      aria-label="Select all"
                  />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Details</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
                <TableRow><TableCell colSpan={5} className="text-center py-4">Loading...</TableCell></TableRow>
            ) : credentials.length === 0 ? (
                <TableRow><TableCell colSpan={5} className="text-center py-4 text-muted-foreground">No credentials found</TableCell></TableRow>
            ) : (
                credentials.map((cred) => (
                    <TableRow key={cred.id}>
                        <TableCell>
                            <Checkbox
                                checked={selectedCredentials.has(cred.id)}
                                onCheckedChange={() => toggleSelect(cred.id)}
                                aria-label={`Select ${cred.name}`}
                            />
                        </TableCell>
                        <TableCell className="font-medium">
                            <div className="flex items-center gap-2">
                                <Key className="h-4 w-4 text-muted-foreground" />
                                {cred.name}
                            </div>
                        </TableCell>
                        <TableCell>
                            {cred.authentication?.apiKey ? "API Key" :
                             cred.authentication?.bearerToken ? "Bearer Token" :
                             cred.authentication?.basicAuth ? "Basic Auth" :
                             cred.authentication?.oauth2 ? "OAuth 2.0" : "Unknown"}
                        </TableCell>
                        <TableCell className="text-muted-foreground text-sm">
                            {cred.authentication?.apiKey && (
                                <span>{cred.authentication.apiKey.paramName} ({cred.authentication.apiKey.in === 0 ? "Header" : "Query"})</span>
                            )}
                            {cred.authentication?.bearerToken && <span>Bearer</span>}
                            {cred.authentication?.basicAuth && <span>{cred.authentication.basicAuth.username}</span>}
                        </TableCell>
                        <TableCell className="text-right flex items-center justify-end gap-1">
                             {cred.authentication?.oauth2 && (
                                 <Button variant="outline" size="sm" onClick={() => handleConnect(cred)}>
                                     <ExternalLink className="mr-2 h-3.5 w-3.5" />
                                     {cred.token?.accessToken ? "Reconnect" : "Authorize"}
                                 </Button>
                             )}
                             <Button variant="ghost" size="sm" onClick={() => handleEdit(cred)}>Edit</Button>
                             <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => handleDelete(cred.id)} aria-label="Delete"><Trash className="h-4 w-4" /></Button>
                        </TableCell>
                    </TableRow>
                ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
