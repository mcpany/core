// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mcpany/core/server/pkg/validation"
)

// checkForShellInjection checks if the provided value contains any shell injection attempts
// given the context of the command execution.
//
// Parameters:
//   - val: The value to check for injection.
//   - template: The argument template containing the placeholder (used to determine quoting).
//   - placeholder: The placeholder string in the template (e.g. "{{arg}}").
//   - command: The command being executed (to determine interpreter context).
//   - isShell: Whether the command is a shell or not.
func checkForShellInjection(val string, template string, placeholder string, command string, isShell bool) error {
	// Determine the quoting context of the placeholder in the template
	quoteLevel := analyzeQuoteContext(template, placeholder)

	base := strings.ToLower(filepath.Base(command))
	isWindowsCmd := base == "cmd.exe" || base == "cmd"
	if isWindowsCmd && quoteLevel == 2 {
		quoteLevel = 0
	}

	// Sentinel Security Update: Interpreter Injection Protection
	// Check if the main command is an interpreter
	if isInterpreter(command) {
		if err := checkInterpreterInjection(val, template, base, quoteLevel); err != nil {
			return err
		}
		// Sentinel Security Update: Interpreter Strict Mode
		// Block dangerous function calls and keywords commonly used for RCE
		// in both single and double-quoted strings (which might be evaluated).
		if quoteLevel == 1 || quoteLevel == 2 {
			if err := checkInterpreterFunctionCalls(val, base); err != nil {
				return err
			}
		}
	}

	// Check if the argument itself (template) invokes an interpreter.
	// This covers cases where the main command is a shell or runner (e.g. bash -c "awk ...")
	// and the argument is the command line for that interpreter.
	args := strings.Fields(template)
	if len(args) > 0 {
		argBase := strings.ToLower(filepath.Base(args[0]))
		// Avoid double checking if it's the same command (already checked above)
		if argBase != base && isInterpreter(argBase) {
			if err := checkInterpreterInjection(val, template, argBase, quoteLevel); err != nil {
				return fmt.Errorf("argument interpreter injection detected (%s): %w", argBase, err)
			}
			// Also check function calls for the detected interpreter context
			if quoteLevel == 1 || quoteLevel == 2 {
				if err := checkInterpreterFunctionCalls(val, argBase); err != nil {
					return fmt.Errorf("argument interpreter injection detected (%s): %w", argBase, err)
				}
			}
		}
	}

	if quoteLevel == 3 { // Backticked
		return checkBacktickInjection(val, command)
	}

	if quoteLevel == 2 { // Single Quoted
		if strings.Contains(val, "'") {
			return fmt.Errorf("shell injection detected: value contains single quote which breaks out of single-quoted argument")
		}

		// Block backticks (used by Perl, Ruby, PHP for execution)
		if strings.Contains(val, "`") {
			return fmt.Errorf("shell injection detected: value contains backtick inside single-quoted argument (potential interpreter abuse)")
		}

		// Block dangerous function calls (system, exec, popen, eval) followed by open parenthesis
		// We use a case-insensitive check for robustness, although most interpreters are case-sensitive.
		// We normalize by removing whitespace to detect "system (" or "system\t(".
		var b strings.Builder
		b.Grow(len(val))
		for _, r := range val {
			if !unicode.IsSpace(r) {
				b.WriteRune(r)
			}
		}
		cleanVal := strings.ToLower(b.String())

		dangerousCalls := []string{"system(", "exec(", "popen(", "eval("}
		for _, call := range dangerousCalls {
			if strings.Contains(cleanVal, call) {
				return fmt.Errorf("shell injection detected: value contains dangerous function call %q inside single-quoted argument (potential interpreter abuse)", call)
			}
		}

		return nil
	}

	if quoteLevel == 1 { // Double Quoted
		// In double quotes, dangerous characters are double quote, $, and backtick
		// We also need to block backslash because it can be used to escape the closing quote
		// % is also dangerous in Windows CMD inside double quotes
		if idx := strings.IndexAny(val, "\"$`\\%"); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside double-quoted argument", val[idx])
		}
		return nil
	}

	return checkUnquotedInjection(val, command, isShell)
}

func getCommentStyles(language string) (bool, bool, bool) {
	switch language {
	case "python", "ruby", "perl", "sh", "bash", "zsh", "dash", "ash", "ksh", "csh", "tcsh", "fish":
		return true, false, false
	case "node", "nodejs", "bun", "deno", "java", "c", "cpp", "go", "rust", "swift", "kotlin", "scala", "groovy":
		return false, true, true
	case "php":
		return true, true, true
	default:
		// Default to strict: strip all known comment types if unsure
		return true, true, true
	}
}

