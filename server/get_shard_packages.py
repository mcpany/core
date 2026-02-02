import sys

def main():
    if len(sys.argv) < 3:
        print("Usage: python3 get_shard_packages.py <shard_index> <shard_total>")
        sys.exit(1)

    shard_index = int(sys.argv[1])
    shard_total = int(sys.argv[2])

    packages = []
    try:
        # Read packages from stdin
        for line in sys.stdin:
            line = line.strip()
            if line:
                packages.append(line)
    except Exception as e:
        print(f"Error reading from stdin: {e}", file=sys.stderr)
        sys.exit(1)

    selected_packages = []
    count = 0
    for pkg in packages:
        # 1-based index logic from run_shard.sh:
        # if [ $(( (COUNT % SHARD_TOTAL) + 1 )) -eq "$SHARD_INDEX" ]; then
        if (count % shard_total) + 1 == shard_index:
            selected_packages.append(pkg)
        count += 1

    print(" ".join(selected_packages))

if __name__ == "__main__":
    main()
