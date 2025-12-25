
import { NextResponse } from 'next/server';

export async function GET() {
  // Mock data for dashboard metrics
  // In a real scenario, this would aggregate data from the backend
  const metrics = [
    {
      label: "Total Requests",
      value: "2,345",
      change: "+20.1%",
      trend: "up",
      icon: "Activity",
    },
    {
      label: "Active Services",
      value: "12",
      change: "+2",
      trend: "up",
      icon: "Server",
    },
    {
      label: "Connected Tools",
      value: "573",
      change: "+201",
      trend: "up",
      icon: "Zap",
    },
    {
      label: "Active Users",
      value: "45",
      change: "+12%",
      trend: "up",
      icon: "Users",
    },
  ];

  return NextResponse.json(metrics);
}