type interpreterCommentStripper struct {
	inLine, inBlock bool
	inS, inD, inB   bool
	escaped         bool

	useHash, useSlash, useBlock bool
}

func (l *interpreterCommentStripper) process(val string) string {
	var b strings.Builder
	b.Grow(len(val))

	for i := 0; i < len(val); i++ {
		char := val[i]

		if l.inLine {
			if char == '\n' {
				l.inLine = false
				b.WriteByte(char)
			}
			continue
		}
		if l.inBlock {
			if char == '*' && i+1 < len(val) && val[i+1] == '/' {
				l.inBlock = false
				i++
			}
			continue
		}

		if l.escaped {
			l.escaped = false
			b.WriteByte(char)
			continue
		}

		if l.handleQuote(char, &b) {
			continue
		}

		// Comments
		if l.useHash && char == '#' {
			l.inLine = true
			continue
		}
		if (l.useSlash || l.useBlock) && char == '/' && i+1 < len(val) {
			if l.useSlash && val[i+1] == '/' {
				l.inLine = true
				i++
				continue
			}
			if l.useBlock && val[i+1] == '*' {
				l.inBlock = true
				i++
				continue
			}
		}

		// Line continuation
		if char == '\\' {
			continue
		}

		b.WriteByte(char)
	}
	return b.String()
}

func (l *interpreterCommentStripper) handleQuote(char byte, b *strings.Builder) bool {
	// Quote Toggling
	if !l.inD && !l.inB && char == '\'' {
		l.inS = !l.inS
		b.WriteByte(char)
		return true
	}
	if !l.inS && !l.inB && char == '"' {
		l.inD = !l.inD
		b.WriteByte(char)
		return true
	}
	if !l.inS && !l.inD && char == '`' {
		l.inB = !l.inB
		b.WriteByte(char)
		return true
	}

	if l.inS || l.inD || l.inB {
		if char == '\\' {
			l.escaped = true
			b.WriteByte(char)
		} else {
			b.WriteByte(char)
		}
		return true
	}
	return false
}

func stripInterpreterComments(val, language string) string {
	useHash, useSlash, useBlock := getCommentStyles(language)
	lexer := &interpreterCommentStripper{
		useHash:  useHash,
		useSlash: useSlash,
		useBlock: useBlock,
	}
	return lexer.process(val)
}

