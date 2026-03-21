/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { applyTemplateFields } from "./template-utils";
import { ServiceTemplate } from "./templates";

describe("applyTemplateFields", () => {
  it("should return config as-is if no fields are defined", () => {
    const template: ServiceTemplate = {
      id: "test",
      name: "Test",
      description: "Test",
      icon: null,
      config: { name: "test-service" },
    };
    const result = applyTemplateFields(template, {});
    expect(result).toEqual({ name: "test-service" });
  });

  it("should set values at specified keys", () => {
    const template: ServiceTemplate = {
      id: "test",
      name: "Test",
      description: "Test",
      icon: null,
      config: {
          name: "test",
          commandLineService: {
              env: { "API_KEY": "" }
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any
      },
      fields: [
        {
          name: "apiKey",
          label: "API Key",
          placeholder: "",
          key: "commandLineService.env.API_KEY"
        }
      ]
    };

    const result = applyTemplateFields(template, { apiKey: "secret-123" });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((result as any).commandLineService.env.API_KEY).toBe("secret-123");
  });

  it("should replace tokens in strings", () => {
    const template: ServiceTemplate = {
      id: "test",
      name: "Test",
      description: "Test",
      icon: null,
      config: {
          name: "test",
          commandLineService: {
              command: "run {{ARG}}"
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any
      },
      fields: [
        {
          name: "arg",
          label: "Arg",
          placeholder: "",
          key: "commandLineService.command",
          replaceToken: "{{ARG}}"
        }
      ]
    };

    const result = applyTemplateFields(template, { arg: "my-value" });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((result as any).commandLineService.command).toBe("run my-value");
  });

  it("should handle nested object creation if path does not exist", () => {
      const template: ServiceTemplate = {
        id: "test",
        name: "Test",
        description: "Test",
        icon: null,
        config: { name: "test" },
        fields: [
            {
                name: "deep",
                label: "Deep",
                placeholder: "",
                key: "a.b.c"
            }
        ]
      };

      const result = applyTemplateFields(template, { deep: "val" });
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect((result as any).a.b.c).toBe("val");
  });
});
