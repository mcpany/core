sed -i 's|"@com_github_mcpany_core//operator/api/v1alpha1"|"//k8s/operator/api/v1alpha1"|g' k8s/operator/cmd/BUILD.bazel
sed -i 's|"@com_github_mcpany_core//operator/controllers"|"//k8s/operator/controllers"|g' k8s/operator/cmd/BUILD.bazel
