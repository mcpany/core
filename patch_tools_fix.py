import re

with open("server/pkg/tool/types.go", "r") as f:
    content = f.read()

ruby_replacement = """	// Unquoted Ruby: @var and $var are dangerous.
	// $var is blocked by generic checks. @var is not.
	if quoteLevel == 0 {
		// Sentinel Security Update: Block %x(...) execution syntax
		// %x is dangerous in Unquoted contexts (code execution)
		if strings.Contains(val, "%x") {
			return fmt.Errorf("ruby execution injection detected: value contains '%%x'")
		}
		if strings.Contains(val, "@") {
			return fmt.Errorf("ruby variable injection detected: value contains '@'")
		}
	}

	// Sentinel Security Update:
	// Block leading pipe | to prevent open("|cmd") injection (the "Magic Open" vulnerability).
	// In Ruby, open() treats strings starting with '|' as commands to execute, even if the string
	// is quoted (single or double). This check prevents RCE when user input is interpolated into open().
	// We check for levels 1 (double), 2 (single), and 3 (backtick). Level 0 is blocked by unquoted checks.
	if quoteLevel >= 1 && quoteLevel <= 3 {
		if strings.HasPrefix(strings.TrimSpace(val), "|") {
			return fmt.Errorf("ruby open injection detected: value starts with '|'")
		}
	}

	return nil
}"""

content = re.sub(r'\t// Unquoted Ruby: @var and \$var are dangerous.\n\t// \$var is blocked by generic checks. @var is not.\n\tif quoteLevel == 0 {\n\t\t// Sentinel Security Update: Block %x\(\.\.\.\) execution syntax\n\t\t// %x is dangerous in Unquoted contexts \(code execution\)\n\t\tif strings.Contains\(val, "%x"\) {\n\t\t\treturn fmt.Errorf\("ruby execution injection detected: value contains \'%%x\'"\)\n\t\t}\n\t\tif strings.Contains\(val, "@"\) {\n\t\t\treturn fmt.Errorf\("ruby variable injection detected: value contains \'@\'"\)\n\t\t}\n\t}\n\n\treturn nil\n}', ruby_replacement, content)


perl_replacement = """					if next == -1 {
						break
					}
					idx += 2 + next
				}
			}
		}

		// Sentinel Security Update:
		// Block leading pipe | to prevent open(FH, "|cmd") injection in Perl.
		// This works in single quotes (level 2) as well.
		if quoteLevel >= 1 && quoteLevel <= 3 {
			if strings.HasPrefix(strings.TrimSpace(val), "|") {
				return fmt.Errorf("perl open injection detected: value starts with '|'")
			}
		}
	}
	return nil
}"""

content = re.sub(r'\t\t\t\t\tif next == -1 {\n\t\t\t\t\t\tbreak\n\t\t\t\t\t}\n\t\t\t\t\tidx \+= 2 \+ next\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n\treturn nil\n}', perl_replacement, content)

awk_replacement = """				if end >= len(val) || !isWordChar(val[end]) {
					// check start boundary
					if i == 0 || !isWordChar(val[i-1]) {
						return fmt.Errorf("awk injection detected: value contains 'system'")
					}
				}
			}
		}
		// Block '@' to prevent indirect function calls (e.g., @var(args)) which can bypass keyword checks,
		// as well as gawk extensions like @load and @include.
		if strings.Contains(val, "@") {
			return fmt.Errorf("awk injection detected: value contains '@'")
		}
	}
	return nil
}"""

content = re.sub(r'\t\t\t\tif end >= len\(val\) \|\| !isWordChar\(val\[end\]\) {\n\t\t\t\t\t// check start boundary\n\t\t\t\t\tif i == 0 \|\| !isWordChar\(val\[i-1\]\) {\n\t\t\t\t\t\treturn fmt.Errorf\("awk injection detected: value contains \'system\'"\)\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n\treturn nil\n}', awk_replacement, content)

unquoted_replacement = """	// We also block quotes and backslashes to prevent argument splitting and interpretation abuse
	// We also block control characters that could act as separators or cause confusion (\\r, \\t, \\v, \\f)
	// Sentinel Security Update: Added space (' ') to block list to prevent argument injection in shell commands
	// Sentinel Security Update: Space is only dangerous for Shells, not for Interpreters when using exec.Command
	dangerousChars := ";|&$`(){}!<>\\"\\n\\r\\t\\v\\f*?[]~#%^'\\\\"
	if isShell {
		dangerousChars += " "
	}

	charsToCheck := dangerousChars"""

content = re.sub(r'\t// We also block quotes and backslashes to prevent argument splitting and interpretation abuse\n\t// We also block control characters that could act as separators or cause confusion \(\\r, \\t, \\v, \\f\)\n\t// Sentinel Security Update: Added space \(\' \'\) to block list to prevent argument injection in shell commands\n\tdangerousChars := ";\|&\$`\(\){}!<>\\"\\n\\r\\t\\v\\f\*\?\[\]~#%\^\'\\\\ "\n\n\tcharsToCheck := dangerousChars', unquoted_replacement, content)

with open("server/pkg/tool/types.go", "w") as f:
    f.write(content)
