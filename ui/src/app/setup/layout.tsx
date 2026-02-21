/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Setup Wizard - MCP Any",
  description: "Configure your first MCP Gateway service.",
};

export default function SetupLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-4rem)] bg-muted/20">
      {children}
    </div>
  );
}
