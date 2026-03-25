module github.com/mcpany/core

go 1.24

replace github.com/mcpany/core => ./

require (
	google.golang.org/genproto/googleapis/api v0.0.0-20260319201613-d00831a3d3e7
	k8s.io/api v0.35.3
	k8s.io/apimachinery v0.35.3
	sigs.k8s.io/controller-runtime v0.23.3
)
