"use client"

import ServiceForm from "@/components/service-form"

export default function NewServicePage() {
  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Add Service</h2>
        <p className="text-muted-foreground mt-2">
          Configure a new upstream MCP service.
        </p>
      </div>
      <ServiceForm />
    </div>
  )
}
