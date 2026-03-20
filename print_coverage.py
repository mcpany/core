import re
import sys

def parse_lcov(file_path):
    coverage_data = {}
    current_file = None
    lines_found = 0
    lines_hit = 0

    with open(file_path, 'r') as f:
        for line in f:
            if line.startswith('SF:'):
                current_file = line.strip().split(':')[1]
                coverage_data[current_file] = {'lines_found': 0, 'lines_hit': 0}
            elif line.startswith('LF:'):
                coverage_data[current_file]['lines_found'] = int(line.strip().split(':')[1])
            elif line.startswith('LH:'):
                coverage_data[current_file]['lines_hit'] = int(line.strip().split(':')[1])

    return coverage_data

def main():
    if len(sys.argv) < 2:
        print("Usage: python print_coverage.py <lcov_file>")
        sys.exit(1)

    coverage_data = parse_lcov(sys.argv[1])

    results = []
    for file, data in coverage_data.items():
        if data['lines_found'] > 0:
            pct = (data['lines_hit'] / data['lines_found']) * 100
            results.append((file, pct, data['lines_found'], data['lines_hit']))

    # Sort by number of uncovered lines descending (highest risk)
    results.sort(key=lambda x: x[2] - x[3], reverse=True)

    print(f"{'File':<60} | {'Found':<6} | {'Hit':<6} | {'Uncov':<6} | {'Pct':<6}")
    print("-" * 92)
    for file, pct, found, hit in results[:50]:
        print(f"{file:<60} | {found:<6} | {hit:<6} | {found-hit:<6} | {pct:.1f}%")

if __name__ == '__main__':
    main()
