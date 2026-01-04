
import { NextResponse } from "next/server";

export async function GET() {
  // Mock data for dashboard metrics
  // In a real scenario, this would aggregate data from the backend
  const metrics = [
    {
      label: "Total Requests",
      value: "2,345",
      change: "+12.5%",
      trend: "up",
      icon: "Activity",
      subLabel: "vs. last hour"
    },
    {
      label: "Avg. Latency",
      value: "45ms",
      change: "-5.2%",
      trend: "down", // down is good for latency
      icon: "Clock",
      subLabel: "vs. last hour"
    },
    {
      label: "Active Services",
      value: "12",
      change: "+2",
      trend: "up",
      icon: "Server",
      subLabel: "2 registering"
    },
    {
      label: "Error Rate",
      value: "0.05%",
      change: "+0.01%",
      trend: "up", // up is bad for errors
      icon: "AlertCircle",
      subLabel: "vs. last hour"
    }
  ];

  return NextResponse.json(metrics);
}
