/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

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
import { Plus, Trash, Key, Lock, Globe } from "lucide-react"
import { useToast } from "@/hooks/use-toast"

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
      data.sort((a: Credential, b: Credential) => a.name.localeCompare(b.name))
      setCredentials(data)
    } catch (error) {

      console.error(error)
      toast({ variant: "destructive", description: "Failed to load credentials" })
    } finally {
      setLoading(false)
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

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Details</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
                <TableRow><TableCell colSpan={4} className="text-center py-4">Loading...</TableCell></TableRow>
            ) : credentials.length === 0 ? (
                <TableRow><TableCell colSpan={4} className="text-center py-4 text-muted-foreground">No credentials found</TableCell></TableRow>
            ) : (
                credentials.map((cred) => (
                    <TableRow key={cred.id}>
                        <TableCell className="font-medium flex items-center gap-2">
                            <Key className="h-4 w-4 text-muted-foreground" />
                            {cred.name}
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
                        <TableCell className="text-right">
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
