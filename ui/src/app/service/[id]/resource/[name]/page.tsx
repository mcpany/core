/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ResourceDetail } from "@/components/resource-detail";
import { Breadcrumbs, BreadcrumbItem } from "@/components/breadcrumbs";
import { useState, useEffect, use } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";

export default function ResourceDetailPage({ params: paramsPromise }: { params: Promise<{ id: string, name: string }> }) {
    const params = use(paramsPromise);
    const [service, setService] = useState<UpstreamServiceConfig | null>(null);

    useEffect(() => {
        apiClient.getService(params.id).then(res => setService(res));
    }, [params.id]);

    const breadcrumbItems: BreadcrumbItem[] = service ? [
        { label: service.name, href: `/service/${params.id}` },
        { label: decodeURIComponent(params.name), href: `/service/${params.id}/resource/${params.name}` }
    ] : [];

    return (
        <main className="flex min-h-screen flex-col items-center bg-background p-4 sm:p-8">
            <Breadcrumbs items={breadcrumbItems} className="max-w-4xl"/>
            <ResourceDetail serviceId={params.id} resourceName={params.name} />
        </main>
    );
}
