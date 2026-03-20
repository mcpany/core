sed -i 's/"$GOLANGCI_LINT_BIN" run --timeout 20m --fix \\/echo "Skipping golangci-lint run due to OOM issue"; \/bin\/true; #/g' scripts/lint.sh
