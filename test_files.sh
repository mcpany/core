for file in $(find server -name "*.go" | grep -v "test.go"); do
    echo "$file $(wc -l < $file) $(git blame $file 2>/dev/null | grep -i 'Coverage' | wc -l)"
done
