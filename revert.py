import os
import subprocess

def run_cmd(cmd):
    try:
        return subprocess.check_output(cmd, shell=True).decode('utf-8')
    except subprocess.CalledProcessError as e:
        return e.output.decode('utf-8')

# Revert the changes to see if linter is broken purely by my additions
run_cmd("git restore .")
