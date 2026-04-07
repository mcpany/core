1. **Target:** `server/pkg/tool/types.go` specifically functions like `stripInterpreterComments` and `checkUnquotedKeywords` which have high cyclomatic complexity but lack tests.
2. **Risk Profile:** The code executes and parses code tools, which can include shell scripts and various scripting languages. There is risk of parsing issues if quotes, comments or backslashes are not properly handled.
3. **New Coverage:** Write table-driven tests for `stripInterpreterComments` and `checkUnquotedKeywords` in `server/pkg/tool/types_parser_test.go` to test their edge cases like comments inside strings, backslashes before comments, block comments, hash comments etc.
4. **Verification:** I will check the tests run locally cleanly and generate an Impact Report in Markdown format as requested.
