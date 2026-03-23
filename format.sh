for file in $(find server/pkg -name "*.go" | grep -v "_test.go"); do
    gofmt -w $file
done
