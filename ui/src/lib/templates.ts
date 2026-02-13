/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/types";
import { Database, FileText, Github, Globe, Server, Activity, Cloud, MessageSquare, Map, Clock, Zap, CheckCircle2 } from "lucide-react";

/**
 * A template for creating a new service configuration.
 */
export interface ServiceTemplate {
  /** Unique identifier for the template. */
  id: string;
  /** Display name of the template. */
  name: string;
  /** Description of what the template provides. */
  description: string;
  /** Icon component for the template. */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon: any; // Lucide icon component or string identifier
  /** Partial configuration provided by the template. */
  config: Partial<UpstreamServiceConfig>;
  /** The category of the service. */
  category?: string;
  /** Whether the service is featured. */
  featured?: boolean;
  /**
   * Optional list of fields that need to be filled in by the user.
   */
  fields?: {
    /** The name of the field (internal identifier). */
    name: string;
    /** The label to display for the field. */
    label: string;
    /** Placeholder text for the input. */
    placeholder: string;
    /** Key path in the config object where the value should be set (e.g. "httpService.address"). */
    key: string;
    /**
     * If set, the value will not replace the entire key content but will substitute this token.
     * Useful for command line arguments.
     */
    replaceToken?: string;
    /** Default value for the field. */
    defaultValue?: string;
    /** Input type (text, password, etc). Defaults to text. */
    type?: string;
  }[];
  /** Tags for the service */
  tags?: string[];
}

export const IconMap: Record<string, any> = {
    "server": Server,
    "cloud": Cloud,
    "map": Map,
    "message-square": MessageSquare,
    "check-circle-2": CheckCircle2,
    "clock": Clock,
    "database": Database,
    "file-text": FileText,
    "github": Github,
    "activity": Activity,
    "globe": Globe,
    "zap": Zap,
    "default": Server
};

/**
 * A list of built-in service templates.
 * Deprecated: Templates are now fetched from the API.
 */
export const SERVICE_TEMPLATES: ServiceTemplate[] = [
  {
    id: "empty",
    name: "Custom Service",
    description: "Configure a service from scratch.",
    category: "Other",
    icon: Server,
    config: {
      name: "",
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      httpService: { address: "" } as any,
    },
  },
];
