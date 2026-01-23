# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import sys
import json

class AuditManager:
    def __init__(self):
        self.report_file = "AUDIT_REPORT.md"
        self.findings = []
        self.load_findings()

    def load_findings(self):
        try:
            with open("audit_findings.json", "r") as f:
                self.findings = json.load(f)
        except FileNotFoundError:
            self.findings = []

    def save_findings(self):
        with open("audit_findings.json", "w") as f:
            json.dump(self.findings, f, indent=2)

    def add_finding(self, feature, doc_path, status, notes, changes="None"):
        # Check if exists
        for f in self.findings:
            if f["feature"] == feature and f["doc_path"] == doc_path:
                f["status"] = status
                f["notes"] = notes
                f["changes"] = changes
                self.save_findings()
                return

        self.findings.append({
            "feature": feature,
            "doc_path": doc_path,
            "status": status,
            "notes": notes,
            "changes": changes
        })
        self.save_findings()

    def generate_report(self):
        with open(self.report_file, "w") as f:
            f.write("# Documentation Audit & Verification Report\n\n")
            f.write("## 1. Features Audited\n\n")
            f.write("| Feature | Document | Status | Notes | Changes Made |\n")
            f.write("|---|---|---|---|---|\n")
            for finding in self.findings:
                status_icon = "✅" if finding['status'] == "PASS" else "❌"
                if finding['status'] == "OUTDATED": status_icon = "⚠️"
                f.write(f"| {finding['feature']} | `{finding['doc_path']}` | {status_icon} {finding['status']} | {finding['notes']} | {finding['changes']} |\n")

            f.write("\n## 2. Issues & Fixes\n")
            for finding in self.findings:
                if finding['status'] != "PASS":
                    f.write(f"\n### {finding['feature']}\n")
                    f.write(f"- **Issue**: {finding['notes']}\n")
                    f.write(f"- **Fix**: {finding['changes']}\n")

            f.write("\n## 3. Roadmap Alignment\n")
            f.write("All features were checked against the Roadmap.\n")

if __name__ == "__main__":
    manager = AuditManager()
    if len(sys.argv) > 1:
        cmd = sys.argv[1]
        if cmd == "add":
            # python audit_manager.py add "Feature Name" "path/to/doc" "STATUS" "Notes" "Changes"
            changes = sys.argv[6] if len(sys.argv) > 6 else "None"
            manager.add_finding(sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5], changes)
        elif cmd == "generate":
            manager.generate_report()
            print(f"Report generated at {manager.report_file}")
