import { render, screen, fireEvent } from "@testing-library/react";
import { SmartTemplateEditor } from "./smart-template-editor";
import { describe, it, expect, vi } from "vitest";

// Mock useToast hook since it's used in the component
vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock ResizeObserver which is often needed for UI components in tests
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

describe("SmartTemplateEditor", () => {
  it("renders with initial value", () => {
    render(
      <SmartTemplateEditor
        value="Hello {{ name }}"
        onChange={() => {}}
        variables={["name"]}
      />
    );
    expect(screen.getByDisplayValue("Hello {{ name }}")).toBeInTheDocument();
  });

  it("calls onChange when typing in template", () => {
    const handleChange = vi.fn();
    render(
      <SmartTemplateEditor
        value=""
        onChange={handleChange}
      />
    );
    const textarea = screen.getByPlaceholderText(/Enter Jinja2 template/i);
    fireEvent.change(textarea, { target: { value: "New Template" } });
    expect(handleChange).toHaveBeenCalledWith("New Template");
  });

  it("displays variables as badges", () => {
    render(
      <SmartTemplateEditor
        value=""
        onChange={() => {}}
        variables={["user_id", "query"]}
      />
    );
    expect(screen.getByText("user_id")).toBeInTheDocument();
    expect(screen.getByText("query")).toBeInTheDocument();
  });
});
