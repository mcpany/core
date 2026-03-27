package dummy

import (
	_ "k8s.io/api/apps/v1"
	_ "k8s.io/apimachinery/pkg/runtime"
	_ "sigs.k8s.io/controller-runtime/pkg/client"
)
