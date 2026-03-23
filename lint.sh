export GOLANGCI_LINT_BIN=/home/jules/go/bin/golangci-lint
cd server
$GOLANGCI_LINT_BIN run --out-format=line-number > lint_out.txt
cat lint_out.txt | wc -l
