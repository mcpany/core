#!/bin/bash
sed -i '/interface HITLApproval/d' ui/src/components/hitl/hitl-dashboard.tsx
sed -i '/id: string;/d' ui/src/components/hitl/hitl-dashboard.tsx
sed -i '/tool: string;/d' ui/src/components/hitl/hitl-dashboard.tsx
sed -i '/intent: string;/d' ui/src/components/hitl/hitl-dashboard.tsx
sed -i '/status: string;/d' ui/src/components/hitl/hitl-dashboard.tsx
sed -i '/requireMfa: boolean;/d' ui/src/components/hitl/hitl-dashboard.tsx
sed -i 's|import { apiClient }|import { apiClient, HITLApproval }|g' ui/src/components/hitl/hitl-dashboard.tsx
sed -i 's|className="backdrop-blur-sm bg-background/50 flex flex-col justify-between"|data-testid="hitl-card" className="backdrop-blur-sm bg-background/50 flex flex-col justify-between"|g' ui/src/components/hitl/hitl-dashboard.tsx
