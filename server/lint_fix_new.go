package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "pkg/app/api.go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	b, err := os.ReadFile("pkg/app/api.go")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ExprStmt:
			if call, ok := x.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "Write" || sel.Sel.Name == "Marshal" || sel.Sel.Name == "Encode" || sel.Sel.Name == "Close" {
						lineNum := fset.Position(x.Pos()).Line - 1
						if !strings.Contains(lines[lineNum], "//nolint:errcheck") {
                            // Only flag if it doesn't already have it
                            // Actually it shouldn't just be an ExprStmt, wait.
                            // If it has //nolint:errcheck it's technically fine.
                            fmt.Printf("Line %d: unassigned call to %s without nolint\n", lineNum+1, sel.Sel.Name)
						}
					}
				}
			}
		}
		return true
	})
}
