package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "pkg/app/api.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			isUnderscore := false
			for _, lhs := range x.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
					isUnderscore = true
					break
				}
			}

			if isUnderscore && len(x.Rhs) > 0 {
				if call, ok := x.Rhs[0].(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if sel.Sel.Name == "Write" || sel.Sel.Name == "Encode" || sel.Sel.Name == "Close" || sel.Sel.Name == "Marshal" {
							// Change `_ = x` to `x` and append `//nolint:errcheck`
							// Wait, if we change `_ = x` to just `x`, golangci-lint STILL reports unhandled error unless we put `//nolint:errcheck`.
							// Wait, `errcheck` looks for unassigned errors. So `w.Write(...)` by itself triggers errcheck.
							// The correct way to suppress errcheck on a single line is to have the comment on the SAME line as the statement.
							// Like:
							// _ = w.Write(...) //nolint:errcheck
							// Let's modify the file directly by lines.
						}
					}
				}
			}
		}
		return true
	})
}
