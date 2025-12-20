
# Shim Makefile to forward commands to server/Makefile
.PHONY: all test lint build run clean gen

all test lint build run clean gen:
	$(MAKE) -C server $@

# Forward any other targets
%:
	$(MAKE) -C server $@
