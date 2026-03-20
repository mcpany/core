import sys
content = open("server/pkg/tool/types.go").read()

if "execfile" not in content:
    content = content.replace('''"syscall", "dlopen", "fiddle", "send", "__send__", "public_send",''', '''"syscall", "dlopen", "fiddle", "send", "__send__", "public_send",
		"execfile", "evalfile", "passthru", "shell_exec", "proc_open", "popen",
		"setTimeout", "setInterval", "setImmediate", "Function",''')
    content = content.replace('''"compile", "globals", "locals", "vars",''', '''"compile", "globals", "locals", "vars",
			"execfile", "evalfile", "passthru", "shell_exec", "proc_open",
			"setTimeout", "setInterval", "setImmediate", "Function",''')

open("server/pkg/tool/types.go", "w").write(content)
