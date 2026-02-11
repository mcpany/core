
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ServiceDiagnostics } from "./service-diagnostics";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock apiClient
vi.mock("@/lib/client", async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...actual,
    apiClient: {
      validateService: vi.fn(),
      getServiceStatus: vi.fn(),
      listTools: vi.fn(),
    },
  };
});

describe("ServiceDiagnostics", () => {
  const mockService: UpstreamServiceConfig = {
    id: "test-service",
    name: "test-service",
    version: "1.0.0",
    disable: false,
    priority: 0,
    httpService: {
      address: "http://example.com/api",
    } as any, // Cast to any to avoid full partial implementation details
  };

  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  it("runs diagnostics including browser connectivity check", async () => {
    // Setup mocks
    (apiClient.validateService as any).mockResolvedValue({ valid: true });
    (apiClient.getServiceStatus as any).mockResolvedValue({ status: "Active" });
    (apiClient.listTools as any).mockResolvedValue({ tools: [] });
    (global.fetch as any).mockResolvedValue({
      ok: true,
      status: 200,
    });

    render(<ServiceDiagnostics service={mockService} />);

    // Click run button
    fireEvent.click(screen.getByText("Run Diagnostics"));

    // Check for "Browser Connectivity"
    await waitFor(() => {
      expect(screen.getByText("Browser Connectivity")).toBeDefined();
    });

    // Verify fetch was called
    expect(global.fetch).toHaveBeenCalledWith("http://example.com/api", expect.anything());

    // Verify success status (green check is implicit by text check usually, but let's check text)
    await waitFor(() => {
        expect(screen.getByText("Service is reachable from browser.")).toBeDefined();
    });
  });

  it("reports warning when browser connectivity fails (CORS/Network)", async () => {
     // Setup mocks
     (apiClient.validateService as any).mockResolvedValue({ valid: true });
     (apiClient.getServiceStatus as any).mockResolvedValue({ status: "Active" });
     (apiClient.listTools as any).mockResolvedValue({ tools: [] });
     (global.fetch as any).mockRejectedValue(new Error("Network Error"));

     render(<ServiceDiagnostics service={mockService} />);

     fireEvent.click(screen.getByText("Run Diagnostics"));

     await waitFor(() => {
       expect(screen.getByText("Browser Connectivity")).toBeDefined();
     });

     await waitFor(() => {
         expect(screen.getByText("Failed to reach service from browser.")).toBeDefined();
         expect(screen.getByText(/Network Error/)).toBeDefined();
     });
  });

  it("warns about localhost usage (Docker heuristic)", async () => {
    const localhostService: UpstreamServiceConfig = {
        ...mockService,
        httpService: {
            address: "http://localhost:3000"
        } as any
    };

    (apiClient.validateService as any).mockResolvedValue({ valid: true });
    (apiClient.getServiceStatus as any).mockResolvedValue({ status: "Active" });
    (apiClient.listTools as any).mockResolvedValue({ tools: [] });
    (global.fetch as any).mockResolvedValue({ ok: true });

    render(<ServiceDiagnostics service={localhostService} />);

    fireEvent.click(screen.getByText("Run Diagnostics"));

    await waitFor(() => {
        expect(screen.getByText("Localhost Configuration")).toBeDefined();
    });

    expect(screen.getByText(/You are using 'localhost'/)).toBeDefined();
    expect(screen.getByText(/host.docker.internal/)).toBeDefined();
  });
});
