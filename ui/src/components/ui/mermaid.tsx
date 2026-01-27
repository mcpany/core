/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useRef, useState } from "react";
import mermaid from "mermaid";
import { useTheme } from "next-themes";
import { Loader2 } from "lucide-react";

interface MermaidProps {
  chart: string;
}

export function Mermaid({ chart }: MermaidProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [rendered, setRendered] = useState(false);
  const { theme, systemTheme } = useTheme();

  useEffect(() => {
    const currentTheme = theme === "system" ? systemTheme : theme;
    mermaid.initialize({
      startOnLoad: false,
      theme: currentTheme === "dark" ? "dark" : "default",
      securityLevel: "strict",
      fontFamily: "var(--font-inter)",
      sequence: {
        actorFontSize: 14,
        noteFontSize: 14,
        messageFontSize: 14,
        mirrorActors: false,
        useMaxWidth: false,
      }
    });
  }, [theme, systemTheme]);

  useEffect(() => {
    const renderChart = async () => {
      setRendered(false);
      if (ref.current && chart) {
        try {
          // Unique ID for each render to prevent conflicts
          const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`;
          const { svg } = await mermaid.render(id, chart);
          if (ref.current) {
            ref.current.innerHTML = svg;
            setRendered(true);
          }
        } catch (error) {
          console.error("Mermaid render error:", error);
          if (ref.current) {
             // In case of syntax error, showing it might be helpful for debugging
            ref.current.innerText = "Failed to render diagram.";
            setRendered(true);
          }
        }
      }
    };

    renderChart();
  }, [chart, theme, systemTheme]);

  return (
    <div className="w-full flex justify-center overflow-auto p-4 bg-white/50 dark:bg-black/20 rounded-lg">
      {!rendered && <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />}
      <div ref={ref} className={rendered ? "w-full flex justify-center" : "hidden"} />
    </div>
  );
}
