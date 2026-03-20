#!/bin/bash
export PATH=$PATH:$HOME/bin
./bazelisk test //ui:playwright_tests_e2e_audit_log_spec_ts --test_output=errors
