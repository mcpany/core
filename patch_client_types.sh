#!/bin/bash
sed -i '/export const apiClient = {/i \
export interface HITLApproval {\
    id: string;\
    tool: string;\
    intent: string;\
    status: string;\
    requireMfa: boolean;\
}\
' ui/src/lib/client.ts

sed -i 's|getHITLApprovals: async (): Promise<any\[\]>|getHITLApprovals: async (): Promise<HITLApproval\[\]>|g' ui/src/lib/client.ts
