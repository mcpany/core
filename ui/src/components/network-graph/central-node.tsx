/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Cpu } from 'lucide-react';
import { Card, CardHeader, CardTitle } from '@/components/ui/card';

const CentralNode = ({ data }: NodeProps) => {
  return (
    <Card className="w-[250px] shadow-xl border-primary/50 bg-primary/5 backdrop-blur-md">
       <Handle type="source" position={Position.Bottom} className="!bg-primary" />
      <CardHeader className="p-4 flex flex-row items-center gap-3 justify-center text-center">
        <div className="bg-primary/20 p-2 rounded-full">
             <Cpu className="size-8 text-primary" />
        </div>
        <div>
            <CardTitle className="text-lg font-bold text-primary">MCP Any</CardTitle>
            <div className="text-xs text-muted-foreground">Gateway Core</div>
        </div>
      </CardHeader>
    </Card>
  );
};

export default memo(CentralNode);
