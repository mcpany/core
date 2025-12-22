"use client"

import { useEffect, useState } from "react"
import { apiClient } from "@/lib/mock-client"
import { UpstreamServiceConfig } from "@/lib/types"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import { MoreHorizontal, Plus, Settings } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchServices()
  }, [])

  const fetchServices = async () => {
    setLoading(true)
    try {
      const { services } = await apiClient.listServices()
      setServices(services)
    } catch (error) {
      console.error("Failed to fetch services", error)
    } finally {
      setLoading(false)
    }
  }

  const handleToggleService = async (id: string, currentStatus: boolean) => {
    try {
      await apiClient.setServiceStatus(id, !currentStatus) // !currentStatus because toggle flips it. Wait, disable=true means DISABLED.
      // If currently disabled (true), we want to enable (false).
      // If currently enabled (false), we want to disable (true).
      // So passing !currentStatus where currentStatus is 'disable' field is correct?
      // Wait, let's check mock client. setServiceStatus(id, disabled: boolean).
      // If I pass true, it disables it.
      // If service.disable is true (disabled), switch is OFF. I want to turn it ON (enable).
      // So I should pass false.
      // If service.disable is false (enabled), switch is ON. I want to turn it OFF (disable).
      // So I should pass true.
      // So yes, I should pass !service.disable.

      // Update local state optimistically or re-fetch
      setServices(services.map(s => s.id === id ? { ...s, disable: !s.disable } : s))
    } catch (error) {
      console.error("Failed to toggle service", error)
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Services</h1>
          <p className="text-muted-foreground">Manage upstream MCP services.</p>
        </div>
        <Sheet>
            <SheetTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Add Service
                </Button>
            </SheetTrigger>
            <SheetContent className="w-[400px] sm:w-[540px]">
                <SheetHeader>
                    <SheetTitle>Add Service</SheetTitle>
                    <SheetDescription>
                        Configure a new upstream service connection.
                    </SheetDescription>
                </SheetHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">
                            Name
                        </Label>
                        <Input id="name" placeholder="my-service" className="col-span-3" />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="url" className="text-right">
                            URL
                        </Label>
                        <Input id="url" placeholder="https://api.example.com" className="col-span-3" />
                    </div>
                     <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="type" className="text-right">
                            Type
                        </Label>
                        <Input id="type" placeholder="HTTP / gRPC" className="col-span-3" />
                    </div>
                </div>
                <div className="flex justify-end">
                     <Button type="submit">Save changes</Button>
                </div>
            </SheetContent>
        </Sheet>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Tools</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Enabled</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {services.map((service) => (
              <TableRow key={service.id}>
                <TableCell className="font-medium">
                   <div className="flex flex-col">
                        <span>{service.name}</span>
                        <span className="text-xs text-muted-foreground">{service.version}</span>
                    </div>
                </TableCell>
                <TableCell>
                    <Badge variant="outline">
                        {service.grpc_service ? "gRPC" :
                         service.http_service ? "HTTP" :
                         service.command_line_service ? "CMD" : "Other"}
                    </Badge>
                </TableCell>
                <TableCell>
                    {(service.grpc_service?.tools?.length || service.http_service?.tools?.length || service.command_line_service?.tools?.length || 0)}
                </TableCell>
                <TableCell>
                     {service.disable ? (
                        <span className="text-muted-foreground">Inactive</span>
                    ) : (
                        <span className="text-green-600 font-medium">Active</span>
                    )}
                </TableCell>
                <TableCell>
                  <Switch
                    checked={!service.disable}
                    onCheckedChange={(checked) => handleToggleService(service.id!, !checked)}
                  />
                </TableCell>
                <TableCell className="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" className="h-8 w-8 p-0">
                        <span className="sr-only">Open menu</span>
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Actions</DropdownMenuLabel>
                      <DropdownMenuItem onClick={() => navigator.clipboard.writeText(service.id!)}>
                        Copy ID
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem>View Details</DropdownMenuItem>
                      <DropdownMenuItem>Edit Configuration</DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-red-600">Delete Service</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))}
             {loading && (
                 <TableRow>
                     <TableCell colSpan={6} className="text-center h-24">
                        Loading services...
                     </TableCell>
                 </TableRow>
            )}
            {!loading && services.length === 0 && (
                <TableRow>
                    <TableCell colSpan={6} className="text-center h-24">
                        No services found.
                    </TableCell>
                </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
