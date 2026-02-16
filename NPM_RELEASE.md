# NPM Release Guide

This document describes how to release the Kosli CLI as an npm package.

## Overview

The Kosli CLI is primarily a Go binary, but we also publish it as an npm package for easy installation in JavaScript/TypeScript projects. The npm package downloads the appropriate platform-specific binary during installation.

## Package Structure

All npm packaging files are located in the `npm-package/` directory:

- `npm-package/package.json` - npm package metadata
- `npm-package/index.js` - Entry point that spawns the Kosli binary
- `npm-package/install.js` - Post-install script that downloads the binary
- `npm-package/test.js` - Simple test to verify installation
- `npm-package/.npmignore` - Files to exclude from the npm package
- `npm-package/bin/` - Directory where the binary is placed (created during build/test)

## Release Process

### 1. Test Locally

The test script automatically extracts the version from the kosli binary and updates `package.json`. Before publishing, test the package locally:

```bash
# Run the automated test script
./test-npm-package.sh
```

This script will:
- Build the kosli binary (if not already built)
- Copy the binary to `npm-package/bin/`
- Extract the version from the binary and update `package.json`
- Pack the npm package
- Install it in a temporary directory
- Run tests to verify installation
- Clean up

### 2. Publish to npm

```bash
# Login to npm (first time only)
npm login

# Navigate to npm-package directory
cd npm-package

# Publish the package (npm will automatically pack it)
npm publish --access public
```

**Note:** Make sure you have permission to publish to the `@kosli` organization on npm.

### 3. Automate with GitHub Actions (Recommended)

You can automate npm publishing by adding a step to your existing release workflow:

```yaml
- name: Publish to npm
  if: startsWith(github.ref, 'refs/tags/v')
  run: |
    # Extract version from tag (removes 'v' prefix)
    VERSION=${GITHUB_REF#refs/tags/v}

    # Copy binary to npm-package structure
    mkdir -p npm-package/bin
    cp ./kosli npm-package/bin/
    chmod +x npm-package/bin/kosli

    # Update package.json version
    cd npm-package
    npm version $VERSION --no-git-tag-version --allow-same-version

    # Publish to npm
    echo "//registry.npmjs.org/:_authToken=${{ secrets.NPM_TOKEN }}" > .npmrc
    npm publish --access public
  env:
    NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
```

You'll need to add an `NPM_TOKEN` secret to your GitHub repository settings.

## Installation for Users

Once published, users can install the CLI via npm:

```bash
# Install globally
npm install -g @kosli/cli

# Or use in a project
npm install --save-dev @kosli/cli

# Run via npx
npx @kosli/cli version
```

## How It Works

1. When `npm install` runs, the `postinstall` script (`install.js`) executes
2. The script checks if the binary already exists (for local testing)
3. If not present, it detects the platform (OS and architecture)
4. It downloads the appropriate binary from GitHub releases
5. The binary is extracted to the `bin/` directory
6. The `index.js` wrapper script forwards all commands to the binary

**Note:** For local testing, when the binary is already packaged in `npm-package/bin/`, the download step is skipped.

## Supported Platforms

The npm package supports the same platforms as the Go binary:

- **OS:** macOS (darwin), Linux, Windows
- **Architecture:** x64 (amd64), arm64, arm

## Troubleshooting

### Installation fails to download binary

If the binary download fails, users can manually download it from:
```
https://github.com/kosli-dev/cli/releases
```

### Binary not found after installation

Check that the binary exists in the package:
```bash
ls -la node_modules/@kosli/cli/bin/
```

### Test the installation

Run the included test:
```bash
npm test
```

## Development with devcontainer

The devcontainer includes Node.js and npm, so you can test the npm package workflow:

```bash
# Inside the devcontainer
./test-npm-package.sh
```

## Manual Package Creation

If you want to create the package without running tests:

```bash
# Build the binary
make build

# Copy to npm-package structure
mkdir -p npm-package/bin
cp ./kosli npm-package/bin/
chmod +x npm-package/bin/kosli

# Extract version and update package.json
KOSLI_VERSION=$(./kosli version 2>/dev/null | head -1 | grep -oP 'Version:"v\K[^"]+')
cd npm-package
npm version "$KOSLI_VERSION" --no-git-tag-version --allow-same-version

# Pack the package
npm pack
```

This creates `kosli-cli-<version>.tgz` in the `npm-package/` directory.

## Notes

- The npm package version is automatically extracted from the kosli binary
- For published packages, the release must exist on GitHub before users can install it
- Users need internet access during installation to download the binary (unless the binary is pre-packaged)
- Consider caching the binary in CI/CD pipelines to avoid repeated downloads
- All npm packaging is isolated in the `npm-package/` directory to avoid conflicts with other project files
