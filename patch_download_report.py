import sys

with open("ui/src/app/page.tsx", "r") as f:
    content = f.read()

# Add necessary imports
new_imports = """import { DownloadReportButton } from "@/components/dashboard/download-report-button";
"""
content = content.replace('import { Loader2 } from "lucide-react";\n', 'import { Loader2 } from "lucide-react";\n' + new_imports)

# Replace <Button>Download Report</Button> with our new component
content = content.replace("<Button>Download Report</Button>", "<DownloadReportButton />")
content = content.replace('import { Button } from "@/components/ui/button";\n', "")

with open("ui/src/app/page.tsx", "w") as f:
    f.write(content)
