package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "server/pkg/app/api.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
            if len(assign.Lhs) > 0 {
                if ident, ok := assign.Lhs[0].(*ast.Ident); ok && ident.Name == "_" {
                    // Turn `_, err = ...` into `_, _ = ...` or remove the underscore?
                    // Actually, if it's `_, err := w.Write(...)`, then using the err is good. The linter complains about the first `_`?
                    // NO! errcheck complains about `_, _ = w.Write(...)`!
                    // If we did `_, err := w.Write(...)`, it does NOT complain if we use `err`.
                    // But in my latest commit I did:
                    // _, err := w.Write(...)
                    // if err != nil ...
                    // Is `_` assignment causing a different lint error? "errcheck"? No, ineffassign?
                    // Let's check `errcheck` output. Oh, the previous `errcheck` output said:
                    // `internal error: package "github.com/puzpuzpuz/xsync/v4" without types was imported from "github.com/mcpany/core/server/pkg/auth"`
                }
            }
		}
		return true
	})
}
