import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";

export function useServiceSiblings(currentServiceId: string) {
    const [siblings, setSiblings] = useState<{ label: string; href: string }[]>([]);

    useEffect(() => {
        apiClient.listServices().then(services => {
            const list = Array.isArray(services) ? services : [];
            setSiblings(list
                .filter((s: any) => s.id !== currentServiceId)
                .map((s: any) => ({ label: s.name, href: `/service/${s.id}` }))
            );
        });
    }, [currentServiceId]);

    return siblings;
}

export function useToolSiblings(serviceId: string, currentToolName: string) {
    const [siblings, setSiblings] = useState<{ label: string; href: string }[]>([]);

    useEffect(() => {
        apiClient.listTools().then(res => {
            const tools = res.tools || [];
            const decodedName = decodeURIComponent(currentToolName);
            setSiblings(tools
                .filter((t: any) => t.service_id === serviceId && t.name !== decodedName)
                .map((t: any) => ({ label: t.name, href: `/service/${serviceId}/tool/${t.name}` }))
            );
        });
    }, [serviceId, currentToolName]);

    return siblings;
}
