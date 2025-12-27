"use client"

import ServiceForm from "@/components/service-form"
import { useParams } from "next/navigation"

export default function EditServicePage() {
    const params = useParams()
    const id = params.id as string

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Edit Service</h2>
        <p className="text-muted-foreground mt-2">
          Update configuration for {id}.
        </p>
      </div>
      <ServiceForm serviceId={id} />
    </div>
  )
}
