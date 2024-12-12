<p align="center">
  <img src="./unit.png" alt="CLI Unit Test Generator Logo" />
</p>

# CLI Unit Test Generator

`Unit` is a CLI tool designed to analyze source code and generate unit tests efficiently using AI.

Currently, the tool accepts a single source code file as input. It uses AI to analyze the source code, list possible unit tests, and generate the desired unit test code based on user input.

## Features

- **AI-powered code analysis**: Analyze source code to identify unit test cases.
- **Multi-provider support**: Works with OpenAI and Anthropic AI models.
- **Caching**: Reduces redundant API calls by caching analysis results.
- **Customizable configuration**: Supports CLI flags and YAML-based configuration.

## Prerequisites

Before using Unit, ensure you have the following:

1. **AI API Key**: A valid API key from OpenAI or Anthropic.
2. **Go Installed**: The tool is built using Go, and `make` is used for building the application.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/cyber-nic/unit
   ```

2. Build the tool:
   ```bash
   make
   ```

## AI Providers and Models

Unit supports the following AI providers and models:

- **Anthropic** (default): `claude-3-5-haiku-latest`
- **OpenAI**: `gpt-4o-mini`

## Configuration

Unit allows configuration through CLI flags or YAML files.

### CLI Flags

The following configuration values can be set via CLI flags:

- `--color`: Enable/disable color output
- `--debug`: Enable/disable debug mode
- `--provider`: Specify the AI provider: `anthropic` / `openai`
- `--write`: Automatically write generated test files

Example:

```bash
./unit ./main.go --color=false --debug=true --write=true
```

### YAML Configuration

You can also configure Unit using a `.unit.yaml` file, either in the current directory or in `$HOME`.

Example:

```yaml
# ~/.unit.yaml
debug: false # default
color: true # default
write: false # default
provider: anthropic # default
secret_path: $HOME/.secrets/anthropic_api_key
```

To set up a default configuration file:

```bash
cp .unit-example.yaml ~/.unit.yaml
```

### Default Values

If no configuration is provided, Unit uses the following defaults:

- debug: `false`
- color: `true`
- write: `false`
- provider: `anthropic`
- secret_path: `.provider_api_key`

## API Key Management

Unit requires an AI provider API key, which can be provided in two ways:

1. **Environment Variable**:

   ```bash
   export AI_API_KEY=sk-...
   ```

2. **Secret File** (recommended):
   Save your API key to a secure file:
   ```bash
   echo "sk-..." > $HOME/.secrets/anthropic_api_key
   ```
   Ensure the `secret_path` in the configuration points to this file.

## Write

At the moment enabling the `write` setting simply outputs the generated unit test to `unit_test.go`.

## Caching

Unit minimizes API calls by caching the results of source code analysis. Cache files are stored under:

```bash
/tmp/unit/<source-file-sha256>
```

This improves performance and reduces the cost of repeated analyses.

## Usage

### Basic Usage

To generate a unit test, provide the path to a source file:

```bash
unit ./main.go
```

### Using a Specific Provider

To specify the AI provider:

- Anthropic (default):
  ```bash
  unit ./client-anthropic.go
  ```
- OpenAI:
  ```bash
  unit ./main.go --provider=openai
  ```

## Examples

### Generate Unit Tests

1. Use Anthropic to generate a unit test for `client-anthropic.go`:

   ```bash
   unit ./client-anthropic.go
   ```

2. Use OpenAI to generate a unit test for `main.go`:
   ```bash
   unit ./main.go --provider=openai
   ```

## License

This project is licensed under [The MIT License](https://opensource.org/licenses/mit/). For more details, see the [`LICENSE`](LICENSE) file.
