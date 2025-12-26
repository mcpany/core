/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import StackDetailClient from "./StackDetailClient";

export default async function StackDetailPage({ params }: { params: Promise<{ stackId: string }> }) {
  const { stackId } = await params;
  return <StackDetailClient stackId={stackId} />;
}
