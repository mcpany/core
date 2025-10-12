package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDockerComposeDeployment(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("Skipping Docker tests because SKIP_DOCKER_TESTS is set")
	}
	if !IsDockerSocketAccessible() {
		t.Skip("Skipping Docker tests because Docker socket is not accessible")
	}

	t.Parallel()
	root, err := GetProjectRoot()
	require.NoError(t, err)

	composeFile := filepath.Join(root, "docker-compose.yml")
	_, err = os.Stat(composeFile)
	require.NoError(t, err, "docker-compose.yml not found")

	dockerExe, dockerBaseArgs := getDockerCommand()
	args := append(dockerBaseArgs, "compose", "-f", composeFile, "up", "--build", "--abort-on-container-exit")
	composeProcess := NewManagedProcess(t, "docker-compose", dockerExe, args, nil)

	err = composeProcess.Start()
	require.NoError(t, err)

	defer composeProcess.Stop()

	// Check that the mcpxy server is healthy
	mcpxyHealthURL := "http://localhost:50050/healthz"
	WaitForHTTPHealth(t, mcpxyHealthURL, 60*time.Second)

	// Check that the http-echo-server is healthy
	echoHealthURL := "http://localhost:8080/health"
	WaitForHTTPHealth(t, echoHealthURL, 60*time.Second)

	// Stop the docker-compose process and check the exit code
	composeProcess.Stop()
	require.True(t, composeProcess.Cmd().ProcessState.Success(), "docker-compose up should exit cleanly")
}

func TestHelmChartDeployment(t *testing.T) {
	if os.Getenv("SKIP_K8S_TESTS") == "" {
		t.Skip("Skipping Kubernetes tests because SKIP_K8S_TESTS is not set")
	}
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Skipping Helm test because helm executable not found in PATH")
	}
	t.Parallel()

	root, err := GetProjectRoot()
	require.NoError(t, err)

	chartPath := filepath.Join(root, "helm", "mcpxy")
	releaseName := "mcpxy-test"

	// Install the chart
	installArgs := []string{"install", releaseName, chartPath, "--wait"}
	helmInstall := NewManagedProcess(t, "helm-install", "helm", installArgs, nil)
	err = helmInstall.Start()
	require.NoError(t, err)
	helmInstall.Stop() // Wait for the process to finish

	// Check the status of the release
	statusArgs := []string{"status", releaseName}
	helmStatus := NewManagedProcess(t, "helm-status", "helm", statusArgs, nil)
	err = helmStatus.Start()
	require.NoError(t, err)
	helmStatus.Stop() // Wait for the process to finish
	require.Contains(t, helmStatus.StdoutString(), "STATUS: deployed")

	// Port forward to the service
	port := FindFreePort(t)
	portForwardArgs := []string{"port-forward", fmt.Sprintf("service/%s-mcpxy", releaseName), fmt.Sprintf("%d:50050", port)}
	portForward := NewManagedProcess(t, "kubectl-port-forward", "kubectl", portForwardArgs, nil)
	err = portForward.Start()
	require.NoError(t, err)
	defer portForward.Stop()

	// Check that the server is healthy
	healthURL := fmt.Sprintf("http://localhost:%d/healthz", port)
	WaitForHTTPHealth(t, healthURL, 30*time.Second)

	// Uninstall the chart
	uninstallArgs := []string{"uninstall", releaseName}
	helmUninstall := NewManagedProcess(t, "helm-uninstall", "helm", uninstallArgs, nil)
	err = helmUninstall.Start()
	require.NoError(t, err)
	helmUninstall.Stop()
}