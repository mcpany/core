// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/spf13/cobra"
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively create a new configuration file",
	Long: `Guided wizard to help you create a valid configuration file for MCP Any.
It will ask you a series of questions about the service you want to add and generate a YAML configuration file.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		reader := bufio.NewReader(cmd.InOrStdin())
		out := cmd.OutOrStdout()

		_, _ = fmt.Fprintln(out, "🚀 Welcome to MCP Any Configuration Wizard!")
		_, _ = fmt.Fprintln(out, "This wizard will help you create a configuration file for your first service.")
		_, _ = fmt.Fprintln(out, "")

		// 1. Service Name
		serviceName := prompt(reader, out, "What should we name this service? (e.g., my-weather-api)", "my-service")

		// 2. Service Type
		_, _ = fmt.Fprintln(out, "\nWhich type of upstream service do you want to connect?")
		_, _ = fmt.Fprintln(out, "1. HTTP (REST API)")
		_, _ = fmt.Fprintln(out, "2. Command (stdio/CLI tool)")
		_, _ = fmt.Fprintln(out, "3. SQL Database (Postgres, SQLite, MySQL)")
		_, _ = fmt.Fprintln(out, "4. gRPC")

		typeChoice := prompt(reader, out, "Enter the number:", "1")

		var configContent string

		switch typeChoice {
		case "1": // HTTP
			address := prompt(reader, out, "Enter the base URL of the API (e.g., https://api.weather.gov):", "https://api.example.com")
			configContent = generateHTTPConfig(serviceName, address)
		case "2": // Command
			command := prompt(reader, out, "Enter the command to run (e.g., python):", "python")
			argsInput := prompt(reader, out, "Enter arguments separated by spaces (e.g., script.py --flag):", "script.py")
			argsList := strings.Fields(argsInput)
			configContent = generateCommandConfig(serviceName, command, argsList)
		case "3": // SQL
			driver := prompt(reader, out, "Enter database driver (postgres, mysql, sqlite3):", "postgres")
			dsn := prompt(reader, out, "Enter connection string (DSN):", "user:pass@tcp(localhost:5432)/dbname")
			configContent = generateSQLConfig(serviceName, driver, dsn)
		case "4": // gRPC
			address := prompt(reader, out, "Enter the gRPC server address (e.g., localhost:50051):", "localhost:50051")
			configContent = generateGRPCConfig(serviceName, address)
		default:
			_, _ = fmt.Fprintln(out, "Invalid choice, defaulting to HTTP.")
			address := prompt(reader, out, "Enter the base URL of the API:", "https://api.example.com")
			configContent = generateHTTPConfig(serviceName, address)
		}

		// 3. Output File
		filename := prompt(reader, out, "\nWhere should we save this config?", "config.yaml")

		// Check if file exists
		if _, err := os.Stat(filename); err == nil {
			overwrite := prompt(reader, out, fmt.Sprintf("File '%s' already exists. Overwrite? (y/N)", filename), "n")
			if strings.ToLower(overwrite) != "y" {
				_, _ = fmt.Fprintln(out, "Aborted.")
				return nil
			}
		}

		// Change permission to 0600 to satisfy gosec G306, though 0644 is usually fine for config unless it has secrets
		// But let's be safe.
		err := os.WriteFile(filename, []byte(configContent), 0600)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		_, _ = fmt.Fprintf(out, "\n✅ Configuration saved to %s\n", filename)
		_, _ = fmt.Fprintln(out, "\nYou can now run the server with:")
		_, _ = fmt.Fprintf(out, "  %s run --config-path %s\n", appconsts.Name, filename)

		return nil
	},
}

func prompt(reader *bufio.Reader, out interface{ Write([]byte) (int, error) }, question, defaultVal string) string {
	_, _ = fmt.Fprintf(out, "%s [%s]: ", question, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func generateHTTPConfig(name, address string) string {
	return fmt.Sprintf(`upstream_services:
  - name: "%s"
    http_service:
      address: "%s"
      # Add tools manually or via OpenAPI spec (see docs)
`, name, address)
}

func generateCommandConfig(name, command string, args []string) string {
	argsStr := ""
	for _, arg := range args {
		argsStr += fmt.Sprintf("\n        - \"%s\"", arg)
	}
	return fmt.Sprintf(`upstream_services:
  - name: "%s"
    command_service:
      command: "%s"
      args:%s
`, name, command, argsStr)
}

func generateSQLConfig(name, driver, dsn string) string {
	return fmt.Sprintf(`upstream_services:
  - name: "%s"
    sql_service:
      driver: "%s"
      dsn: "%s"
      # Define safe queries as tools below
      calls:
        get_users:
          query: "SELECT * FROM users LIMIT 10"
`, name, driver, dsn)
}

func generateGRPCConfig(name, address string) string {
	return fmt.Sprintf(`upstream_services:
  - name: "%s"
    grpc_service:
      address: "%s"
      # Reflection is enabled by default to discover tools
`, name, address)
}
