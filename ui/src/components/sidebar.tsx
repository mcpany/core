
import { NavItem } from "./nav-item"
import { LayoutDashboard, Server, Wrench, Database, MessageSquare, Settings, Activity, GitBranch } from "lucide-react"
import Link from "next/link"

export function Sidebar() {
  return (
    <div className="hidden border-r bg-muted/40 md:block min-h-screen w-[220px] lg:w-[280px]">
      <div className="flex h-full max-h-screen flex-col gap-2">
        <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6">
          <Link href="/" className="flex items-center gap-2 font-semibold">
            <Activity className="h-6 w-6 text-primary" />
            <span className="">MCP Any</span>
          </Link>
        </div>
        <div className="flex-1">
          <nav className="grid items-start px-2 text-sm font-medium lg:px-4">
            <NavItem href="/" icon={LayoutDashboard} title="Dashboard" isActive={true} />
            <NavItem href="/services" icon={Server} title="Services" />
            <NavItem href="/tools" icon={Wrench} title="Tools" />
            <NavItem href="/resources" icon={Database} title="Resources" />
            <NavItem href="/prompts" icon={MessageSquare} title="Prompts" />
            <div className="my-2 border-t" />
             <NavItem href="/middleware" icon={GitBranch} title="Middleware" />
            <NavItem href="/settings" icon={Settings} title="Settings" />
          </nav>
        </div>
        <div className="mt-auto p-4">
            <div className="text-xs text-muted-foreground text-center">
                v1.0.0
            </div>
        </div>
      </div>
    </div>
  )
}