func checkInterpreterFunctionCalls(val, language string) error {
	// Strip comments and line continuations first
	val = stripInterpreterComments(val, language)

	// Normalize value to detect obfuscation (e.g. system ( ) )
	var b strings.Builder
	b.Grow(len(val))
	for _, r := range val {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	cleanVal := strings.ToLower(b.String())

	dangerousKeywords := []string{
		"system", "exec", "popen", "eval",
		"spawn", "fork",
		"import", "require",
		"subprocess", "child_process", "os", "sys",
		"open", "read", "write",
	}

	for _, kw := range dangerousKeywords {
		// Sentinel Security Update: Check for keyword followed by delimiters other than '('
		// Languages like Ruby and Perl allow calling functions without parentheses (e.g. system 'ls').
		// We check against cleanVal (no whitespace), so 'system "ls"' becomes 'system"ls"'.
		if strings.Contains(cleanVal, kw+"(") ||
			strings.Contains(cleanVal, kw+"'") ||
			strings.Contains(cleanVal, kw+"\"") ||
			strings.Contains(cleanVal, kw+"`") {
			return fmt.Errorf("interpreter injection detected: value contains dangerous function call %q", kw)
		}
	}

	if strings.Contains(cleanVal, "__import__") {
		return fmt.Errorf("interpreter injection detected: value contains '__import__'")
	}
	return nil
}

func checkInterpreterInjection(val, template, base string, quoteLevel int) error {
	if err := checkTarInjection(val, base); err != nil {
		return err
	}
	if err := checkPythonInjection(val, template, base); err != nil {
		return err
	}
	if err := checkRubyInjection(val, base, quoteLevel); err != nil {
		return err
	}
	if err := checkNodePerlPhpInjection(val, base, quoteLevel); err != nil {
		return err
	}
	if err := checkAwkInjection(val, base); err != nil {
		return err
	}
	if err := checkSQLInjection(val, base, quoteLevel); err != nil {
		return err
	}
	return nil
}

func checkTarInjection(val, base string) error {
	// Tar Injection Check
	// Block RCE via --checkpoint-action and --to-command flags by detecting execution directives
	// inside flag values.
	isTar := base == "tar" || base == "gtar" || base == "bsdtar"
	if isTar {
		valLower := strings.ToLower(val)
		// Check for execution directives commonly used in --checkpoint-action and --to-command
		if strings.Contains(valLower, "exec=") || strings.Contains(valLower, "command=") {
			return fmt.Errorf("tar injection detected: value contains execution directive")
		}
		// Also block checkpoint-action keyword itself if it somehow appears in value
		if strings.Contains(valLower, "checkpoint-action") {
			return fmt.Errorf("tar injection detected: value contains 'checkpoint-action'")
		}
	}
	return nil
}

func checkSQLInjection(val, base string, quoteLevel int) error {
	// SQL Injection Check
	// If the command is a SQL client (psql, mysql, sqlite3) and the value is unquoted (Level 0),
	// we must prevent SQL injection by blocking SQL keywords.
	isSQL := base == "psql" || base == "mysql" || base == "sqlite3"
	if isSQL && quoteLevel == 0 {
		// Block common SQL keywords and comment markers
		// We check for keywords surrounded by word boundaries or at start/end of string.
		// val is user input, e.g. "1 OR 1=1"
		upperVal := strings.ToUpper(val)
		keywords := []string{
			"OR", "AND", "UNION", "SELECT", "FROM", "WHERE", "JOIN",
			"DROP", "ALTER", "CREATE", "INSERT", "UPDATE", "DELETE",
			"--",
		}

		// Helper to check word boundary
		isBoundary := func(r byte) bool {
			return !isWordChar(r)
		}

		for _, kw := range keywords {
			if kw == "--" {
				if strings.Contains(upperVal, "--") {
					return fmt.Errorf("SQL injection detected: value contains '--'")
				}
				continue
			}

			idx := strings.Index(upperVal, kw)
			for idx != -1 {
				// Check boundaries
				startOk := idx == 0 || isBoundary(upperVal[idx-1])
				endOk := idx+len(kw) == len(upperVal) || isBoundary(upperVal[idx+len(kw)])

				if startOk && endOk {
					return fmt.Errorf("SQL injection detected: value contains SQL keyword %q in unquoted context", kw)
				}
				// Find next occurrence
				nextIdx := strings.Index(upperVal[idx+1:], kw)
				if nextIdx == -1 {
					break
				}
				idx += 1 + nextIdx
			}
		}
	}
	return nil
}

func checkPythonInjection(val, template, base string) error {
	// Python: Check for f-string prefix in template
	if strings.HasPrefix(base, "python") {
		// Scan template to find the prefix of the quote containing the placeholder
		// Given complexity, we use a heuristic: if template contains f" or f', enforce checks.
		hasFString := false
		for i := 0; i < len(template)-1; i++ {
			if template[i+1] == '\'' || template[i+1] == '"' {
				prefix := strings.ToLower(getPrefix(template, i+1))
				if prefix == "f" || prefix == "fr" || prefix == "rf" {
					hasFString = true
					break
				}
			}
		}
		if hasFString {
			if strings.ContainsAny(val, "{}") {
				return fmt.Errorf("python f-string injection detected: value contains '{' or '}'")
			}
		}
	}
	return nil
}

func checkRubyInjection(val, base string, quoteLevel int) error {
	// Ruby: #{...} works in double quotes AND backticks
	if strings.HasPrefix(base, "ruby") && (quoteLevel == 1 || quoteLevel == 3) { // Double Quoted or Backticked
		if strings.Contains(val, "#{") {
			return fmt.Errorf("ruby interpolation injection detected: value contains '#{'")
		}
		// Block leading pipe | to prevent open("|cmd") injection
		if strings.HasPrefix(strings.TrimSpace(val), "|") {
			return fmt.Errorf("ruby open injection detected: value starts with '|'")
		}
	}
	return nil
}

func checkNodePerlPhpInjection(val, base string, quoteLevel int) error {
	// Node/JS/Perl/PHP: ${...} works in backticks (JS) or double quotes (Perl/PHP)
	isNode := strings.HasPrefix(base, "node") || base == "bun" || base == "deno"
	isPerl := strings.HasPrefix(base, "perl")
	isPhp := strings.HasPrefix(base, "php")

	if isNode && quoteLevel == 3 { // Backtick
		if strings.Contains(val, "${") {
			return fmt.Errorf("javascript template literal injection detected: value contains '${'")
		}
	}
	// Perl and PHP interpolate variables in both double quotes and backticks
	if (isPerl || isPhp) && (quoteLevel == 1 || quoteLevel == 3) { // Double Quoted or Backticked
		if strings.Contains(val, "${") {
			return fmt.Errorf("variable interpolation injection detected: value contains '${'")
		}
	}

	if isPerl {
		// Sentinel Security Update:
		// Block qx operator (command execution) regardless of quoting.
		// qx can be used in unquoted contexts with safe delimiters (e.g. qx/cmd/)
		// avoiding common shell injection filters.
		if strings.Contains(val, "qx") {
			// Sentinel Security Update: Block qx in all contexts (unquoted, double-quoted, backticked).
			// While qx// inside double quotes is technically a string literal in some contexts,
			// sophisticated interpolation attacks or misinterpretation of quote context makes it risky.
			// Blocking "qx" is aggressive but necessary for strict security on Perl input.
			if quoteLevel == 0 || quoteLevel == 1 || quoteLevel == 3 {
				return fmt.Errorf("shell injection detected: perl qx execution")
			}
		}

		if quoteLevel == 1 || quoteLevel == 3 {
			if strings.Contains(val, "@{") {
				return fmt.Errorf("perl array interpolation injection detected: value contains '@{'")
			}
		}
	}
	return nil
}

func checkAwkInjection(val, base string) error {
	// Awk: Block pipe | to prevent external command execution
	// Also block redirection > and < to prevent arbitrary file read/write
	// And block getline to prevent file reading
	isAwk := strings.HasPrefix(base, "awk") || strings.HasPrefix(base, "gawk") || strings.HasPrefix(base, "nawk") || strings.HasPrefix(base, "mawk")
	if isAwk {
		if strings.Contains(val, "|") {
			return fmt.Errorf("awk injection detected: value contains '|'")
		}
		if strings.Contains(val, ">") {
			return fmt.Errorf("awk injection detected: value contains '>'")
		}
		if strings.Contains(val, "<") {
			return fmt.Errorf("awk injection detected: value contains '<'")
		}
		if strings.Contains(val, "getline") {
			return fmt.Errorf("awk injection detected: value contains 'getline'")
		}
	}
	return nil
}

func checkBacktickInjection(val, command string) error {
	// Backticks in Shell are command substitution (Level 0 danger).
	// Unless it is a known interpreter that uses backticks safely (like JS template literals),
	// we must enforce strict checks.
	if !isSafeBacktickLanguage(command) {
		const dangerousChars = ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\ "
		if idx := strings.IndexAny(val, dangerousChars); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside backticks", val[idx])
		}
	}
	// For interpreters (like JS), we already handled specific injections above.
	// We should still prevent breaking out of backticks.
	if strings.Contains(val, "`") {
		return fmt.Errorf("backtick injection detected")
	}
	return nil
}

