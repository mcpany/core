/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// This file provides mock definitions for protobuf generated types
// used in tests and mock APIs.

/** Empty interface for test contexts */
export interface Empty {}

/** Error mock */
export interface Error {
    code: number;
    message: string;
    details: any[];
}

/** Mock type placeholders for config-related proto messages. */
export const UpstreamServiceConfig = {};
export const ServiceTemplate = {};
export const ProfileDefinition = {};
export const Profile = {};
export const ToolDefinition = {};
export const ResourceDefinition = {};
export const PromptDefinition = {};
export const WebhookDefinition = {};
export const SkillDefinition = {};

/**
 * Mock type placeholders for auth-related proto messages.
 */
export const Authentication = {};
export const Credential = {};

/**
 * Mock type placeholders for policy-related proto messages.
 */
export const CallPolicy = {};
/** CallPolicyRule */
export const CallPolicyRule = {};
/** ExportPolicy */
export const ExportPolicy = {};
/** ExportRule */
export const ExportRule = {};

/**
 * Mock constant for ProfileDefinition used in tests.
 */
export const MOCK_PROFILE: any = {
    id: "default",
    name: "Default Profile",
};
