# The tests failure:
# gentestmain: ParseFile("server/tests/framework/aitool.go"): server/tests/framework/aitool.go:16:2: expected declaration, found GetProjectRoot
# This implies my doc script inserted something bad in aitool.go on line 16. Let's just restore aitool.go!
import subprocess
subprocess.run("git checkout server/tests/framework/aitool.go", shell=True)
subprocess.run("git checkout server/tests/framework/copilot.go", shell=True)
subprocess.run("git checkout server/tests/framework/claude.go", shell=True)
subprocess.run("git checkout server/tests/framework/framework_test.go", shell=True)
