package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// ProfileDefinition matches the JSON structure expected by the API
type ProfileDefinition struct {
	Name             string            `json:"name"`
	RequiredRoles    []string          `json:"required_roles"`
	ParentProfileIDs []string          `json:"parent_profile_ids"`
	Selector         ProfileSelector   `json:"selector"`
	ServiceConfig    map[string]any    `json:"service_config"`
	Secrets          map[string]any    `json:"secrets"`
}

type ProfileSelector struct {
	Tags           []string          `json:"tags"`
	ToolProperties map[string]string `json:"tool_properties"`
}

func main() {
	baseURL := os.Getenv("MCP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:50050"
	}

	profiles := []ProfileDefinition{
		{
			Name:          "development",
			RequiredRoles: []string{"developer"},
			Selector: ProfileSelector{
				Tags: []string{"env:dev"},
			},
		},
		{
			Name:          "production",
			RequiredRoles: []string{"admin"},
			Selector: ProfileSelector{
				Tags: []string{"env:prod"},
			},
		},
		{
			Name:             "staging",
			RequiredRoles:    []string{"developer", "qa"},
			ParentProfileIDs: []string{"development"},
			Selector: ProfileSelector{
				Tags: []string{"env:staging"},
			},
		},
	}

	for _, p := range profiles {
		fmt.Printf("Seeding profile: %s... ", p.Name)
		if err := createProfile(baseURL, p); err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Println("Success")
		}
	}
}

func createProfile(baseURL string, p ProfileDefinition) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/profiles", baseURL), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("api returned status: %s", resp.Status)
	}

	return nil
}
