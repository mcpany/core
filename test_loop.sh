for i in {1..20}; do
  echo "Iteration $i"
  go test -v ./server/pkg/app -run TestConfigLoadingWithArg
done
