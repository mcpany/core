import re
import sys

def parse_lcov(filename):
    coverage = {}
    with open(filename, 'r') as f:
        current_file = None
        lines_found = 0
        lines_hit = 0
        for line in f:
            line = line.strip()
            if line.startswith('SF:'):
                current_file = line[3:]
                lines_found = 0
                lines_hit = 0
            elif line.startswith('DA:'):
                parts = line[3:].split(',')
                if int(parts[1]) > 0:
                    lines_hit += 1
                lines_found += 1
            elif line == 'end_of_record':
                if current_file:
                    coverage[current_file] = (lines_hit, lines_found)

    return coverage

cov = parse_lcov(sys.argv[1])
for file, (hit, found) in sorted(cov.items(), key=lambda x: (x[1][0]/x[1][1] if x[1][1] > 0 else 0)):
    if found > 0:
        print(f"{hit/found*100:.1f}%\t{hit}/{found}\t{file}")
