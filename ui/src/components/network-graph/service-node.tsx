/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Server, Globe, Database, Box } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const ServiceNode = ({ data }: NodeProps) => {
  const isHttp = data.type === 'HTTP';
  const isGrpc = data.type === 'gRPC';
  const isActive = data.status === 'active';

  return (
    <Card className={`w-[250px] shadow-lg border-l-4 ${isActive ? 'border-l-blue-500' : 'border-l-gray-300'} bg-background/80 backdrop-blur-md`}>
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground" />
      <CardHeader className="p-3 pb-0">
        <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-semibold truncate flex items-center gap-2">
                {isHttp ? <Globe className="size-4 text-blue-500" /> :
                 isGrpc ? <Server className="size-4 text-green-500" /> :
                 <Box className="size-4 text-orange-500" />}
                {data.label as string}
            </CardTitle>
            <Badge variant={isActive ? "default" : "secondary"} className="text-[10px] px-1 h-5">
                {data.version as string}
            </Badge>
        </div>
      </CardHeader>
      <CardContent className="p-3 pt-2 text-xs text-muted-foreground">
        <div>Type: {data.type as string}</div>
        <div>Status: {isActive ? "Online" : "Disabled"}</div>
      </CardContent>
    </Card>
  );
};

export default memo(ServiceNode);
