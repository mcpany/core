cd ui
pnpm run lint > lint-out.txt || true
grep -C 3 "rich-result-viewer" lint-out.txt
