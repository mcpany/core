# Shim Makefile to forward commands to server/Makefile and ui/Makefile
.PHONY: all test lint build run clean gen

# Targets that should run on both (or just server generally, but test on both)
all:
	$(MAKE) -C server all
	$(MAKE) -C ui build

test:
	$(MAKE) -C server test
	$(MAKE) -C ui test

k8s-e2e:
	./tests/k8s/test_operator.sh

lint:
	$(MAKE) -C server lint

# Run runs server
run:
	$(MAKE) -C server run

clean:
	$(MAKE) -C server clean
	# ui clean if needed, likely just removing node_modules or build artifacts?

gen:
	$(MAKE) -C server gen

# Forward other targets to server by default
%:
	$(MAKE) -C server $@
