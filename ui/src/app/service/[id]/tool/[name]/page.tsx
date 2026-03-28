/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ToolDetail } from "@/components/tool-detail";
import { Breadcrumbs, BreadcrumbItem } from "@/components/breadcrumbs";
import { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";
import { useServiceSiblings, useToolSiblings } from "@/hooks/use-siblings";

/**
 * Intent: Document ToolDetailPage
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
 * ToolDetailPage component.
 * @returns The rendered component.
 */
export default function ToolDetailPage() {
    const { id = "", name = "" } = useParams<{ id: string; name: string }>();
    const [service, setService] = useState<UpstreamServiceConfig | null>(null);
    const serviceSiblings = useServiceSiblings(id);
    const toolSiblings = useToolSiblings(id, name);

    useEffect(() => {
        if (id) apiClient.getService(id).then(res => setService(res.service || null));
    }, [id]);

    const breadcrumbItems: BreadcrumbItem[] = service ? [
        {
            label: service.name,
            href: `/service/${id}`,
            siblings: serviceSiblings
        },
        {
            label: decodeURIComponent(name),
            href: `/service/${id}/tool/${name}`,
            siblings: toolSiblings
        }
    ] : [];

  return (
    <main className="flex min-h-screen flex-col items-center bg-background p-4 sm:p-8">
        <Breadcrumbs items={breadcrumbItems} className="max-w-4xl"/>
        <ToolDetail serviceId={id} toolName={name} />
    </main>
  );
}
