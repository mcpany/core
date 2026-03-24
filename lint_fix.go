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
        // Change `if _, err := w.Write([]byte("{}")); err != nil {`
        //   logging.GetLogger().Error("failed to write response", "error", err.Error())
        // }
        // Back to `_, _ = w.Write([]byte("{}"))` to be consistent with the rest of the file
        // because golangci-lint doesn't complain about _, _ = w.Write(...) when errcheck is configured properly,
        // UNLESS the tool configuration specifically enforces it.
        // Wait, did my first errcheck patch get merged or did it fail because of `opts.Marshal`?
        // Actually, my PR originally failed the `ci/circleci: lint` check BEFORE I touched `w.Write`.
        // The original issue was that `opts.Marshal(t.Tool())` returned an error that I ignored with `_`.
        // But then I used `w.Write` instead of `json.NewEncoder(w).Encode(toolList)`, so I introduced `_, _ = w.Write(buf)` inside handleTools.
        // It's highly likely that circleci uses a linter configuration that complains about ANY unhandled errors for `w.Write` inside our changes.
        // Let's revert my `if _, err := w.Write(...); err != nil` to `_, _ = w.Write(...)` if we aren't sure.
        // Wait, if I am forced to handle it by the CI, then `_, _ = w.Write` will fail CI again!
        // But what if the CI is just complaining about the `_` assignment to `err`?
        // No, CI complains about `opts.Marshal(t.Tool())` because it returns `([]byte, error)`. If we assign `b, _ := ...`, errcheck flags it.
		return true
	})
}
