module github.com/mcpany/core/k8s/operator

go 1.24

replace github.com/mcpany/core => ../../

require (
	github.com/mcpany/core v0.0.0
	k8s.io/api v0.35.3
	k8s.io/apimachinery v0.35.3
	sigs.k8s.io/controller-runtime v0.23.3
)
