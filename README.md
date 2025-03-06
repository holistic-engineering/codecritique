# CodeCritique

CodeCritique is a powerful tool for automated code review of pull requests, leveraging AI to provide insightful feedback on your code changes.

## Features

- Automated code review for GitHub and GitLab pull requests
- AI-powered analysis using various LLM providers (Groq, Ollama, etc.)
- Multiple output formats (Markdown, JSON, HTML)
- Easy to configure and extend
- Containerized deployment support

## Installation

### Prerequisites

- Go 1.22 or higher
- Git
- Docker (optional, for containerized usage)

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/holistic-engineering/codecritique.git
   cd codecritique
   ```

2. Build the application:
   ```bash
   make build
   ```

### Using Docker

1. Build the Docker image:
   ```bash
   make docker-build
   ```

## Configuration

CodeCritique uses a TOML configuration file located at `settings/settings.toml`. You can customize the following settings:

### Git Provider

```toml
[git]
provider = "GitHub" # Options: GitHub, GitLab
token = "" # Set this via environment variable
```

### AI Provider

```toml
[ai]
provider = "Groq" # Options: Anthropic, Groq, Ollama, OpenAI
ollama_url = "http://localhost:11434/api/generate"
ollama_model = "llama3.1"
groq_api_key = "" # Set this via environment variable
groq_model = "mixtral-8x7b-32768"
```

### Output Format

```toml
[printer]
kind = "markdown" # Options: json, html, markdown
```

## Usage

### Basic Usage

```bash
./codecritique <owner/repo> <pr_number>
```

Example:
```bash
./codecritique holistic-engineering/codecritique 42
```

### Using Docker

```bash
docker run -v $(pwd)/settings:/root/settings codecritique:latest <owner/repo> <pr_number>
```

## Development

### Running Tests

```bash
make test
```

### Linting

```bash
make lint
```

### Security Check

```bash
make security-check
```

### Code Coverage

```bash
make coverage
```

## Project Structure

- `cmd/cli`: Command-line interface entry point
- `config`: Configuration handling
- `internal/critique`: Core code review logic
- `internal/infra`: Infrastructure components
  - `ai`: AI provider integrations
  - `git`: Git provider integrations
  - `printer`: Output formatting

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).
