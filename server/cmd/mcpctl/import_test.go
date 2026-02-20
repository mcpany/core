package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImportCmd(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpctl-import-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		inputFile      string
		inputContent   string
		args           []string
		expectedConfig *McpAnyConfig
		expectedError  string
		checkFile      bool
		outputFile     string
	}{
		{
			name:      "Happy Path Stdout",
			inputFile: "claude_config.json",
			inputContent: `{
				"mcpServers": {
					"server1": {
						"command": "npx",
						"args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/test/Desktop"],
						"env": {
							"NODE_ENV": "development"
						}
					}
				}
			}`,
			expectedConfig: &McpAnyConfig{
				UpstreamServices: []UpstreamService{
					{
						Name: "server1",
						McpService: &McpService{
							StdioConnection: &StdioConnection{
								Command: "npx",
								Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/Users/test/Desktop"},
								Env:     map[string]string{"NODE_ENV": "development"},
							},
						},
					},
				},
			},
		},
		{
			name:      "Happy Path File Output",
			inputFile: "claude_config_file.json",
			inputContent: `{
				"mcpServers": {
					"server2": {
						"command": "python3",
						"args": ["server.py"]
					}
				}
			}`,
			args:       []string{"--output", "output.yaml"},
			checkFile:  true,
			outputFile: "output.yaml",
			expectedConfig: &McpAnyConfig{
				UpstreamServices: []UpstreamService{
					{
						Name: "server2",
						McpService: &McpService{
							StdioConnection: &StdioConnection{
								Command: "python3",
								Args:    []string{"server.py"},
								Env:     nil, // Empty in input
							},
						},
					},
				},
			},
		},
		{
			name:          "File Not Found",
			inputFile:     "non_existent.json",
			inputContent:  "",
			expectedError: "failed to read input file",
		},
		{
			name:      "Invalid JSON",
			inputFile: "invalid.json",
			inputContent: `{
				"mcpServers": {
					"server1": {
			}`,
			expectedError: "failed to parse Claude Desktop config",
		},
		{
			name:      "Empty Config",
			inputFile: "empty.json",
			inputContent: `{
				"mcpServers": {}
			}`,
			expectedConfig: &McpAnyConfig{
				UpstreamServices: []UpstreamService{},
			},
		},
		{
			name:      "Multiple Servers",
			inputFile: "multi.json",
			inputContent: `{
				"mcpServers": {
					"s1": {"command": "c1"},
					"s2": {"command": "c2"}
				}
			}`,
			expectedConfig: &McpAnyConfig{
				UpstreamServices: []UpstreamService{
					{
						Name: "s1",
						McpService: &McpService{
							StdioConnection: &StdioConnection{Command: "c1", Args: []string{}},
						},
					},
					{
						Name: "s2",
						McpService: &McpService{
							StdioConnection: &StdioConnection{Command: "c2", Args: []string{}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(tempDir, tc.inputFile)
			if tc.inputContent != "" {
				err := os.WriteFile(inputPath, []byte(tc.inputContent), 0644)
				require.NoError(t, err)
			}

			// Prepare args
			cmdArgs := append([]string{inputPath}, tc.args...)

			// Adjust output path to be within tempDir
			if tc.checkFile {
				for i, arg := range cmdArgs {
					if arg == "--output" || arg == "-o" {
						if i+1 < len(cmdArgs) {
							cmdArgs[i+1] = filepath.Join(tempDir, cmdArgs[i+1])
						}
					}
				}
			}

			cmd := newImportCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(cmdArgs)

			err := cmd.Execute()

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)

				var outputContent []byte
				if tc.checkFile {
					outputPath := ""
					for i, arg := range cmdArgs {
						if arg == "--output" || arg == "-o" {
							outputPath = cmdArgs[i+1]
							break
						}
					}
					content, err := os.ReadFile(outputPath)
					require.NoError(t, err)
					outputContent = content

					// Also verify success message on stdout
					assert.Contains(t, buf.String(), "Successfully imported configuration")
				} else {
					outputContent = buf.Bytes()
				}

				var actualConfig McpAnyConfig
				err = yaml.Unmarshal(outputContent, &actualConfig)
				require.NoError(t, err)

				assert.Equal(t, len(tc.expectedConfig.UpstreamServices), len(actualConfig.UpstreamServices))

				findService := func(name string) *UpstreamService {
					for _, s := range actualConfig.UpstreamServices {
						if s.Name == name {
							return &s
						}
					}
					return nil
				}

				for _, expectedSvc := range tc.expectedConfig.UpstreamServices {
					actualSvc := findService(expectedSvc.Name)
					require.NotNil(t, actualSvc, "Service %s not found", expectedSvc.Name)

					// Deep compare fields
					assert.Equal(t, expectedSvc.Name, actualSvc.Name)
					if expectedSvc.McpService != nil {
						require.NotNil(t, actualSvc.McpService)
						if expectedSvc.McpService.StdioConnection != nil {
							require.NotNil(t, actualSvc.McpService.StdioConnection)
							assert.Equal(t, expectedSvc.McpService.StdioConnection.Command, actualSvc.McpService.StdioConnection.Command)
							assert.Equal(t, expectedSvc.McpService.StdioConnection.Args, actualSvc.McpService.StdioConnection.Args)
							assert.Equal(t, expectedSvc.McpService.StdioConnection.Env, actualSvc.McpService.StdioConnection.Env)
						}
					}
				}
			}
		})
	}
}
