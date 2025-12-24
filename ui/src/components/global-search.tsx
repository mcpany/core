"use client"

import * as React from "react"
import {
    CreditCard,
    Settings,
    User,
    LayoutDashboard,
    Server,
    Wrench,
    BookOpen,
    Play,
    Search
} from "lucide-react"

import {
    CommandDialog,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
    CommandSeparator,
    CommandShortcut,
} from "@/components/ui/command"
import { useRouter } from "next/navigation"
import { useTheme } from "next-themes"
import { Button } from "@/components/ui/button"

export function GlobalSearch() {
    const [open, setOpen] = React.useState(false)
    const router = useRouter()
    const { setTheme } = useTheme()

    React.useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
                e.preventDefault()
                setOpen((open) => !open)
            }
        }

        document.addEventListener("keydown", down)
        return () => document.removeEventListener("keydown", down)
    }, [])

    const runCommand = React.useCallback((command: () => unknown) => {
        setOpen(false)
        command()
    }, [])

    return (
        <>
            <Button
                variant="outline"
                className="relative h-9 w-full justify-start rounded-[0.5rem] bg-background text-sm font-normal text-muted-foreground shadow-none sm:pr-12 md:w-40 lg:w-64"
                onClick={() => setOpen(true)}
            >
                <span className="hidden lg:inline-flex">Search...</span>
                <span className="inline-flex lg:hidden">Search...</span>
                <kbd className="pointer-events-none absolute right-[0.3rem] top-[0.3rem] hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
                    <span className="text-xs">⌘</span>K
                </kbd>
            </Button>
            <CommandDialog open={open} onOpenChange={setOpen}>
                <CommandInput placeholder="Type a command or search..." />
                <CommandList>
                    <CommandEmpty>No results found.</CommandEmpty>
                    <CommandGroup heading="Suggestions">
                        <CommandItem onSelect={() => runCommand(() => router.push("/"))}>
                            <LayoutDashboard className="mr-2 h-4 w-4" />
                            <span>Dashboard</span>
                        </CommandItem>
                        <CommandItem onSelect={() => runCommand(() => router.push("/services"))}>
                            <Server className="mr-2 h-4 w-4" />
                            <span>Services</span>
                        </CommandItem>
                        <CommandItem onSelect={() => runCommand(() => router.push("/tools"))}>
                            <Wrench className="mr-2 h-4 w-4" />
                            <span>Tools</span>
                        </CommandItem>
                        <CommandItem onSelect={() => runCommand(() => router.push("/playground"))}>
                            <Play className="mr-2 h-4 w-4" />
                            <span>Playground</span>
                        </CommandItem>
                    </CommandGroup>
                    <CommandSeparator />
                    <CommandGroup heading="Services">
                        <CommandItem onSelect={() => runCommand(() => router.push("/services/payment-gateway"))}>
                            <CreditCard className="mr-2 h-4 w-4" />
                            <span>Payment Gateway</span>
                        </CommandItem>
                        <CommandItem onSelect={() => runCommand(() => router.push("/services/user-service"))}>
                            <User className="mr-2 h-4 w-4" />
                            <span>User Service</span>
                        </CommandItem>
                    </CommandGroup>
                     <CommandSeparator />
                    <CommandGroup heading="Tools">
                         <CommandItem onSelect={() => runCommand(() => router.push("/tools/get_weather"))}>
                            <Wrench className="mr-2 h-4 w-4" />
                            <span>get_weather</span>
                        </CommandItem>
                         <CommandItem onSelect={() => runCommand(() => router.push("/tools/list_files"))}>
                            <Wrench className="mr-2 h-4 w-4" />
                            <span>list_files</span>
                        </CommandItem>
                    </CommandGroup>
                    <CommandSeparator />
                    <CommandGroup heading="Theme">
                        <CommandItem onSelect={() => runCommand(() => setTheme("light"))}>
                            <SunIcon className="mr-2 h-4 w-4" />
                            <span>Light</span>
                        </CommandItem>
                        <CommandItem onSelect={() => runCommand(() => setTheme("dark"))}>
                            <MoonIcon className="mr-2 h-4 w-4" />
                            <span>Dark</span>
                        </CommandItem>
                        <CommandItem onSelect={() => runCommand(() => setTheme("system"))}>
                            <LaptopIcon className="mr-2 h-4 w-4" />
                            <span>System</span>
                        </CommandItem>
                    </CommandGroup>
                     <CommandSeparator />
                    <CommandGroup heading="Settings">
                        <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
                            <Settings className="mr-2 h-4 w-4" />
                            <span>Settings</span>
                            <CommandShortcut>⌘S</CommandShortcut>
                        </CommandItem>
                    </CommandGroup>
                </CommandList>
            </CommandDialog>
        </>
    )
}

function SunIcon(props: React.SVGProps<SVGSVGElement>) {
    return (
        <svg
            {...props}
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <circle cx="12" cy="12" r="4" />
            <path d="M12 2v2" />
            <path d="M12 20v2" />
            <path d="m4.93 4.93 1.41 1.41" />
            <path d="m17.66 17.66 1.41 1.41" />
            <path d="M2 12h2" />
            <path d="M20 12h2" />
            <path d="m6.34 17.66-1.41 1.41" />
            <path d="m19.07 4.93-1.41 1.41" />
        </svg>
    )
}

function MoonIcon(props: React.SVGProps<SVGSVGElement>) {
    return (
        <svg
            {...props}
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
        </svg>
    )
}

function LaptopIcon(props: React.SVGProps<SVGSVGElement>) {
    return (
        <svg
            {...props}
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <rect width="18" height="12" x="3" y="4" rx="2" ry="2" />
            <line x1="2" x2="22" y1="20" y2="20" />
        </svg>
    )
}
