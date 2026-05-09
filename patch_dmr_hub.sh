sed -i 's/failedNodes = append(failedNodes, id)/failedNodes = append(failedNodes, id)\n\t\t\tdelete(h.nodes, id)/' server/pkg/dmr/hub.go
