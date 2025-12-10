package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
)

// MockSurveyAsker is a mock implementation of the SurveyAsker interface.
type MockSurveyAsker struct {
	Answers struct {
		ServiceType    string
		ServiceName    string
		ServiceAddress string
	}
	Error error
}

// Ask is the mock implementation of the Ask method.
func (m *MockSurveyAsker) Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	if m.Error != nil {
		return m.Error
	}
	res := response.(*struct {
		ServiceType    string `survey:"serviceType"`
		ServiceName    string `survey:"serviceName"`
		ServiceAddress string `survey:"serviceAddress"`
	})
	res.ServiceType = m.Answers.ServiceType
	res.ServiceName = m.Answers.ServiceName
	res.ServiceAddress = m.Answers.ServiceAddress
	return nil
}

func TestRootCmd(t *testing.T) {
	testCases := []struct {
		name         string
		serviceType  string
		serviceName  string
		serviceAddr  string
		expectedYaml string
		err          error
	}{
		{
			name:        "HTTP Service",
			serviceType: "HTTP",
			serviceName: "my-http-service",
			serviceAddr: "http://localhost:8080",
			expectedYaml: `
upstreamServices:
- name: my-http-service
  httpService:
    address: http://localhost:8080
`,
		},
		{
			name:        "gRPC Service",
			serviceType: "gRPC",
			serviceName: "my-grpc-service",
			serviceAddr: "localhost:50051",
			expectedYaml: `
upstreamServices:
- name: my-grpc-service
  grpcService:
    address: localhost:50051
`,
		},
		{
			name:        "OpenAPI Service",
			serviceType: "OpenAPI",
			serviceName: "my-openapi-service",
			serviceAddr: "http://localhost:8081",
			expectedYaml: `
upstreamServices:
- name: my-openapi-service
  httpService:
    address: http://localhost:8081
`,
		},
		{
			name: "User cancel",
			err:  errors.New("user canceled"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock asker
			asker = &MockSurveyAsker{
				Answers: struct {
					ServiceType    string
					ServiceName    string
					ServiceAddress string
				}{
					ServiceType:    tc.serviceType,
					ServiceName:    tc.serviceName,
					ServiceAddress: tc.serviceAddr,
				},
				Error: tc.err,
			}

			// Keep a reference to the original os.Stdout
			old := os.Stdout
			// Create a new pipe
			r, w, _ := os.Pipe()
			// Set os.Stdout to the write end of the pipe
			os.Stdout = w

			// Create a command to test
			cmd := newRootCmd()
			cmd.SetArgs([]string{})

			// Execute the command
			err := cmd.Execute()
			if tc.err != nil {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				if err.Error() != tc.err.Error() {
					t.Errorf("Expected error '%s', but got '%s'", tc.err.Error(), err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Close the write end of the pipe
			w.Close()
			// Restore os.Stdout
			os.Stdout = old

			// Read the output from the read end of the pipe
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check the output
			if tc.err == nil {
				if !strings.Contains(buf.String(), strings.TrimSpace(tc.expectedYaml)) {
					t.Errorf("Expected output to contain '%s', but got '%s'", tc.expectedYaml, buf.String())
				}
			}
		})
	}
}

func TestVersionCmd(t *testing.T) {
	// Keep a reference to the original os.Stdout
	old := os.Stdout
	// Create a new pipe
	r, w, _ := os.Pipe()
	// Set os.Stdout to the write end of the pipe
	os.Stdout = w

	// Create a command to test
	cmd := newRootCmd()
	cmd.AddCommand(versionCmd)
	cmd.SetArgs([]string{"version"})

	// Execute the command
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Close the write end of the pipe
	w.Close()
	// Restore os.Stdout
	os.Stdout = old

	// Read the output from the read end of the pipe
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output
	expected := "mcp-any-cli v0.1"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("Expected output to contain '%s', but got '%s'", expected, buf.String())
	}
}
