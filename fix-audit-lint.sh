#!/bin/bash
sed -i 's/const filters: any = {/const filters: Record<string, string | number> = {/' ui/src/components/audit/audit-log-viewer.tsx
sed -i 's/const filters: any = {};/const filters: Record<string, string | number> = {};/' ui/src/components/audit/audit-log-viewer.tsx
sed -i 's/catch (e: any) {/catch (e: unknown) {/' ui/src/components/audit/audit-log-viewer.tsx
