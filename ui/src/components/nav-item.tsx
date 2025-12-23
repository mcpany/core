
import { cn } from "@/lib/utils"
import { LucideIcon } from "lucide-react"
import Link from "next/link"

interface NavItemProps {
  href: string
  icon: LucideIcon
  title: string
  isActive?: boolean
}

export function NavItem({ href, icon: Icon, title, isActive }: NavItemProps) {
  return (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all hover:text-primary",
        isActive
          ? "bg-muted text-primary"
          : "text-muted-foreground"
      )}
    >
      <Icon className="h-4 w-4" />
      {title}
    </Link>
  )
}
