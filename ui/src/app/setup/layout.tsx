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
    <div className="bg-background min-h-full">
      {/* We could hide sidebar here if we were using a different root layout */}
      {children}
    </div>
  );
}
