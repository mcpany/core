import re
import sys

def parse_lcov(filename, target_file):
    coverage = []
    with open(filename, 'r') as f:
        current_file = None
        lines_found = 0
        lines_hit = 0
        in_target = False
        for line in f:
            line = line.strip()
            if line.startswith('SF:'):
                current_file = line[3:]
                in_target = current_file.endswith(target_file)
            elif line.startswith('DA:') and in_target:
                parts = line[3:].split(',')
                if int(parts[1]) == 0:
                    coverage.append(int(parts[0]))
    return coverage

cov = parse_lcov(sys.argv[1], sys.argv[2])
ranges = []
if len(cov) > 0:
    start = cov[0]
    prev = cov[0]
    for c in cov[1:]:
        if c == prev + 1:
            prev = c
        else:
            if start == prev:
                ranges.append(str(start))
            else:
                ranges.append(f"{start}-{prev}")
            start = c
            prev = c
    if start == prev:
        ranges.append(str(start))
    else:
        ranges.append(f"{start}-{prev}")
print(f"Missing lines in {sys.argv[2]}: {', '.join(ranges)}")
