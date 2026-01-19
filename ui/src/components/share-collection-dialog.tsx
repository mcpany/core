/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { Check, Copy, ExternalLink, Share2 } from "lucide-react"
import * as jsyaml from "js-yaml"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Checkbox } from "@/components/ui/checkbox"
import { Textarea } from "@/components/ui/textarea"
import { apiClient, UpstreamServiceConfig } from "@/lib/client"
import { useToast } from "@/hooks/use-toast"
import { ScrollArea } from "@/components/ui/scroll-area"

export function ShareCollectionDialog() {
  const [open, setOpen] = React.useState(false)
  const [services, setServices] = React.useState<UpstreamServiceConfig[]>([])
  const [selected, setSelected] = React.useState<Set<string>>(new Set())
  const [generatedConfig, setGeneratedConfig] = React.useState("")
  const [loading, setLoading] = React.useState(false)
  const { toast } = useToast()

  React.useEffect(() => {
    if (open) {
      setLoading(true)
      apiClient.listServices()
        .then((data) => {
             const list = Array.isArray(data) ? data : (data.services || []);
             setServices(list);
             // Default Select All? Or None? Let's say None.
        })
        .catch((err) => {
            console.error("Failed to list services", err)
            toast({
                title: "Error fetching services",
                description: "Could not load current services.",
                variant: "destructive"
            })
        })
        .finally(() => setLoading(false))
    }
  }, [open, toast])

  const toggleSelect = (name: string) => {
    const newSelected = new Set(selected)
    if (newSelected.has(name)) {
      newSelected.delete(name)
    } else {
      newSelected.add(name)
    }
    setSelected(newSelected)
  }

  const toggleSelectAll = () => {
    if (selected.size === services.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(services.map(s => s.name)))
    }
  }

  const generateConfig = () => {
    const selectedServices = services.filter(s => selected.has(s.name))
    // Clean up for export - remove IDs if they are system generated?
    // Usually keep basic config.
    // Helper to sanitize
    const sanitized = selectedServices.map(s => {
        // Create a clean copy conformant to UpstreamServiceConfig for export
        // We might want to remove 'connectionPool' status etc.
        const { connectionPool, id, ...rest } = s as any;
        return rest;
    })

    const collection = {
        name: "My Shared Collection",
        description: "Exported from My MCP Any Instance",
        services: sanitized
    }

    try {
        const yamlStr = jsyaml.dump(collection)
        setGeneratedConfig(yamlStr)
    } catch(e) {
        console.error("YAML dump failed", e)
        setGeneratedConfig(JSON.stringify(collection, null, 2))
    }
  }

  const copyToClipboard = () => {
    navigator.clipboard.writeText(generatedConfig)
    toast({
      title: "Copied!",
      description: "Collection configuration copied to clipboard.",
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="gap-2">
          <Share2 className="h-4 w-4" />
          Share Your Config
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Share Service Collection</DialogTitle>
          <DialogDescription>
            Select the services you want to include in your shared collection.
          </DialogDescription>
        </DialogHeader>

        {!generatedConfig ? (
            <div className="grid gap-4 py-4">
            <div className="rounded-md border max-h-[300px] overflow-auto">
                <Table>
                <TableHeader>
                    <TableRow>
                    <TableHead className="w-[50px]">
                        <Checkbox
                            checked={services.length > 0 && selected.size === services.length}
                            onCheckedChange={toggleSelectAll}
                        />
                    </TableHead>
                    <TableHead>Service Name</TableHead>
                    <TableHead>Type</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {loading ? (
                        <TableRow>
                            <TableCell colSpan={3} className="text-center">Loading...</TableCell>
                        </TableRow>
                    ) : services.length === 0 ? (
                        <TableRow>
                            <TableCell colSpan={3} className="text-center">No services found.</TableCell>
                        </TableRow>
                    ) : (
                        services.map((service) => (
                        <TableRow key={service.name}>
                            <TableCell>
                            <Checkbox
                                checked={selected.has(service.name)}
                                onCheckedChange={() => toggleSelect(service.name)}
                            />
                            </TableCell>
                            <TableCell className="font-medium">{service.name}</TableCell>
                            <TableCell>
                                {service.commandLineService ? "Command" :
                                 service.httpService ? "HTTP" :
                                 service.grpcService ? "gRPC" :
                                 service.mcpService ? "MCP" : "Unknown"}
                            </TableCell>
                        </TableRow>
                        ))
                    )}
                </TableBody>
                </Table>
            </div>
            <DialogFooter>
                <Button onClick={generateConfig} disabled={selected.size === 0}>
                    Generate Configuration
                </Button>
            </DialogFooter>
            </div>
        ) : (
            <div className="grid gap-4 py-4">
                <div className="relative">
                    <Textarea
                        value={generatedConfig}
                        readOnly
                        className="min-h-[300px] font-mono text-xs"
                    />
                    <Button
                        size="icon"
                        variant="ghost"
                        className="absolute right-2 top-2 h-8 w-8 bg-muted/50 hover:bg-muted"
                        onClick={copyToClipboard}
                    >
                        <Copy className="h-4 w-4" />
                    </Button>
                </div>
                <DialogFooter>
                    <Button variant="ghost" onClick={() => setGeneratedConfig("")}>
                        Back to Selection
                    </Button>
                    <Button onClick={() => setOpen(false)}>
                        Done
                    </Button>
                </DialogFooter>
            </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
