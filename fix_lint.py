import re
import os

with open('server/lint.out', 'r') as f:
    for line in f:
        if 'revive: exported' in line and 'should have comment or be unexported' in line:
            parts = line.split(':')
            filepath = os.path.join('server', parts[0])
            line_num = int(parts[1])
            msg = line.split('revive:')[1].strip()
            print(f"{filepath}:{line_num} {msg}")
