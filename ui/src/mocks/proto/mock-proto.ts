/**
 * Copyright 2026 Author(s) of MCP Any.
 * SPDX-License-Identifier: Apache-2.0.
 */

/**
 * Mock class for GrpcWebImpl used in tests.
 */
export class GrpcWebImpl { }

/**
 * Mock class for RegistrationServiceClientImpl used in tests.
 */
export class RegistrationServiceClientImpl { }

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
  /** Allow the action. */
  ALLOW = 0,
  /** Deny the action. */
  DENY = 1,
  /** Save to cache. */
  SAVE_CACHE = 2,
  /** Delete from cache. */
  DELETE_CACHE = 3,
}

/**
 * Mock enum for ExportPolicy_Action used in tests.
 */
export enum ExportPolicy_Action {
  /** Export the resource. */
  EXPORT = 0,
  /** Unexport the resource. */
  UNEXPORT = 1,
}

/**
 * Mock type placeholders for policy-related proto messages.
 */
export const CallPolicy = {};
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
  /** Unspecified method. */
  HTTP_METHOD_UNSPECIFIED = 0,
  /** GET method. */
  HTTP_METHOD_GET = 1,
  /** POST method. */
  HTTP_METHOD_POST = 2,
  /** PUT method. */
  HTTP_METHOD_PUT = 3,
  /** DELETE method. */
  HTTP_METHOD_DELETE = 4,
  /** PATCH method. */
  HTTP_METHOD_PATCH = 5,
}

/**
 * Mock enum for OutputTransformer_OutputFormat used in tests.
 */
export enum OutputTransformer_OutputFormat {
  /** JSON format. */
  JSON = 0,
  /** XML format. */
  XML = 1,
  /** TEXT format. */
  TEXT = 2,
  /** RAW BYTES format. */
  RAW_BYTES = 3,
  /** JQ format. */
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
  /** String type. */
  STRING = 0,
  /** Number type. */
  NUMBER = 1,
  /** Integer type. */
  INTEGER = 2,
  /** Boolean type. */
  BOOLEAN = 3,
  /** Array type. */
  ARRAY = 4,
  /** Object type. */
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
