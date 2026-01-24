/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface ProfileDefinition {
    name: string;
    selector?: ProfileSelector;
    requiredRoles?: string[];
    parentProfileIds?: string[];
    serviceConfig?: Record<string, ProfileServiceConfig>;
    secrets?: Record<string, SecretValue>;
}

export interface ProfileSelector {
    tags?: string[];
    toolProperties?: Record<string, string>;
}

export interface ProfileServiceConfig {
    enabled?: boolean;
}

export interface SecretValue {
    plainText?: string;
    environmentVariable?: string;
    filePath?: string;
    remoteContent?: RemoteContent;
    vault?: VaultSecret;
    awsSecretManager?: AwsSecretManagerSecret;
    validationRegex?: string;
}

export interface RemoteContent {
    httpUrl?: string;
    // auth omitted for simplicity as it creates cycle or complex import
}

export interface VaultSecret {
    address?: string;
    path?: string;
    key?: string;
}

export interface AwsSecretManagerSecret {
    secretId?: string;
    jsonKey?: string;
    versionStage?: string;
    versionId?: string;
    region?: string;
    profile?: string;
}
