/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { PromptDetail } from "@/components/prompt-detail";
import { Breadcrumbs, BreadcrumbItem } from "@/components/breadcrumbs";
import { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";
import { useServiceSiblings } from "@/hooks/use-siblings";

/**
 * PromptDetailPage component.
 * @returns The rendered component.
 */
export default function PromptDetailPage() {
    const { id = "", name = "" } = useParams<{ id: string; name: string }>();
    const [service, setService] = useState<UpstreamServiceConfig | null>(null);
    const serviceSiblings = useServiceSiblings(id);

    useEffect(() => {
        if (id) apiClient.getService(id).then(res => setService(res.service || null));
    }, [id]);

    const breadcrumbItems: BreadcrumbItem[] = service ? [
        { label: service.name, href: `/service/${id}`, siblings: serviceSiblings },
        { label: decodeURIComponent(name), href: `/service/${id}/prompt/${name}` }
    ] : [];

    return (
        <main className="flex min-h-screen flex-col items-center bg-background p-4 sm:p-8">
            <Breadcrumbs items={breadcrumbItems} className="max-w-4xl"/>
            <PromptDetail serviceId={id} promptName={name} />
        </main>
    );
}
