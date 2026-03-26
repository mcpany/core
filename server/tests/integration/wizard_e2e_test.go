//go:build e2e

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/playwright-community/playwright-go"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
)

func TestMarketplaceWizardE2E(t *testing.T) {
	// 1. Start Server
	serverInfo := StartMCPANYServer(t, "WizardE2ETest")
	defer serverInfo.CleanupFunc()

	ctx := context.Background()

	// Seed the database with a manual template to start the wizard
	seedData := app.SeedRequest{
		TemplatesRaw: []json.RawMessage{
			[]byte(`{
				"id": "e2e-test-template",
				"name": "E2E Test Template",
				"description": "Template for E2E testing",
				"service_config": {
					"name": "e2e-test-service",
					"command_line_service": {
						"command": "python3"
					}
				}
			}`),
		},
	}
	err := serverInfo.App.SeedDataPublic(ctx, seedData)
	require.NoError(t, err)

	// 2. Start Playwright
	err = playwright.Install()
	require.NoError(t, err)

	pw, err := playwright.Run()
	require.NoError(t, err)
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err)
	defer browser.Close()

	page, err := browser.NewPage()
	require.NoError(t, err)

	// 3. Navigate to Marketplace
	// Wait for the UI server to be up? In integration tests, the UI is served by the backend
	// via a static filesystem.
	_, err = page.Goto(serverInfo.HTTPAddress + "/marketplace")
	require.NoError(t, err)

	// Wait for network idle or specific element
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	// 4. Open Wizard
	err = page.Click("text=Create Custom Service")
	require.NoError(t, err)

	// Step 1: Service Type
	err = page.SelectOption("select[id='service-template']", playwright.SelectOptionValues{
		Values: playwright.StringSlice("e2e-test-template"),
	})
	require.NoError(t, err)
	err = page.Click("text=Next")
	require.NoError(t, err)

	// Step 2: Parameters (Our new UI)
	err = page.Click("text=Add Argument")
	require.NoError(t, err)

	// Type into the first argument input
	err = page.Fill("input[placeholder='Argument 1']", "--script.py")
	require.NoError(t, err)

	err = page.Click("text=Add Argument")
	require.NoError(t, err)

	// Type into the second argument input
	err = page.Fill("input[placeholder='Argument 2']", "--verbose")
	require.NoError(t, err)

	// Go next all the way to review
	err = page.Click("text=Next")
	require.NoError(t, err)
	err = page.Click("text=Next")
	require.NoError(t, err)
	err = page.Click("text=Next")
	require.NoError(t, err)

	// Review & Submit
	err = page.Click("text=Finish & Save to Local Marketplace")
	require.NoError(t, err)

	// Wait a moment for backend save
	time.Sleep(500 * time.Millisecond)

	// 5. Verify Backend State Change
	services, err := serverInfo.App.Storage.ListServices(ctx)
	require.NoError(t, err)

	var foundService *configv1.UpstreamServiceConfig
	for _, s := range services {
		if s.GetName() == "e2e-test-service" {
			foundService = s
			break
		}
	}

	require.NotNil(t, foundService, "Service should have been created in the backend")

	cmdService := foundService.GetCommandLineService()
	require.NotNil(t, cmdService, "Should be a command line service")
	require.Equal(t, "python3", cmdService.GetCommand())

	args := cmdService.GetArgs()
	require.Len(t, args, 2)
	require.Equal(t, "--script.py", args[0])
	require.Equal(t, "--verbose", args[1])
}
