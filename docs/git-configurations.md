# Loading Configurations from Git Repositories

MCP Any supports loading configurations from Git repositories, which allows you to easily share and reuse tool definitions. This feature is particularly useful for teams that want to maintain a central repository of tool configurations and share them across multiple projects.

## How it Works

When you specify a Git repository as a source for your configuration, MCP Any will clone the repository to a temporary directory and then load the specified configuration file from the cloned repository. The server will then start as usual with the loaded configuration.

## Usage

To load a configuration from a Git repository, use the `--config-git-url` and `--config-git-path` flags.

- `--config-git-url`: The URL of the Git repository to clone.
- `--config-git-path`: The path to the configuration file within the repository.

### Example

Here is an example of how to load a configuration from a public Git repository:

```bash
make run ARGS="--config-git-url https://github.com/mcpany/examples.git --config-git-path basic-http-service/config.yaml"
```

This command will:

1.  Clone the `mcpany/examples` repository to a temporary directory.
2.  Load the `config.yaml` file from the `basic-http-service` directory in the cloned repository.
3.  Start the MCP Any server with the loaded configuration.

## Private Repositories

Currently, MCP Any only supports cloning public Git repositories. Support for private repositories may be added in a future release.

## Security Considerations

When loading configurations from Git repositories, it is important to be aware of the security implications. Cloning and loading configurations from untrusted repositories can expose your system to security vulnerabilities.

**Only load configurations from trusted sources.**

Before loading a configuration from a Git repository, make sure that you trust the source of the repository and have reviewed the contents of the configuration file.
