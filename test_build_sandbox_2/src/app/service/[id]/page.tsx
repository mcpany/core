/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ServiceDetail } from "@/components/service-detail";
import { Breadcrumbs, BreadcrumbItem } from "@/components/breadcrumbs";
import { useState, useEffect, use } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";
import { useServiceSiblings } from "@/hooks/use-siblings";

/**
 * Page component for displaying service details.
 *
 * @param props - The component props.
 * @param props.params - Promise resolving to route parameters containing the service ID.
 * @returns The service detail page.
 */
export default function ServiceDetailPage({ params: paramsPromise }: { params: Promise<{ id: string }> }) {
  const params = use(paramsPromise);
  const [service, setService] = useState<UpstreamServiceConfig | null>(null);
  const siblings = useServiceSiblings(params.id);

  useEffect(() => {
    apiClient.getService(params.id).then(res => setService(res.service || null));
  }, [params.id]);

  const breadcrumbItems: BreadcrumbItem[] = service ? [{
      label: service.name,
      href: `/service/${params.id}`,
      siblings: siblings
  }] : [];

  return (
    <main className="flex min-h-screen flex-col items-center bg-background p-4 sm:p-8">
      <Breadcrumbs items={breadcrumbItems} />
      <ServiceDetail serviceId={params.id} />
    </main>
  );
}
