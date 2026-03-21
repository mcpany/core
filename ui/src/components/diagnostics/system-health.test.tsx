import { render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { SystemHealth } from "./system-health";
import { apiClient } from "@/lib/client";

// Mock the API client
vi.mock("@/lib/client", () => ({
  apiClient: {
    getDoctorStatus: vi.fn(),
    getSystemStatus: vi.fn(),
  },
}));

describe("SystemHealth", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("renders loading state initially", () => {
    // Return unresolved promises to stay in loading state
    vi.mocked(apiClient.getDoctorStatus).mockReturnValue(new Promise(() => {}));
    vi.mocked(apiClient.getSystemStatus).mockReturnValue(new Promise(() => {}));

    render(<SystemHealth />);

    expect(screen.getByText("Evaluating system integrity...")).toBeInTheDocument();
  });

  it("renders healthy report correctly", async () => {
    const mockReport = {
      status: "healthy",
      timestamp: new Date().toISOString(),
      checks: {
        network: { status: "ok", message: "All clear", latency: "2ms" },
      },
    };
    const mockSysStatus = {
      active_connections: 5,
      bound_http_port: 8080,
      bound_grpc_port: 50051,
      security_warnings: [],
      has_tls: true,
      has_auth: true,
      uptime_seconds: 3600,
      version: "1.0.0"
    };
    vi.mocked(apiClient.getDoctorStatus).mockResolvedValue(mockReport as any);
    vi.mocked(apiClient.getSystemStatus).mockResolvedValue(mockSysStatus);

    render(<SystemHealth />);

    await waitFor(() => {
        expect(screen.getByText('Network is Secure')).toBeInTheDocument();
        // The text is split into two elements: 100 and /100
        expect(screen.getByText('100')).toBeInTheDocument();
        expect(screen.getByText('/100')).toBeInTheDocument();
        expect(screen.getByText(/network/i)).toBeInTheDocument();
        expect(screen.getByText(/All clear/i)).toBeInTheDocument();
    });
  });

  it("renders error state when api fails", async () => {
    vi.mocked(apiClient.getDoctorStatus).mockRejectedValue(new Error("API Error"));
    vi.mocked(apiClient.getSystemStatus).mockRejectedValue(new Error("API Error"));

    render(<SystemHealth />);

    await waitFor(() => {
      expect(screen.getByText("Diagnostics Failed")).toBeInTheDocument();
      expect(screen.getByText(/Failed to retrieve diagnostics report/i)).toBeInTheDocument();
    });
  });

  it("renders degraded state correctly", async () => {
    const mockReport = {
      status: "degraded",
      timestamp: new Date().toISOString(),
      checks: {
        database: { status: "warning", message: "High latency", latency: "500ms" },
      },
    };
    const mockSysStatus = {
        active_connections: 5,
        bound_http_port: 8080,
        bound_grpc_port: 50051,
        security_warnings: ["No TLS configured on public interface"],
        has_tls: false,
        has_auth: true,
        uptime_seconds: 3600,
        version: "1.0.0"
    };

    vi.mocked(apiClient.getDoctorStatus).mockResolvedValue(mockReport as any);
    vi.mocked(apiClient.getSystemStatus).mockResolvedValue(mockSysStatus);

    render(<SystemHealth />);

     await waitFor(() => {
        expect(screen.getByText('Action Recommended')).toBeInTheDocument();
        expect(screen.getByText(/No TLS configured on public interface/i)).toBeInTheDocument();
        expect(screen.getByText('database')).toBeInTheDocument();
        // The text is split into two elements: 70 and /100
        expect(screen.getByText('70')).toBeInTheDocument();
        expect(screen.getByText('/100')).toBeInTheDocument();
    });
  });
});
