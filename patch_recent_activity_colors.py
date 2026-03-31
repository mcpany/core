import re

with open('ui/src/components/dashboard/recent-activity-widget.tsx', 'r') as f:
    content = f.read()

# Update old "CheckCircle2" colors
content = content.replace(
    'isError ? "border-destructive/50 text-destructive" : hasResponseDiff ? "border-blue-500/50 text-blue-500" : "border-emerald-500/50 text-emerald-500"',
    'isError ? "border-destructive/50 text-destructive" : hasResponseDiff ? "border-blue-500/50 text-blue-500" : "border-emerald-500/50 text-emerald-500"'
)

with open('ui/src/components/dashboard/recent-activity-widget.tsx', 'w') as f:
    f.write(content)
