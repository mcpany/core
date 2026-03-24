#!/bin/bash
sed -i 's/callDef := configv1.HttpCallDefinition_builder{/_ = configv1.HttpCallDefinition_builder{/' server/tests/integration/e2e_helpers.go
