/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import dynamic from "next/dynamic";

export const RequestVolumeChart = dynamic(
  () =>
    import("./request-volume-chart").then((mod) => mod.RequestVolumeChart),
  {
    loading: () => <div className="h-[300px] w-full animate-pulse bg-muted/20 rounded-xl" />,
    ssr: false,
  }
);
