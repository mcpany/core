#!/bin/bash
sed -i 's/_ = configv1.HttpCallDefinition_builder{/callDef := configv1.HttpCallDefinition_builder{/' server/tests/integration/e2e_helpers.go