func isSafeBacktickLanguage(command string) bool {
	base := strings.ToLower(filepath.Base(command))
	// Only JS/TS runtimes treat backticks as strings (template literals)
	safe := []string{"node", "nodejs", "bun", "deno"}
	for _, s := range safe {
		if base == s || strings.HasPrefix(base, s) {
			return true
		}
	}
	return false
}

func checkUnquotedInjection(val, command string, isShell bool) error {
	// Unquoted (or unknown quoting): strict check
	// Block common shell metacharacters and globbing/expansion characters
	// % and ^ are Windows CMD metacharacters
	// We also block quotes and backslashes to prevent argument splitting and interpretation abuse
	// We also block control characters that could act as separators or cause confusion (\r, \t, \v, \f)
	// Sentinel Security Update: Added space (' ') to block list to prevent argument injection in shell commands
	// Sentinel Security Update: Space is only dangerous for Shells, not for Interpreters when using exec.Command
	dangerousChars := ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\"
	if isShell {
		dangerousChars += " "
	}

	charsToCheck := dangerousChars
	// For 'env' command, '=' is dangerous as it allows setting arbitrary environment variables
	if filepath.Base(command) == "env" {
		charsToCheck += "="
	}

	if idx := strings.IndexAny(val, charsToCheck); idx != -1 {
		return fmt.Errorf("shell injection detected: value contains dangerous character %q", val[idx])
	}
	return nil
}

