
export class GrpcWebImpl {
    constructor(host: string, options: any) {}
}

export class RegistrationServiceClientImpl {
    constructor(rpc: any) {}
    GetService(request: any, metadata: any) { return Promise.resolve({}); }
}

export interface UpstreamServiceConfig {
    id?: string;
    name?: string;
    version?: string;
    disable?: boolean;
    priority?: number;
    loadBalancingStrategy?: string;
    httpService?: any;
    grpcService?: any;
    commandLineService?: any;
    mcpService?: any;
    preCallHooks?: any[];
    postCallHooks?: any[];
    [key: string]: any;
}

export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

export interface Credential {
    id?: string;
    [key: string]: any;
}

export interface Authentication {
    [key: string]: any;
}

export type ListServicesResponse = any;
export type GetServiceResponse = any;
export type GetServiceStatusResponse = any;
