/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    ParameterType,
    HttpCallDefinition_HttpMethod,
    OutputTransformer_OutputFormat,
    HttpCallDefinition
} from "./mock-proto";

describe("mock-proto definitions", () => {
    it("should export ParameterType enum correctly", () => {
        expect(ParameterType.STRING).toBe(0);
        expect(ParameterType.NUMBER).toBe(1);
    });

    it("should export HttpCallDefinition_HttpMethod enum correctly", () => {
        expect(HttpCallDefinition_HttpMethod.HTTP_METHOD_GET).toBe(1);
    });

    it("should export OutputTransformer_OutputFormat enum correctly", () => {
        expect(OutputTransformer_OutputFormat.JSON).toBe(0);
    });

    it("should allow creating an object matching HttpCallDefinition interface", () => {
        const def: HttpCallDefinition = {
            method: HttpCallDefinition_HttpMethod.HTTP_METHOD_GET,
            endpointPath: "/test",
            parameters: []
        };
        expect(def).toBeDefined();
        expect(def.method).toBe(HttpCallDefinition_HttpMethod.HTTP_METHOD_GET);
    });
});