func isInterpreter(command string) bool {
	base := strings.ToLower(filepath.Base(command))
	interpreters := []string{
		// Common interpreters and runners that can execute code
		"python", "ruby", "perl", "php",
		"node", "nodejs", "bun", "deno",
		"lua", "awk", "gawk", "nawk", "mawk", "sed",
		"jq",
		"psql", "mysql", "sqlite3",
		"docker",
		"tclsh", "wish",
		"irb", "php-cgi",
		// Editors and pagers that can execute commands
		"vi", "vim", "nvim", "emacs", "nano",
		"less", "more", "man",
		// Build tools and others that can execute commands
		"find", "xargs", "tee",
		"make", "rake", "ant", "mvn", "gradle",
		"npm", "yarn", "pnpm", "npx", "bunx", "go", "cargo", "pip",
		// Cloud/DevOps tools that can execute commands or have sensitive flags
		"kubectl", "helm", "aws", "gcloud", "az", "terraform", "ansible", "ansible-playbook",
		// Additional interpreters and compilers that can execute code
		"r", "rscript", "julia", "groovy", "jshell",
		"scala", "kotlin", "swift",
		"elixir", "iex", "erl", "escript",
		"ghci", "clisp", "sbcl", "lisp", "scheme", "racket",
		"lua", "luajit",
		"gcc", "g++", "clang", "java",
		// Additional dangerous tools
		"zip", "unzip", "rsync", "nmap", "tcpdump", "gdb", "lldb",
		"tar", "gtar", "bsdtar",
	}
	for _, interp := range interpreters {
		if base == interp || strings.HasPrefix(base, interp) {
			return true
		}
	}

	// Check for script extensions that indicate interpretation
	ext := strings.ToLower(filepath.Ext(base))
	scriptExts := []string{
		".js", ".mjs", ".ts",
		".py", ".pyc", ".pyo", ".pyd",
		".rb", ".pl", ".pm", ".php",
		".lua", ".r",
	}
	for _, scriptExt := range scriptExts {
		if ext == scriptExt {
			return true
		}
	}

	return false
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func getPrefix(s string, idx int) string {
	// idx is index of quote char
	start := idx - 1
	for start >= 0 {
		c := s[start]
		if !isWordChar(c) {
			break
		}
		start--
	}
	return s[start+1 : idx]
}

func analyzeQuoteContext(template, placeholder string) int {
	if template == "" || placeholder == "" {
		return 0
	}

	// Levels: 0 = Unquoted (Strict), 1 = Double, 2 = Single, 3 = Backtick
	minLevel := 3

	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	foundAny := false

	for i := 0; i < len(template); i++ {
		// Check if we match placeholder at current position
		if strings.HasPrefix(template[i:], placeholder) {
			foundAny = true
			currentLevel := 0
			switch {
			case inSingle:
				currentLevel = 2
			case inBacktick:
				currentLevel = 3
			case inDouble:
				currentLevel = 1
			}

			if currentLevel < minLevel {
				minLevel = currentLevel
			}

			// Advance past placeholder
			i += len(placeholder) - 1
			continue
		}

		char := template[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' && !inSingle {
			escaped = true
			continue
		}

		switch char {
		case '\'':
			if !inDouble && !inBacktick {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && !inBacktick {
				inDouble = !inDouble
			}
		case '`':
			if !inSingle {
				inBacktick = !inBacktick
			}
		}
	}

	if !foundAny {
		return 0 // Should not happen if called correctly, fallback to strict
	}

	// logging.GetLogger().Info("analyzeQuoteContext", "template", template, "placeholder", placeholder, "level", minLevel)
	return minLevel
}

func checkEnvInjection(val string) error {
	// Relaxed check for environment variables.
	// Allows spaces, but blocks shell metacharacters.
	// We rely on validateSafePathAndInjection to prevent argument injection (flags starting with -).
	const dangerousChars = ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\" // Space removed
	if idx := strings.IndexAny(val, dangerousChars); idx != -1 {
		return fmt.Errorf("shell injection detected: value contains dangerous character %q", val[idx])
	}
	return nil
}

func validateSafePathAndInjection(val string, isDocker bool, commandName string) error {
	// Sentinel Security Update: Trim whitespace to prevent bypasses using leading spaces
	val = strings.TrimSpace(val)

	// Sentinel Security Update: Enforce SSRF protection on arguments that look like URLs.
	// We check for "://" to capture any scheme (http, https, ftp, gopher, etc.).
	// IsSafeURL will block any scheme other than http/https, and verify IPs for those.
	if strings.Contains(val, "://") {
		if err := validation.IsSafeURL(val); err != nil {
			return fmt.Errorf("unsafe url argument: %w", err)
		}
	}

	if err := checkForPathTraversal(val); err != nil {
		return err
	}
	// Also check decoded value just in case the input was already encoded
	if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
		if err := checkForPathTraversal(decodedVal); err != nil {
			return fmt.Errorf("%w (decoded)", err)
		}
	}

	if !isDocker {
		if err := checkForLocalFileAccess(val); err != nil {
			return err
		}
		// Also check decoded value for local file access (e.g. %66ile://)
		if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
			if err := checkForLocalFileAccess(decodedVal); err != nil {
				return fmt.Errorf("%w (decoded)", err)
			}
		}
	}

	if err := checkForArgumentInjection(val); err != nil {
		return err
	}
	// Also check decoded value for argument injection (e.g. %2drf)
	if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
		if err := checkForArgumentInjection(decodedVal); err != nil {
			return fmt.Errorf("%w (decoded)", err)
		}
	}

	// Sentinel Security Update: Block dangerous pseudo-protocols/schemes
	// We ONLY block these for tools known to be vulnerable (ImageMagick, FFmpeg, Git, etc.)
	// Blocking them for generic tools (like echo) causes false positives (usability regression).
	if isVulnerableToSchemes(commandName) {
		if err := checkForDangerousSchemes(val); err != nil {
			return err
		}
		// Also check decoded value for dangerous schemes
		if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
			if err := checkForDangerousSchemes(decodedVal); err != nil {
				return fmt.Errorf("%w (decoded)", err)
			}
		}
	}

	return nil
}

