# NPM Packaging

This directory contains the npm package structure for distributing the Kosli CLI via npm, following the same pattern used by [esbuild](https://github.com/evanw/esbuild).

## Structure

```
npm/
├── wrapper/              # @kosli-dev/cli — the package users install
│   ├── package.json      # declares optionalDependencies for all platforms
│   ├── bin/kosli         # JS shim that detects the platform and runs the binary
│   └── install.js        # postinstall script that validates the binary
├── cli-linux-x64/        # @kosli-dev/cli-linux-x64
├── cli-linux-arm64/      # @kosli-dev/cli-linux-arm64
├── cli-darwin-x64/       # @kosli-dev/cli-darwin-x64
└── cli-darwin-arm64/     # @kosli-dev/cli-darwin-arm64
    ├── package.json      # declares os/cpu fields for platform filtering
    └── bin/kosli         # the native binary — see below
```

## How it works

Users install a single package:

```sh
npm install @kosli-dev/cli
```

npm resolves the `optionalDependencies` declared in the wrapper's `package.json` and installs only the platform-specific package that matches the current OS and CPU architecture — all non-matching packages are silently skipped. The wrapper's `bin/kosli` JS shim then locates the binary inside the installed platform package and executes it.

## The `bin/` directories are populated by goreleaser

The platform package `bin/` directories are **not committed to git**. They are populated automatically during the release process by a post-build hook in [`.goreleaser.yml`](../.goreleaser.yml):

```yaml
hooks:
  post:
    - cmd: bash -c 'ARCH="{{ .Arch }}"; [ "$ARCH" = "amd64" ] && ARCH="x64"; mkdir -p npm/cli-{{ .Os }}-${ARCH}/bin && cp "{{ .Path }}" npm/cli-{{ .Os }}-${ARCH}/bin/kosli && chmod +x npm/cli-{{ .Os }}-${ARCH}/bin/kosli'
```

This hook runs once per build target immediately after goreleaser compiles the binary. It maps goreleaser's architecture naming (`amd64` → `x64`, `arm64` → `arm64`) to the npm naming convention and copies the binary into the correct platform package directory.

## Publishing

Platform packages must be published before the wrapper, since the wrapper's `optionalDependencies` references them. After a goreleaser build has populated the `bin/` directories:

```sh
# Publish platform packages first
(cd npm/cli-linux-x64   && npm publish)
(cd npm/cli-linux-arm64 && npm publish)
(cd npm/cli-darwin-x64  && npm publish)
(cd npm/cli-darwin-arm64 && npm publish)

# Then publish the wrapper
(cd npm/wrapper && npm publish)
```

Or as a one-liner that publishes platform packages first, then the wrapper:

```sh
find npm -name package.json ! -path "npm/wrapper/*" | while read f; do pushd "$(dirname "$f")" && npm publish; popd; done && pushd npm/wrapper && npm publish; popd
```

## Versioning

All packages must share the same version number. When bumping the version, update it in all five `package.json` files — the four platform packages and the wrapper — as well as the `optionalDependencies` versions in `npm/wrapper/package.json`.
