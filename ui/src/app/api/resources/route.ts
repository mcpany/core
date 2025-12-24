
import { NextResponse } from "next/server";

const resources = [
  { id: "res_01", name: "Project Guidelines", type: "text/plain", service: "Documentation Service" },
  { id: "res_02", name: "Logo Assets", type: "application/zip", service: "Asset Store" },
];

export async function GET() {
  return NextResponse.json(resources);
}
