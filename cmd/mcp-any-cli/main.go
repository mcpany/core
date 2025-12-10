package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// SurveyAsker is an interface for asking survey questions.
// This is used to allow for mocking in tests.
type SurveyAsker interface {
	Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error
}

// DefaultSurveyAsker is the default implementation of SurveyAsker that uses the survey library.
type DefaultSurveyAsker struct{}

// Ask asks the questions.
func (d *DefaultSurveyAsker) Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	return survey.Ask(qs, response, opts...)
}

var asker SurveyAsker = &DefaultSurveyAsker{}

type UpstreamService struct {
	Name        string      `yaml:"name"`
	HttpService HttpService `yaml:"httpService,omitempty"`
	GrpcService GrpcService `yaml:"grpcService,omitempty"`
}

type HttpService struct {
	Address string `yaml:"address"`
}

type GrpcService struct {
	Address string `yaml:"address"`
}

type Config struct {
	UpstreamServices []UpstreamService `yaml:"upstreamServices"`
}

// the questions to ask
var qs = []*survey.Question{
	{
		Name: "serviceType",
		Prompt: &survey.Select{
			Message: "Choose a service type:",
			Options: []string{"gRPC", "OpenAPI", "HTTP"},
			Default: "HTTP",
		},
	},
	{
		Name:     "serviceName",
		Prompt:   &survey.Input{Message: "What is the service name?"},
		Validate: survey.Required,
	},
	{
		Name:     "serviceAddress",
		Prompt:   &survey.Input{Message: "What is the service address?"},
		Validate: survey.Required,
	},
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp-any-cli",
		Short: "A CLI tool to generate MCP Any configuration files.",
		Long:  `A CLI tool to generate MCP Any configuration files interactively.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			answers := struct {
				ServiceType    string `survey:"serviceType"`
				ServiceName    string `survey:"serviceName"`
				ServiceAddress string `survey:"serviceAddress"`
			}{}

			// perform the survey
			err := asker.Ask(qs, &answers)
			if err != nil {
				return err
			}

			var config Config
			service := UpstreamService{
				Name: answers.ServiceName,
			}

			switch answers.ServiceType {
			case "HTTP", "OpenAPI":
				service.HttpService = HttpService{Address: answers.ServiceAddress}
			case "gRPC":
				service.GrpcService = GrpcService{Address: answers.ServiceAddress}
			}

			config.UpstreamServices = append(config.UpstreamServices, service)

			yamlData, err := yaml.Marshal(&config)
			if err != nil {
				return err
			}

			fmt.Println("---")
			fmt.Println(string(yamlData))

			return nil
		},
	}
	return cmd
}

var rootCmd = newRootCmd()

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of mcp-any-cli",
	Long:  `All software has versions. This is mcp-any-cli's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mcp-any-cli v0.1")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func execute() error {
	return rootCmd.Execute()
}

func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
