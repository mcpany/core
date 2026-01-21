/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { Component, ErrorInfo, ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { AlertCircle, RefreshCcw, FileText } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface Props {
  children?: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

/**
 * Global Error Boundary to catch and report UI crashes.
 */
export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Uncaught error:", error, errorInfo);
  }

  private handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
          return this.props.fallback;
      }

      return (
        <div className="flex flex-col items-center justify-center min-h-[400px] p-6 text-center space-y-4">
          <div className="bg-destructive/10 p-4 rounded-full text-destructive mb-2">
            <AlertCircle className="h-12 w-12" />
          </div>
          <h2 className="text-2xl font-bold tracking-tight">Something went wrong</h2>
          <p className="text-muted-foreground max-w-md mx-auto">
            The application encountered an unexpected error. We've been notified and are working on a fix.
          </p>

          <Alert variant="destructive" className="max-w-xl text-left font-mono text-xs overflow-hidden">
            <FileText className="h-4 w-4" />
            <AlertTitle>Error Details</AlertTitle>
            <AlertDescription className="mt-2 whitespace-pre-wrap line-clamp-4">
                {this.state.error?.message || "Unknown error occurred"}
            </AlertDescription>
          </Alert>

          <div className="flex gap-4 pt-4">
            <Button variant="default" onClick={this.handleReset}>
              <RefreshCcw className="mr-2 h-4 w-4" />
              Reload Page
            </Button>
            <Button variant="outline" onClick={() => window.location.href = "/"}>
              Go to Home
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
