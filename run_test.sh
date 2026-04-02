#!/bin/bash
./bazelisk test //ui:playwright_tests_verification_services_spec_ts --test_output=errors