func isVulnerableToSchemes(command string) bool {
	base := strings.ToLower(filepath.Base(command))

	// ImageMagick tools
	magickTools := []string{
		"convert", "mogrify", "identify", "composite", "compare", "stream",
		"montage", "display", "animate", "import", "conjure", "magick",
	}
	for _, tool := range magickTools {
		if base == tool {
			return true
		}
	}

	// FFmpeg tools
	ffmpegTools := []string{"ffmpeg", "ffprobe", "ffplay"}
	for _, tool := range ffmpegTools {
		if base == tool {
			return true
		}
	}

	// Git
	if base == gitCommand {
		return true
	}

	// Interpreters (PHP, Expect) are handled by checkForShellInjection but
	// if they are used as the main command, we might want to be extra careful.
	// For now, let's stick to the ones that process "magic" files/URIs.

	return false
}

func checkForDangerousSchemes(val string) error {
	// Check for "scheme:..." pattern. Scheme must be at start.
	// We look for the first colon.
	idx := strings.Index(val, ":")
	if idx == -1 {
		return nil
	}

	// Extract scheme and convert to lower case
	scheme := strings.ToLower(val[:idx])

	// Validate scheme characters (alpha, digit, +, -, .) to prevent false positives on random colons
	for _, r := range scheme {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '+' || r == '-' || r == '.') {
			return nil // Not a valid scheme pattern, likely just text with a colon
		}
	}

	// Blocklist of dangerous schemes used for LFI, RCE, or SSRF in various tools
	dangerous := map[string]bool{
		// Generic / Interpreter
		"file": true, "gopher": true, "expect": true, "php": true,
		"zip": true, "jar": true, "war": true,

		// ImageMagick (convert, mogrify, identify, etc.)
		"mvg": true, "msl": true, "vid": true, "ephemeral": true,
		"label": true, "text": true, "info": true, "pango": true,
		"caption": true, "plasma": true, "xc": true, "inline": true,
		"gradient": true, "pattern": true, "tile": true, "read": true,

		// FFmpeg
		"concat": true, "subfile": true, "crypto": true, "data": true,
		"hls": true, "http": false, "https": false, // explicitly allowed (handled by IsSafeURL if :// present)
		"ftp": true, "rtmp": true, "rtsp": true,

		// Git
		"ext": true, // Block git ext::
	}

	if dangerous[scheme] {
		return fmt.Errorf("dangerous scheme detected: %s", scheme)
	}

	return nil
}
