/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ServiceDetail } from "@/components/service-detail";
import { Breadcrumbs, BreadcrumbItem } from "@/components/breadcrumbs";
import { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";
import { useServiceSiblings } from "@/hooks/use-siblings";

/**
 * Intent: Document ServiceDetailPage
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
 * Page component for displaying service details.
 * @returns The service detail page.
 */
export default function ServiceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [service, setService] = useState<UpstreamServiceConfig | null>(null);
  const siblings = useServiceSiblings(id ?? "");

  useEffect(() => {
    if (id) apiClient.getService(id).then(res => setService(res.service || null));
  }, [id]);

  const breadcrumbItems: BreadcrumbItem[] = service ? [{
      label: service.name,
      href: `/service/${id}`,
      siblings: siblings
  }] : [];

  return (
    <main className="flex min-h-screen flex-col items-center bg-background p-4 sm:p-8">
      <Breadcrumbs items={breadcrumbItems} />
      <ServiceDetail serviceId={id ?? ""} />
    </main>
  );
}
