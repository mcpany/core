sed -i 's|ServiceType: "ClusterIP",|ServiceType: "ClusterIP",\n\t\t\tConfigMap: "my-config-map",|g' k8s/operator/controllers/mcpserver_controller_test.go
