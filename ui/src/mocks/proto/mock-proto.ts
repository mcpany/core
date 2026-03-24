/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock class for GrpcWebImpl used in tests.
 */
export class GrpcWebImpl {}

/**
 * Mock class for RegistrationServiceClientImpl used in tests.
 */
export class RegistrationServiceClientImpl {}

/**
 * Mock constant for HttpUpstreamService used in tests.
 */
export const HttpUpstreamService = {};

/**
 * Mock constant for ServiceProvenance used in tests.
 */
export const ServiceProvenance = {};

/**
 * Mock enum for CallPolicy_Action used in tests.
 */
export enum CallPolicy_Action {
  ALLOW = 0,
  DENY = 1,
  SAVE_CACHE = 2,
  DELETE_CACHE = 3,
}

/**
 * Mock enum for ExportPolicy_Action used in tests.
 */
export enum ExportPolicy_Action {
  EXPORT = 0,
  UNEXPORT = 1,
}

/**
 * Mock type placeholders for policy-related proto messages.
 */
export const CallPolicy = {};

<<<<<<< HEAD
/** Mock type placeholder. */
export const CallPolicyRule = {};

/** Mock type placeholder. */
export const ExportPolicy = {};

/** Mock type placeholder. */
=======
/**
 * Mock type placeholder for CallPolicyRule.
 */
export const CallPolicyRule = {};

/**
 * Mock type placeholder for ExportPolicy.
 */
export const ExportPolicy = {};

/**
 * Mock type placeholder for ExportRule.
 */
>>>>>>> origin/main
export const ExportRule = {};

/**
 * Mock constant for ProfileDefinition used in tests.
 */
export const ProfileDefinition = {};

/**
 * Mock constant for ToolDefinition used in tests.
 */
export const ToolDefinition = {};

/**
 * Mock enum for HttpCallDefinition_HttpMethod used in tests.
 */
export enum HttpCallDefinition_HttpMethod {
  HTTP_METHOD_UNSPECIFIED = 0,
  HTTP_METHOD_GET = 1,
  HTTP_METHOD_POST = 2,
  HTTP_METHOD_PUT = 3,
  HTTP_METHOD_DELETE = 4,
  HTTP_METHOD_PATCH = 5,
}

/**
 * Mock enum for OutputTransformer_OutputFormat used in tests.
 */
export enum OutputTransformer_OutputFormat {
  JSON = 0,
  XML = 1,
  TEXT = 2,
  RAW_BYTES = 3,
  JQ = 4,
}

/**
 * Mock interface/type for HttpCallDefinition used in tests.
 */
export const HttpCallDefinition = {};

/**
 * Mock enum for ParameterType used in tests.
 */
export enum ParameterType {
  STRING = 0,
  NUMBER = 1,
  INTEGER = 2,
  BOOLEAN = 3,
  ARRAY = 4,
  OBJECT = 5,
}

/**
 * Mock type for InputTransformer used in tests.
 */
export const InputTransformer = {};

/**
 * Mock type for OutputTransformer used in tests.
 */
export const OutputTransformer = {};

/**
 * Mock type for HttpParameterMapping used in tests.
 */
export const HttpParameterMapping = {};

/**
 * Mock type for ResourceDefinition used in tests.
 */
export const ResourceDefinition = {};

/**
 * Mock type for PromptDefinition used in tests.
 */
export const PromptDefinition = {};

/**
 * Mock type for Credential used in tests.
 */
export const Credential = {};

/**
 * Mock type for Authentication used in tests.
 */
export const Authentication = {};
