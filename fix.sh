#!/bin/bash
sed -i 's/configv1\.VectorUpstreamService_builder{/configv1.VectorUpstreamService_builder{/g' server/pkg/util/secrets_sanitizer_test.go
