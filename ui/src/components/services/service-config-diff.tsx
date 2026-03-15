/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import yaml from "js-yaml";
import { UpstreamServiceConfig } from "@/lib/types";
import { DiffViewer } from "@/components/services/editor/diff-viewer";

interface ServiceConfigDiffProps {
  original: UpstreamServiceConfig;
  modified: UpstreamServiceConfig;
}

/**
 * ServiceConfigDiff component.
 * @param props - The component props.
 * @param props.original - The original property.
 * @param props.modified - The modified property.
 * @returns The rendered component.
 */
export function ServiceConfigDiff({ original, modified }: ServiceConfigDiffProps) {
  // Dump to YAML
  // We use simple sorting to ensure keys are in consistent order for better diffs
  const originalYaml = yaml.dump(original, { sortKeys: true, indent: 2, lineWidth: -1 });
  const modifiedYaml = yaml.dump(modified, { sortKeys: true, indent: 2, lineWidth: -1 });

  return (
    <div className="h-[400px] w-full">
      <DiffViewer
        original={originalYaml}
        modified={modifiedYaml}
        language="yaml"
      />
    </div>
  );
}
