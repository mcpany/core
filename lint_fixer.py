import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')

    # We want to identify unused parameters and other basic linting issues that typically fail in CI.
    # Actually, we don't know what fails. Let's look at the CI logs if possible.
    pass

# We can run gofmt since gofumpt is often used in CI
os.system("cd server && go fmt ./...")
