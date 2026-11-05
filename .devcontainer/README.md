# Kosli CLI DevContainer

This devcontainer provides a complete development environment for the Kosli CLI project.

## What's Included

### Languages & Runtimes

- **Go 1.25** - Primary development language
- **Node.js LTS** - For npm package management and release

### Development Tools

- **golangci-lint** - Go linting and static analysis
- **goreleaser** - Release automation
- **gotestsum** - Enhanced Go test runner
- **Hugo Extended** - Documentation site generation
- **GitHub CLI (gh)** - GitHub integration

### Infrastructure

- **Docker-in-Docker** - For integration tests (runs Kosli server)
- **Make** - Build automation
- **Git** - Version control

## Getting Started

1. **Open in VS Code**
   - Install the "Dev Containers" extension
   - Open this repository in VS Code
   - Click "Reopen in Container" when prompted (or use Command Palette: "Dev Containers: Reopen in Container")

2. **Wait for Setup**
   - The container will build (first time takes a few minutes)
   - Post-create commands will run automatically (`go mod download`, etc.)

3. **Start Developing**

   ```bash
   # Build the CLI
   make build

   # Run tests
   make test_integration

   # Run linting
   make lint

   # Build documentation
   make hugo-local
   ```

## Container Features

### Forwarded Ports

- **1515** - Hugo documentation server
- **8001** - Local Kosli server (for integration tests)

### Volume Mounts

- Go module cache is persisted in a Docker volume for faster builds
- Docker socket is mounted for Docker-in-Docker functionality

### VS Code Extensions

The following extensions are automatically installed:

- **Go** - Go language support
- **Docker** - Docker file and container management
- **GitLens** - Enhanced Git capabilities

### VS Code Settings

Pre-configured with:

- Go formatting on save
- Automatic import organization
- golangci-lint integration
- Go language server (gopls)

## Common Tasks

### Building the CLI

```bash
make build
./kosli version
```

### Running Tests

```bash
# Integration tests (requires Docker)
make test_integration

# Run all tests including slow ones
make test_integration_full

# Run specific test suite
make test_integration_single TARGET=TestAttestJunitCommandTestSuite
```

### Working with NPM Package

```bash
# Test npm package installation locally
npm install
npm test

# Pack for distribution
npm pack

# Test the packed version
mkdir /tmp/npm-test && cd /tmp/npm-test
npm install /workspace/*.tgz
npx kosli version
```

### Documentation

```bash
# Generate CLI documentation
make cli-docs

# Start Hugo documentation server (accessible at localhost:1515)
make hugo-local
```

### Linting

```bash
make lint
```

## Environment Variables

The following environment variables are pre-configured:

- `CGO_ENABLED=0` - Disable CGO for static binaries
- `GO111MODULE=on` - Use Go modules
- `GOPATH=/go` - Go workspace location

### Required for Testing

You need to set `KOSLI_API_TOKEN_PROD` for integration tests:

```bash
export KOSLI_API_TOKEN_PROD="your-token-here"
```

## Docker Integration

The devcontainer uses Docker-in-Docker to run integration tests. The Kosli server runs in a container alongside your tests.

### Managing Test Containers

```bash
# View server logs
make logs_integration_test_server

# Follow server logs
make follow_integration_test_server

# Enter server container
make enter_integration_test_server
```

## Troubleshooting

### Container won't start

- Ensure Docker Desktop is running
- Check Docker has enough resources (4GB+ RAM recommended)
- Try rebuilding: Command Palette → "Dev Containers: Rebuild Container"

### Tests failing

- Ensure `KOSLI_API_TOKEN_PROD` is set
- Check Docker is working: `docker ps`
- Verify network: `docker network ls` (should see `cli_net`)

### Go modules issues

```bash
go clean -modcache
go mod download
go mod tidy
```

### NPM installation issues

```bash
rm -rf node_modules package-lock.json
npm install
```

## Customization

### Adding VS Code Extensions

Edit [.devcontainer/devcontainer.json](.devcontainer/devcontainer.json):

```json
"extensions": [
  "golang.go",
  "your-extension-id"
]
```

### Installing Additional Tools

Edit [.devcontainer/Dockerfile](.devcontainer/Dockerfile) and rebuild the container.

### Changing Environment Variables

Edit the `remoteEnv` section in [.devcontainer/devcontainer.json](.devcontainer/devcontainer.json).

## Non-Root User

The container runs as the `vscode` user (non-root) for security. Use `sudo` if you need root access:

```bash
sudo apt-get install some-package
```

## References

- [Dev Containers Documentation](https://code.visualstudio.com/docs/devcontainers/containers)
- [Kosli Documentation](https://docs.kosli.com/)
- [Project Development Guide](../dev-guide.md)
