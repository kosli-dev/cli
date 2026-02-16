# Kosli CLI

The Kosli CLI is a command-line tool for recording and querying software delivery events. This npm package provides an easy way to install and use the Kosli CLI in JavaScript/TypeScript projects.

## Installation

### Global Installation

Install globally to use the CLI anywhere:

```bash
npm install -g @kosli/cli
```

### Project Installation

Install as a development dependency in your project:

```bash
npm install --save-dev @kosli/cli
```

## Usage

After installation, you can use the `kosli` command:

```bash
# Check version
kosli version

# Get help
kosli --help

# Example: Report a deployment
kosli report deployment <deployment-name> \
  --environment <environment-name> \
  --api-token <your-api-token>
```

### Using with npx

You can run Kosli without installing it:

```bash
npx @kosli/cli version
```

### Using in npm scripts

Add Kosli commands to your `package.json`:

```json
{
  "scripts": {
    "kosli:version": "kosli version",
    "kosli:report": "kosli report deployment production --environment prod"
  }
}
```

Then run:

```bash
npm run kosli:version
```

## How It Works

This npm package downloads the appropriate platform-specific Kosli CLI binary during installation:

- **Supported Platforms:** macOS (darwin), Linux, Windows
- **Supported Architectures:** x64 (amd64), arm64

The binary is downloaded from the [official Kosli CLI releases](https://github.com/kosli-dev/cli/releases).

## Requirements

- **Node.js:** >= 14.0.0
- **Internet connection:** Required during installation to download the binary

## Documentation

For full documentation, visit:

- [Kosli Documentation](https://docs.kosli.com/)
- [Kosli CLI GitHub Repository](https://github.com/kosli-dev/cli)

## Troubleshooting

### Installation Issues

If the binary download fails during installation, you can:

1. Manually download the binary from [GitHub releases](https://github.com/kosli-dev/cli/releases)
2. Place it in `node_modules/@kosli/cli/bin/`
3. Make it executable: `chmod +x node_modules/@kosli/cli/bin/kosli`

### Version Mismatch

To verify the installed version:

```bash
kosli version
```

The version should match the npm package version.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Install Kosli CLI
  run: npm install -g @kosli/cli

- name: Report deployment
  run: |
    kosli report deployment ${{ github.sha }} \
      --environment production \
      --api-token ${{ secrets.KOSLI_API_TOKEN }}
```

### GitLab CI

```yaml
deploy:
  before_script:
    - npm install -g @kosli/cli
  script:
    - kosli report deployment $CI_COMMIT_SHA
      --environment production
      --api-token $KOSLI_API_TOKEN
```

## Support

- [GitHub Issues](https://github.com/kosli-dev/cli/issues)
- [Kosli Community](https://www.kosli.com/community)
- [Documentation](https://docs.kosli.com/)

## License

MIT License - See the [LICENSE](https://github.com/kosli-dev/cli/blob/main/LICENSE) file for details.

## About Kosli

Kosli provides complete visibility and change tracking for your software delivery pipelines. Learn more at [kosli.com](https://www.kosli.com).
