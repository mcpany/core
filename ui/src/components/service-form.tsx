"use client"

import * as React from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import * as z from "zod"
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
import { Switch } from "@/components/ui/switch"
import { apiClient } from "@/lib/client"
import { useToast } from "@/components/ui/use-toast"
import { useRouter } from "next/navigation"
import { GlassCard } from "@/components/ui-custom/glass-card"

// Simplified schema for the prototype. In real app, this should cover all fields of UpstreamServiceConfig
const serviceFormSchema = z.object({
  name: z.string().min(2, {
    message: "Name must be at least 2 characters.",
  }),
  type: z.enum(["http", "grpc", "command_line", "mcp"]),
  address: z.string().optional(),
  command: z.string().optional(),
  args: z.string().optional(), // Space separated for simplicity
  version: z.string().optional(),
  disable: z.boolean().default(false),
})

type ServiceFormValues = z.infer<typeof serviceFormSchema>

interface ServiceFormProps {
    serviceId?: string
}

export default function ServiceForm({ serviceId }: ServiceFormProps) {
  const { toast } = useToast()
  const router = useRouter()
  const [loading, setLoading] = React.useState(!!serviceId)

  const form = useForm<ServiceFormValues>({
    resolver: zodResolver(serviceFormSchema),
    defaultValues: {
      name: "",
      type: "http",
      address: "",
      command: "",
      args: "",
      version: "1.0.0",
      disable: false,
    },
  })

  React.useEffect(() => {
      if (serviceId) {
          // Fetch existing service data
          const load = async () => {
              try {
                  const data = await apiClient.listServices()
                  const services = data.services || data
                  const service = services.find((s: any) => s.id === serviceId || s.name === serviceId)
                  if (service) {
                      let type: any = "http"
                      let address = ""
                      let command = ""
                      let args = ""

                      if (service.http_service) {
                          type = "http"
                          address = service.http_service.address
                      } else if (service.grpc_service) {
                          type = "grpc"
                          address = service.grpc_service.address
                      } else if (service.command_line_service) {
                          type = "command_line"
                          command = service.command_line_service.command
                          args = service.command_line_service.args?.join(" ") || ""
                      }

                      form.reset({
                          name: service.name,
                          type,
                          address,
                          command,
                          args,
                          version: service.version,
                          disable: !!service.disable,
                      })
                  }
              } catch(e) {
                   toast({
                        title: "Error",
                        description: "Failed to load service",
                        variant: "destructive",
                    })
              } finally {
                  setLoading(false)
              }
          }
          load()
      }
  }, [serviceId, form, toast])

  async function onSubmit(data: ServiceFormValues) {
    try {
        const config: any = {
            name: data.name,
            version: data.version,
            disable: data.disable,
        }

        if (data.type === 'http') {
            config.http_service = { address: data.address }
        } else if (data.type === 'grpc') {
            config.grpc_service = { address: data.address }
        } else if (data.type === 'command_line') {
            config.command_line_service = {
                command: data.command,
                args: data.args ? data.args.split(" ") : []
            }
        }

        if (serviceId) {
            await apiClient.updateService(config)
             toast({ title: "Service updated" })
        } else {
            await apiClient.registerService(config)
             toast({ title: "Service created" })
        }
        router.push("/services")
        router.refresh()
    } catch (error) {
        toast({
            title: "Error",
            description: "Failed to save service",
            variant: "destructive",
        })
    }
  }

  const type = form.watch("type")

  if (loading) return <div>Loading...</div>

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <GlassCard className="p-6">
            <div className="grid gap-6">
                <div className="grid grid-cols-2 gap-4">
                    <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Service Name</FormLabel>
                        <FormControl>
                            <Input placeholder="my-service" {...field} />
                        </FormControl>
                        <FormDescription>
                            Unique identifier for the service.
                        </FormDescription>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                    <FormField
                    control={form.control}
                    name="version"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Version</FormLabel>
                        <FormControl>
                            <Input placeholder="1.0.0" {...field} />
                        </FormControl>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                </div>

                 <FormField
                    control={form.control}
                    name="type"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Service Type</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                            <FormControl>
                            <SelectTrigger>
                                <SelectValue placeholder="Select a type" />
                            </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                            <SelectItem value="http">HTTP Service</SelectItem>
                            <SelectItem value="grpc">gRPC Service</SelectItem>
                            <SelectItem value="command_line">Command Line (Stdio)</SelectItem>
                            <SelectItem value="mcp">MCP Service</SelectItem>
                            </SelectContent>
                        </Select>
                        <FormMessage />
                        </FormItem>
                    )}
                    />

                {type === 'http' || type === 'grpc' ? (
                     <FormField
                        control={form.control}
                        name="address"
                        render={({ field }) => (
                            <FormItem>
                            <FormLabel>Address</FormLabel>
                            <FormControl>
                                <Input placeholder={type === 'http' ? "http://localhost:8080" : "localhost:50051"} {...field} />
                            </FormControl>
                            <FormMessage />
                            </FormItem>
                        )}
                        />
                ) : null}

                 {type === 'command_line' ? (
                     <>
                        <FormField
                            control={form.control}
                            name="command"
                            render={({ field }) => (
                                <FormItem>
                                <FormLabel>Command</FormLabel>
                                <FormControl>
                                    <Input placeholder="npx" {...field} />
                                </FormControl>
                                <FormMessage />
                                </FormItem>
                            )}
                            />
                             <FormField
                            control={form.control}
                            name="args"
                            render={({ field }) => (
                                <FormItem>
                                <FormLabel>Arguments</FormLabel>
                                <FormControl>
                                    <Input placeholder="-y @modelcontextprotocol/server-filesystem ..." {...field} />
                                </FormControl>
                                <FormDescription>Space separated arguments</FormDescription>
                                <FormMessage />
                                </FormItem>
                            )}
                            />
                     </>
                ) : null}


                <FormField
                control={form.control}
                name="disable"
                render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                        <FormLabel className="text-base">Disable Service</FormLabel>
                        <FormDescription>
                        Temporarily disable this service without deleting the configuration.
                        </FormDescription>
                    </div>
                    <FormControl>
                        <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                        />
                    </FormControl>
                    </FormItem>
                )}
                />
            </div>
        </GlassCard>

        <div className="flex justify-end gap-4">
            <Button type="button" variant="ghost" onClick={() => router.back()}>Cancel</Button>
            <Button type="submit">Save Changes</Button>
        </div>
      </form>
    </Form>
  )
}
